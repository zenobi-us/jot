package services

import (
	"sort"
	"strings"

	"github.com/zenobi-us/jot/internal/search"
)

// MatchType identifies how a note entered the hybrid result set.
type MatchType string

const (
	MatchTypeExact    MatchType = "Exact match"
	MatchTypeSemantic MatchType = "Semantic match"
	MatchTypeHybrid   MatchType = "Hybrid"
)

// HybridResult is the merged result record used by semantic+keyword retrieval.
type HybridResult struct {
	Document      search.Document
	Score         float64
	MatchType     MatchType
	KeywordRank   int
	SemanticRank  int
	KeywordScore  float64
	SemanticScore float64
}

// MergeHybridResults merges keyword and semantic candidates using Reciprocal Rank Fusion (RRF).
// Ordering rules:
//  1. Descending RRF score
//  2. Descending source coverage (in both lists first)
//  3. Ascending normalized path (stable tie-break)
func MergeHybridResults(keyword []search.Result, semantic []SemanticResult, rrfK int) []HybridResult {
	if rrfK <= 0 {
		rrfK = 60
	}

	byPath := make(map[string]*HybridResult)

	for i, result := range keyword {
		rank := i + 1
		pathKey := normalizePathForSort(result.Document.Path)

		entry := byPath[pathKey]
		if entry == nil {
			entry = &HybridResult{Document: result.Document}
			byPath[pathKey] = entry
		}

		if entry.KeywordRank == 0 || rank < entry.KeywordRank {
			entry.KeywordRank = rank
			entry.KeywordScore = result.Score
		}
		entry.Score += reciprocalRank(rrfK, rank)
	}

	for i, result := range semantic {
		rank := i + 1
		pathKey := normalizePathForSort(result.Document.Path)

		entry := byPath[pathKey]
		if entry == nil {
			entry = &HybridResult{Document: result.Document}
			byPath[pathKey] = entry
		}

		if entry.SemanticRank == 0 || rank < entry.SemanticRank {
			entry.SemanticRank = rank
			entry.SemanticScore = result.Score
		}
		entry.Score += reciprocalRank(rrfK, rank)
	}

	merged := make([]HybridResult, 0, len(byPath))
	for _, result := range byPath {
		result.MatchType = classifyMatchType(result.KeywordRank, result.SemanticRank)
		merged = append(merged, *result)
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Score != merged[j].Score {
			return merged[i].Score > merged[j].Score
		}

		iCoverage := sourceCoverage(merged[i])
		jCoverage := sourceCoverage(merged[j])
		if iCoverage != jCoverage {
			return iCoverage > jCoverage
		}

		return normalizePathForSort(merged[i].Document.Path) < normalizePathForSort(merged[j].Document.Path)
	})

	return merged
}

func reciprocalRank(rrfK, rank int) float64 {
	return 1.0 / float64(rrfK+rank)
}

func classifyMatchType(keywordRank, semanticRank int) MatchType {
	if keywordRank > 0 && semanticRank > 0 {
		return MatchTypeHybrid
	}
	if keywordRank > 0 {
		return MatchTypeExact
	}
	return MatchTypeSemantic
}

func sourceCoverage(result HybridResult) int {
	coverage := 0
	if result.KeywordRank > 0 {
		coverage++
	}
	if result.SemanticRank > 0 {
		coverage++
	}
	return coverage
}

func normalizePathForSort(path string) string {
	return strings.ToLower(strings.TrimSpace(path))
}
