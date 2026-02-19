package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zenobi-us/jot/internal/search"
)

func TestMergeHybridResults_LabelsAndOrdering(t *testing.T) {
	keyword := []search.Result{
		{Document: search.Document{Path: "notes/a.md"}, Score: 0.8},
		{Document: search.Document{Path: "notes/b.md"}, Score: 0.7},
		{Document: search.Document{Path: "notes/c.md"}, Score: 0.6},
	}
	semantic := []SemanticResult{
		{Document: search.Document{Path: "notes/b.md"}, Score: 0.95},
		{Document: search.Document{Path: "notes/d.md"}, Score: 0.90},
		{Document: search.Document{Path: "notes/a.md"}, Score: 0.85},
	}

	merged := MergeHybridResults(keyword, semantic, 60)

	assert.Len(t, merged, 4)
	assert.Equal(t, "notes/b.md", merged[0].Document.Path)
	assert.Equal(t, MatchTypeHybrid, merged[0].MatchType)

	assert.Equal(t, "notes/a.md", merged[1].Document.Path)
	assert.Equal(t, MatchTypeHybrid, merged[1].MatchType)

	assert.Equal(t, "notes/d.md", merged[2].Document.Path)
	assert.Equal(t, MatchTypeSemantic, merged[2].MatchType)

	assert.Equal(t, "notes/c.md", merged[3].Document.Path)
	assert.Equal(t, MatchTypeExact, merged[3].MatchType)
}

func TestMergeHybridResults_KeywordOnlyFallbackKeepsOrder(t *testing.T) {
	keyword := []search.Result{
		{Document: search.Document{Path: "notes/x.md"}, Score: 0.5},
		{Document: search.Document{Path: "notes/y.md"}, Score: 0.4},
	}

	merged := MergeHybridResults(keyword, nil, 60)
	assert.Len(t, merged, 2)
	assert.Equal(t, "notes/x.md", merged[0].Document.Path)
	assert.Equal(t, MatchTypeExact, merged[0].MatchType)
	assert.Equal(t, "notes/y.md", merged[1].Document.Path)
	assert.Equal(t, MatchTypeExact, merged[1].MatchType)
}

func TestMergeHybridResults_SemanticOnlyFallbackKeepsOrder(t *testing.T) {
	semantic := []SemanticResult{
		{Document: search.Document{Path: "notes/m.md"}, Score: 0.9},
		{Document: search.Document{Path: "notes/n.md"}, Score: 0.8},
	}

	merged := MergeHybridResults(nil, semantic, 60)
	assert.Len(t, merged, 2)
	assert.Equal(t, "notes/m.md", merged[0].Document.Path)
	assert.Equal(t, MatchTypeSemantic, merged[0].MatchType)
	assert.Equal(t, "notes/n.md", merged[1].Document.Path)
	assert.Equal(t, MatchTypeSemantic, merged[1].MatchType)
}

func TestMergeHybridResults_PathTieBreakDeterministic(t *testing.T) {
	keyword := []search.Result{
		{Document: search.Document{Path: "notes/z.md"}, Score: 0.5},
	}
	semantic := []SemanticResult{
		{Document: search.Document{Path: "notes/a.md"}, Score: 0.9},
	}

	// rrfK=0 is normalized to 60, but equal rank (1 vs 1) still ties on score.
	merged := MergeHybridResults(keyword, semantic, 60)
	assert.Len(t, merged, 2)

	assert.Equal(t, "notes/a.md", merged[0].Document.Path)
	assert.Equal(t, "notes/z.md", merged[1].Document.Path)
}

func TestMergeHybridResults_DefaultsRRFKWhenNonPositive(t *testing.T) {
	keyword := []search.Result{{Document: search.Document{Path: "notes/a.md"}, Score: 1.0}}
	semantic := []SemanticResult{{Document: search.Document{Path: "notes/b.md"}, Score: 1.0}}

	withZero := MergeHybridResults(keyword, semantic, 0)
	withDefault := MergeHybridResults(keyword, semantic, 60)

	assert.Equal(t, withDefault, withZero)
}
