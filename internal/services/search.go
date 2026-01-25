package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/sahilm/fuzzy"
)

// QueryCondition represents a single search condition for boolean queries.
type QueryCondition struct {
	Type     string // "and", "or", "not"
	Field    string // "data.tag", "data.status", "path", "title", "links-to", "linked-by"
	Operator string // "=" (currently only equality is supported)
	Value    string // user-provided value
}

// AllowedFields is the whitelist of valid fields for security.
// Only these fields can be queried via boolean conditions.
var AllowedFields = map[string]bool{
	"data.tag":      true,
	"data.tags":     true,
	"data.status":   true,
	"data.priority": true,
	"data.assignee": true,
	"data.author":   true,
	"data.type":     true,
	"data.category": true,
	"data.project":  true,
	"data.sprint":   true,
	"path":          true,
	"title":         true,
	"links-to":      true,
	"linked-by":     true,
}

// MaxValueLength is the maximum allowed length for condition values (security).
const MaxValueLength = 1000

// SearchService provides search operations for notes.
type SearchService struct {
	log zerolog.Logger
}

// NewSearchService creates a new search service.
func NewSearchService() *SearchService {
	return &SearchService{
		log: Log("SearchService"),
	}
}

// fuzzyMatch represents a note with its fuzzy match score.
type fuzzyMatch struct {
	note  Note
	score int
}

// FuzzySearch performs fuzzy matching on notes and returns them ranked by score.
// If query is empty, returns all notes unsorted.
// Title matches are weighted 2x higher than body matches.
func (s *SearchService) FuzzySearch(query string, notes []Note) []Note {
	if len(notes) == 0 {
		return nil
	}

	// Empty query - return all notes
	if query == "" {
		return notes
	}

	var matches []fuzzyMatch

	for _, note := range notes {
		titleScore := 0
		bodyScore := 0

		// Try fuzzy matching on title
		title := note.DisplayName()
		if title != "" {
			titleMatches := fuzzy.Find(query, []string{title})
			if len(titleMatches) > 0 {
				// Title matches are weighted 2x higher
				titleScore = titleMatches[0].Score * 2
			}
		}

		// Try fuzzy matching on body preview (first 500 chars for performance)
		bodyPreview := note.Content
		if len(bodyPreview) > 500 {
			bodyPreview = bodyPreview[:500]
		}
		if bodyPreview != "" {
			bodyMatches := fuzzy.Find(query, []string{bodyPreview})
			if len(bodyMatches) > 0 {
				bodyScore = bodyMatches[0].Score
			}
		}

		// Take the best score
		score := titleScore
		if bodyScore > score {
			score = bodyScore
		}

		// Only include if there's a match
		if score > 0 {
			matches = append(matches, fuzzyMatch{
				note:  note,
				score: score,
			})
		}
	}

	// Sort by score descending (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	// Extract sorted notes
	result := make([]Note, len(matches))
	for i, match := range matches {
		result[i] = match.note
	}

	s.log.Debug().
		Str("query", query).
		Int("total_notes", len(notes)).
		Int("matches", len(result)).
		Msg("fuzzy search completed")

	return result
}

// TextSearch performs exact text matching on notes.
// Searches both content and filepath (case-insensitive).
func (s *SearchService) TextSearch(query string, notes []Note) []Note {
	if query == "" {
		return notes
	}

	var matches []Note
	queryLower := strings.ToLower(query)

	for _, note := range notes {
		// Check content
		if strings.Contains(strings.ToLower(note.Content), queryLower) {
			matches = append(matches, note)
			continue
		}

		// Check filepath
		if strings.Contains(strings.ToLower(note.File.Filepath), queryLower) {
			matches = append(matches, note)
			continue
		}
	}

	s.log.Debug().
		Str("query", query).
		Int("total_notes", len(notes)).
		Int("matches", len(matches)).
		Msg("text search completed")

	return matches
}

// ParseConditions parses CLI flags into QueryConditions.
// andFlags, orFlags, notFlags are arrays of "field=value" strings.
func (s *SearchService) ParseConditions(andFlags, orFlags, notFlags []string) ([]QueryCondition, error) {
	var conditions []QueryCondition

	for _, flag := range andFlags {
		cond, err := s.parseCondition("and", flag)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	for _, flag := range orFlags {
		cond, err := s.parseCondition("or", flag)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	for _, flag := range notFlags {
		cond, err := s.parseCondition("not", flag)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	return conditions, nil
}

// parseCondition parses a single "field=value" string into a QueryCondition.
func (s *SearchService) parseCondition(condType, flag string) (QueryCondition, error) {
	parts := strings.SplitN(flag, "=", 2)
	if len(parts) != 2 {
		return QueryCondition{}, fmt.Errorf("invalid condition format: %s (expected field=value)", flag)
	}

	field, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	// Validate field (security - whitelist only)
	if !AllowedFields[field] {
		return QueryCondition{}, fmt.Errorf("invalid field: %s (allowed: data.tag, data.status, data.priority, data.assignee, data.author, data.type, data.category, data.project, data.sprint, path, title, links-to, linked-by)", field)
	}

	// Validate value length (security)
	if len(value) > MaxValueLength {
		return QueryCondition{}, fmt.Errorf("value too long (max %d chars)", MaxValueLength)
	}

	// Validate value is not empty
	if value == "" {
		return QueryCondition{}, fmt.Errorf("value cannot be empty for field: %s", field)
	}

	return QueryCondition{
		Type:     condType,
		Field:    field,
		Operator: "=",
		Value:    value,
	}, nil
}

// BuildWhereClause constructs a parameterized SQL WHERE clause from conditions.
// Returns the WHERE clause (without "WHERE"), the parameters, and any error.
// SECURITY: Always use parameterized queries - never concatenate values into SQL.
// Deprecated: Use BuildWhereClauseWithGlob for queries that may include linked-by conditions.
func (s *SearchService) BuildWhereClause(conditions []QueryCondition) (string, []interface{}, error) {
	return s.BuildWhereClauseWithGlob(conditions, "")
}

// BuildWhereClauseWithGlob constructs a parameterized SQL WHERE clause from conditions.
// The notebookGlob parameter is required for linked-by queries to access the source document.
// Returns the WHERE clause (without "WHERE"), the parameters, and any error.
// SECURITY: Always use parameterized queries - never concatenate values into SQL.
func (s *SearchService) BuildWhereClauseWithGlob(conditions []QueryCondition, notebookGlob string) (string, []interface{}, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}

	var andParts []string
	var orParts []string
	var notParts []string
	var params []interface{}

	for _, cond := range conditions {
		sqlPart, condParams, err := s.buildConditionSQL(cond, notebookGlob)
		if err != nil {
			return "", nil, err
		}

		switch cond.Type {
		case "and":
			andParts = append(andParts, sqlPart)
		case "or":
			orParts = append(orParts, sqlPart)
		case "not":
			notParts = append(notParts, fmt.Sprintf("NOT (%s)", sqlPart))
		}
		params = append(params, condParams...)
	}

	var whereParts []string

	// AND conditions are joined with AND
	if len(andParts) > 0 {
		whereParts = append(whereParts, strings.Join(andParts, " AND "))
	}

	// OR conditions are grouped together with parentheses
	if len(orParts) > 0 {
		whereParts = append(whereParts, fmt.Sprintf("(%s)", strings.Join(orParts, " OR ")))
	}

	// NOT conditions are each negated and joined with AND
	if len(notParts) > 0 {
		whereParts = append(whereParts, strings.Join(notParts, " AND "))
	}

	whereClause := strings.Join(whereParts, " AND ")

	s.log.Debug().
		Str("whereClause", whereClause).
		Int("paramCount", len(params)).
		Int("conditionCount", len(conditions)).
		Msg("built WHERE clause")

	return whereClause, params, nil
}

// buildConditionSQL builds the SQL fragment for a single condition.
// Returns the SQL fragment (with ? placeholders) and the parameter values.
// The notebookGlob is used for linked-by queries that need to access source documents.
func (s *SearchService) buildConditionSQL(cond QueryCondition, notebookGlob string) (string, []interface{}, error) {
	switch {
	case strings.HasPrefix(cond.Field, "data."):
		// Frontmatter field - use metadata map access
		// For DuckDB, we need to access the metadata MAP column
		fieldName := strings.TrimPrefix(cond.Field, "data.")
		// Using DuckDB MAP syntax: metadata[fieldName] = value
		// Note: We use COALESCE to handle NULL gracefully
		sqlPart := "COALESCE(metadata[?], '') = ?"
		return sqlPart, []interface{}{fieldName, cond.Value}, nil

	case cond.Field == "path":
		// Path field - use filepath with wildcard prefix to match anywhere in path
		// This allows users to specify relative paths like "epics/*" that match
		// the full absolute paths like "/home/user/notebook/epics/epic1.md"
		sqlPart := "file_path LIKE ?"
		// Convert glob-like patterns to LIKE patterns
		likePattern := GlobToLike(cond.Value)
		// Add % prefix to match from any point in the path
		if !strings.HasPrefix(likePattern, "%") {
			likePattern = "%" + likePattern
		}
		return sqlPart, []interface{}{likePattern}, nil

	case cond.Field == "title":
		// Title field - check metadata title or use filename
		sqlPart := "(COALESCE(metadata['title'], '') = ? OR file_path LIKE ?)"
		return sqlPart, []interface{}{cond.Value, "%" + cond.Value + "%"}, nil

	case cond.Field == "links-to":
		// Links-to: find documents whose links array contains the target
		// This uses DuckDB's UNNEST and LIKE for glob support
		likePattern := GlobToLike(cond.Value)
		sqlPart := `EXISTS (
			SELECT 1 FROM (
				SELECT unnest(COALESCE(TRY_CAST(metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
			) AS links_table
			WHERE link LIKE ?
		)`
		return sqlPart, []interface{}{likePattern}, nil

	case cond.Field == "linked-by":
		// Linked-by: find documents that the specified source links TO
		// Returns documents that appear in the source's links array
		// Graph: if source --links--> target, linked-by=source returns target
		//
		// We use a subquery that reads the source document and extracts its links,
		// then check if the current file_path matches any of those links.
		if notebookGlob == "" {
			return "", nil, fmt.Errorf("linked-by queries require notebook context")
		}
		likePattern := GlobToLike(cond.Value)
		// Use LIKE with the link to allow for relative path matching
		// The file_path might be absolute while links might be relative
		sqlPart := `EXISTS (
			SELECT 1 FROM (
				SELECT unnest(COALESCE(TRY_CAST(src.metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
				FROM read_markdown(?, include_filepath:=true) AS src
				WHERE src.file_path LIKE ?
			) AS source_links
			WHERE file_path LIKE '%' || source_links.link OR file_path LIKE '%/' || source_links.link
		)`
		return sqlPart, []interface{}{notebookGlob, "%" + likePattern}, nil

	default:
		return "", nil, fmt.Errorf("unsupported field: %s", cond.Field)
	}
}

// GlobToLike converts glob patterns to SQL LIKE patterns.
// ** -> %
// * -> %
// ? -> _
// Also escapes SQL special characters (%, _) before conversion.
func GlobToLike(pattern string) string {
	// First escape SQL special characters that are NOT glob characters
	// We need to be careful here - if user has literal % or _ in their pattern,
	// we need to escape them first before converting glob chars
	result := strings.ReplaceAll(pattern, "%", "\\%")
	result = strings.ReplaceAll(result, "_", "\\_")

	// Now convert glob patterns to SQL LIKE patterns
	// Handle ** first (matches any path including subdirectories)
	result = strings.ReplaceAll(result, "**", "{{DOUBLESTAR}}")
	// Then handle * (matches any characters)
	result = strings.ReplaceAll(result, "*", "%")
	// Replace the placeholder back
	result = strings.ReplaceAll(result, "{{DOUBLESTAR}}", "%")
	// Handle ? (matches single character)
	result = strings.ReplaceAll(result, "?", "_")

	return result
}
