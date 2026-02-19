package parser

import (
	"testing"

	"github.com/zenobi-us/jot/internal/search"
)

func TestParser_HasKeyword(t *testing.T) {
	p := New()

	tests := []struct {
		name      string
		input     string
		wantField string
	}{
		{
			name:      "has:tag matches notes with any tag",
			input:     "has:tag",
			wantField: "tag",
		},
		{
			name:      "has:status matches notes with status field",
			input:     "has:status",
			wantField: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(result.Expressions) != 1 {
				t.Fatalf("got %d expressions, want 1", len(result.Expressions))
			}

			exists, ok := result.Expressions[0].(search.ExistsExpr)
			if !ok {
				t.Fatalf("expected ExistsExpr, got %T", result.Expressions[0])
			}

			if exists.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", exists.Field, tt.wantField)
			}
			if exists.Negated {
				t.Errorf("Negated = true, want false for has: keyword")
			}
		})
	}
}

func TestParser_MissingKeyword(t *testing.T) {
	p := New()

	tests := []struct {
		name      string
		input     string
		wantField string
	}{
		{
			name:      "missing:tag matches notes without tags",
			input:     "missing:tag",
			wantField: "tag",
		},
		{
			name:      "missing:status matches notes without status field",
			input:     "missing:status",
			wantField: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(result.Expressions) != 1 {
				t.Fatalf("got %d expressions, want 1", len(result.Expressions))
			}

			exists, ok := result.Expressions[0].(search.ExistsExpr)
			if !ok {
				t.Fatalf("expected ExistsExpr, got %T", result.Expressions[0])
			}

			if exists.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", exists.Field, tt.wantField)
			}
			if !exists.Negated {
				t.Errorf("Negated = false, want true for missing: keyword")
			}
		})
	}
}

func TestParser_ExistenceWithOtherTerms(t *testing.T) {
	p := New()

	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "existence with field match",
			input:   "has:tag status:todo",
			wantLen: 2,
		},
		{
			name:    "existence with term",
			input:   "has:status meeting",
			wantLen: 2,
		},
		{
			name:    "multiple existence checks",
			input:   "has:tag missing:status",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(result.Expressions) != tt.wantLen {
				t.Errorf("got %d expressions, want %d", len(result.Expressions), tt.wantLen)
			}

			// First expression should be ExistsExpr
			_, ok := result.Expressions[0].(search.ExistsExpr)
			if !ok {
				t.Errorf("first expression is %T, want ExistsExpr", result.Expressions[0])
			}
		})
	}
}
