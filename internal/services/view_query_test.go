// internal/services/view_query_test.go
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitViewQuery(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFilter string
		wantDirs   string
	}{
		{
			name:       "simple pipe split",
			input:      "tag:work | sort:modified:desc",
			wantFilter: "tag:work",
			wantDirs:   "sort:modified:desc",
		},
		{
			name:       "no pipe returns filter only",
			input:      "tag:work status:todo",
			wantFilter: "tag:work status:todo",
			wantDirs:   "",
		},
		{
			name:       "empty filter with directives",
			input:      "| sort:modified:desc limit:20",
			wantFilter: "",
			wantDirs:   "sort:modified:desc limit:20",
		},
		{
			name:       "pipe inside quoted string is not split",
			input:      `title:"A | B" tag:work`,
			wantFilter: `title:"A | B" tag:work`,
			wantDirs:   "",
		},
		{
			name:       "pipe after quoted string splits correctly",
			input:      `title:"A | B" | sort:title:asc`,
			wantFilter: `title:"A | B"`,
			wantDirs:   "sort:title:asc",
		},
		{
			name:       "trims whitespace",
			input:      "  tag:work  |  sort:modified:desc  ",
			wantFilter: "tag:work",
			wantDirs:   "sort:modified:desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, dirs := SplitViewQuery(tt.input)
			assert.Equal(t, tt.wantFilter, filter)
			assert.Equal(t, tt.wantDirs, dirs)
		})
	}
}

func TestParseDirectives(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSort  string
		wantDir   string
		wantLimit int
		wantGroup string
		wantErr   bool
	}{
		{
			name:      "sort directive",
			input:     "sort:modified:desc",
			wantSort:  "modified",
			wantDir:   "desc",
			wantLimit: 0,
		},
		{
			name:      "sort with default direction",
			input:     "sort:title",
			wantSort:  "title",
			wantDir:   "asc",
			wantLimit: 0,
		},
		{
			name:      "limit directive",
			input:     "limit:20",
			wantLimit: 20,
		},
		{
			name:      "group directive",
			input:     "group:status",
			wantGroup: "status",
		},
		{
			name:      "multiple directives",
			input:     "sort:modified:desc limit:50 group:status",
			wantSort:  "modified",
			wantDir:   "desc",
			wantLimit: 50,
			wantGroup: "status",
		},
		{
			name:    "unknown directive errors",
			input:   "foo:bar",
			wantErr: true,
		},
		{
			name:      "case insensitive",
			input:     "Sort:Modified:DESC Limit:10",
			wantSort:  "modified",
			wantDir:   "desc",
			wantLimit: 10,
		},
		{
			name:      "last directive wins on conflict",
			input:     "limit:10 limit:20",
			wantLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ParseDirectives(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSort, d.SortField)
			assert.Equal(t, tt.wantDir, d.SortDirection)
			assert.Equal(t, tt.wantLimit, d.Limit)
			assert.Equal(t, tt.wantGroup, d.GroupBy)
		})
	}
}
