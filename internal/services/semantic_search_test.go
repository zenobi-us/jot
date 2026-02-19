package services

import (
	"strings"
	"testing"
)

func TestBuildExplainSnippet_HighlightsKeyword(t *testing.T) {
	body := "Sprint planning includes a meeting recap and action list for this week."

	snippet := buildExplainSnippet(body, "meeting", MatchTypeExact)

	if !strings.Contains(snippet, "[meeting]") {
		t.Fatalf("expected highlighted query in snippet, got %q", snippet)
	}
}

func TestBuildExplainSnippet_NoSnippetFallback(t *testing.T) {
	snippet := buildExplainSnippet("", "meeting", MatchTypeExact)

	if snippet != "No snippet available" {
		t.Fatalf("expected no-snippet fallback, got %q", snippet)
	}
}

func TestBuildExplainSnippet_TruncatesLongSemanticSnippet(t *testing.T) {
	body := strings.Repeat("semantic context about architecture decisions and trade-offs. ", 8)

	snippet := buildExplainSnippet(body, "architecture", MatchTypeSemantic)

	if len(snippet) > 160 {
		t.Fatalf("expected semantic snippet length <= 160, got %d", len(snippet))
	}
	if !strings.HasSuffix(snippet, "...") {
		t.Fatalf("expected truncated snippet to end with ellipsis, got %q", snippet)
	}
}
