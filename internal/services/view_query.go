// Package services provides view query parsing utilities.
package services

import (
	"fmt"
	"strconv"
	"strings"
)

// SplitViewQuery splits a view query string on the first unquoted pipe character.
// Returns (filter, directives) where filter is the DSL query portion
// and directives is the presentation options portion.
// Pipe characters inside quoted strings are preserved.
func SplitViewQuery(query string) (filter, directives string) {
	inQuote := false
	for i, ch := range query {
		switch ch {
		case '"':
			inQuote = !inQuote
		case '|':
			if !inQuote {
				return strings.TrimSpace(query[:i]), strings.TrimSpace(query[i+1:])
			}
		}
	}
	return strings.TrimSpace(query), ""
}

// ViewDirectives holds parsed presentation options from the directive portion of a view query.
type ViewDirectives struct {
	SortField     string
	SortDirection string // "asc" or "desc"
	Limit         int
	Offset        int
	GroupBy       string
}

// ParseDirectives parses the directive portion of a view query.
// Valid directives: sort:<field>:<asc|desc>, limit:<n>, offset:<n>, group:<field>
// Directives are case-insensitive. Last directive wins on conflict.
func ParseDirectives(input string) (*ViewDirectives, error) {
	d := &ViewDirectives{}

	if strings.TrimSpace(input) == "" {
		return d, nil
	}

	parts := strings.Fields(input)
	for _, part := range parts {
		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid directive %q: missing colon", part)
		}

		key := strings.ToLower(part[:colonIdx])
		value := part[colonIdx+1:]

		switch key {
		case "sort":
			// sort:field or sort:field:dir
			sortParts := strings.Split(value, ":")
			d.SortField = strings.ToLower(sortParts[0])
			if len(sortParts) > 1 {
				d.SortDirection = strings.ToLower(sortParts[1])
			} else {
				d.SortDirection = "asc" // default direction when sort is specified
			}
		case "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid limit %q: %w", value, err)
			}
			d.Limit = n
		case "offset":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid offset %q: %w", value, err)
			}
			d.Offset = n
		case "group":
			d.GroupBy = strings.ToLower(value)
		default:
			return nil, fmt.Errorf("unknown directive %q. Valid: sort, limit, offset, group", key)
		}
	}

	return d, nil
}
