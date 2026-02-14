package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/zenobi-us/opennotes/internal/search"
)

// RetrievalMode controls which retrieval strategy is used.
type RetrievalMode string

const (
	RetrievalModeHybrid   RetrievalMode = "hybrid"
	RetrievalModeKeyword  RetrievalMode = "keyword"
	RetrievalModeSemantic RetrievalMode = "semantic"
)

// SemanticSearchMeta provides execution metadata for semantic command UX.
type SemanticSearchMeta struct {
	Mode             RetrievalMode
	UsedKeyword      bool
	UsedSemantic     bool
	SemanticFallback bool
}

// ParseRetrievalMode validates and normalizes retrieval mode strings.
func ParseRetrievalMode(raw string) (RetrievalMode, error) {
	mode := RetrievalMode(strings.ToLower(strings.TrimSpace(raw)))
	switch mode {
	case RetrievalModeHybrid, RetrievalModeKeyword, RetrievalModeSemantic:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid mode %q (allowed: hybrid, keyword, semantic)", raw)
	}
}

// SearchSemantic executes semantic/keyword/hybrid retrieval using one condition parser path.
func (s *NoteService) SearchSemantic(
	ctx context.Context,
	query string,
	conditions []QueryCondition,
	mode RetrievalMode,
	topK int,
) ([]Note, SemanticSearchMeta, error) {
	if s.notebookPath == "" {
		return nil, SemanticSearchMeta{}, fmt.Errorf("no notebook selected")
	}

	if topK <= 0 {
		topK = 100
	}

	meta := SemanticSearchMeta{Mode: mode}

	keywordCandidates, err := s.findKeywordCandidates(ctx, query, conditions, topK)
	if err != nil {
		return nil, meta, err
	}

	switch mode {
	case RetrievalModeKeyword:
		meta.UsedKeyword = true
		return searchResultsToNotes(keywordCandidates), meta, nil

	case RetrievalModeSemantic:
		semanticCandidates, semErr := s.findSemanticCandidates(ctx, query, conditions, topK)
		if semErr != nil {
			return nil, meta, semErr
		}
		meta.UsedSemantic = true
		return semanticResultsToNotes(semanticCandidates), meta, nil

	case RetrievalModeHybrid:
		meta.UsedKeyword = true

		semanticCandidates, semErr := s.findSemanticCandidates(ctx, query, conditions, topK)
		if semErr != nil {
			if semErr == ErrSemanticUnavailable {
				meta.SemanticFallback = true
				return searchResultsToNotes(keywordCandidates), meta, nil
			}
			return nil, meta, semErr
		}

		meta.UsedSemantic = true
		merged := MergeHybridResults(keywordCandidates, semanticCandidates, 60)
		if len(merged) > topK {
			merged = merged[:topK]
		}
		return hybridResultsToNotes(merged), meta, nil

	default:
		return nil, meta, fmt.Errorf("unsupported retrieval mode: %s", mode)
	}
}

func (s *NoteService) findKeywordCandidates(
	ctx context.Context,
	query string,
	conditions []QueryCondition,
	topK int,
) ([]search.Result, error) {
	if s.index == nil {
		return nil, fmt.Errorf("index not initialized")
	}

	queryAST, err := s.searchService.BuildQuery(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	if query != "" {
		queryAST.Expressions = append(queryAST.Expressions, search.TermExpr{Value: query})
	}

	opts := search.FindOpts{Query: queryAST}.WithLimit(topK)
	if queryAST.IsEmpty() {
		opts = opts.WithSort(search.SortByPath, search.SortAsc)
	}

	results, err := s.index.Find(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("keyword retrieval failed: %w", err)
	}

	return results.Items, nil
}

func (s *NoteService) findSemanticCandidates(
	ctx context.Context,
	query string,
	conditions []QueryCondition,
	topK int,
) ([]SemanticResult, error) {
	candidates, err := s.FindSemanticCandidates(ctx, query, topK)
	if err != nil {
		return nil, err
	}

	if len(conditions) == 0 {
		return candidates, nil
	}

	queryAST, err := s.searchService.BuildQuery(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	if query != "" {
		queryAST.Expressions = append(queryAST.Expressions, search.TermExpr{Value: query})
	}

	filtered := make([]SemanticResult, 0, len(candidates))
	for _, candidate := range candidates {
		if queryMatchesDocument(queryAST, candidate.Document) {
			filtered = append(filtered, candidate)
		}
	}

	return filtered, nil
}

func searchResultsToNotes(results []search.Result) []Note {
	notes := make([]Note, len(results))
	for i, result := range results {
		notes[i] = documentToNote(result.Document)
	}
	return notes
}

func semanticResultsToNotes(results []SemanticResult) []Note {
	notes := make([]Note, len(results))
	for i, result := range results {
		notes[i] = documentToNote(result.Document)
	}
	return notes
}

func hybridResultsToNotes(results []HybridResult) []Note {
	notes := make([]Note, len(results))
	for i, result := range results {
		notes[i] = documentToNote(result.Document)
	}
	return notes
}

func queryMatchesDocument(query *search.Query, doc search.Document) bool {
	if query == nil || len(query.Expressions) == 0 {
		return true
	}

	for _, expr := range query.Expressions {
		if !exprMatchesDocument(expr, doc) {
			return false
		}
	}

	return true
}

func exprMatchesDocument(expr search.Expr, doc search.Document) bool {
	switch e := expr.(type) {
	case search.TermExpr:
		term := strings.ToLower(strings.TrimSpace(e.Value))
		if term == "" {
			return true
		}
		haystack := strings.ToLower(doc.Title + "\n" + doc.Body + "\n" + doc.Path)
		return strings.Contains(haystack, term)

	case search.FieldExpr:
		actuals := fieldValuesForExpr(doc, e.Field)
		for _, actual := range actuals {
			if compareValue(e.Op, actual, e.Value) {
				return true
			}
		}
		return false

	case search.WildcardExpr:
		if e.Field != "path" {
			return false
		}
		return globMatch(strings.ToLower(e.Pattern), strings.ToLower(doc.Path))

	case search.NotExpr:
		return !exprMatchesDocument(e.Expr, doc)

	case search.OrExpr:
		return exprMatchesDocument(e.Left, doc) || exprMatchesDocument(e.Right, doc)

	default:
		return false
	}
}

func fieldValuesForExpr(doc search.Document, field string) []string {
	switch {
	case field == "path":
		return []string{doc.Path}
	case field == "title":
		return []string{doc.Title}
	case strings.HasPrefix(field, "metadata."):
		key := strings.TrimPrefix(field, "metadata.")
		values := make([]string, 0)

		if key == "tag" {
			values = append(values, doc.Tags...)
		}

		if doc.Metadata != nil {
			if raw, ok := doc.Metadata[key]; ok {
				values = append(values, metadataValues(raw)...)
			}
		}

		return values
	default:
		return nil
	}
}

func metadataValues(raw any) []string {
	switch v := raw.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		values := make([]string, 0, len(v))
		for _, item := range v {
			values = append(values, fmt.Sprintf("%v", item))
		}
		return values
	default:
		return []string{fmt.Sprintf("%v", raw)}
	}
}

func compareValue(op search.CompareOp, actual, expected string) bool {
	actualLower := strings.ToLower(actual)
	expectedLower := strings.ToLower(expected)

	switch op {
	case search.OpEquals:
		return actualLower == expectedLower
	case search.OpPrefix:
		return strings.HasPrefix(actualLower, expectedLower)
	case search.OpSuffix:
		return strings.HasSuffix(actualLower, expectedLower)
	default:
		return false
	}
}

func globMatch(pattern, value string) bool {
	re := regexp.QuoteMeta(pattern)
	re = strings.ReplaceAll(re, `\*\*`, `.*`)
	re = strings.ReplaceAll(re, `\*`, `[^/]*`)
	re = strings.ReplaceAll(re, `\?`, `.`)

	compiled := "^" + re + "$"
	matched, err := regexp.MatchString(compiled, value)
	if err != nil {
		return false
	}
	return matched
}
