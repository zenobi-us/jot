package core

import "encoding/json"

// ViewDefinition defines a named view with a DSL query string.
type ViewDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ViewParameter `json:"parameters,omitempty"`
	Query       string          `json:"query"`          // "filter DSL | directives"
	Type        string          `json:"type,omitempty"` // "query" (default) or "special"
}

// ViewInfo represents view metadata for discovery/listing (includes origin)
type ViewInfo struct {
	Name        string          `json:"name"`
	Origin      string          `json:"origin"` // "built-in", "global", "notebook"
	Description string          `json:"description"`
	Parameters  []ViewParameter `json:"parameters,omitempty"`
}

// ViewParameter defines a parameter that can be substituted into a view query.
type ViewParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "date"
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// IsSpecialView returns true if this view requires special execution (not DSL-based).
func (v *ViewDefinition) IsSpecialView() bool {
	return v.Type == "special"
}

// ViewsConfig represents the views section in config files
type ViewsConfig struct {
	Views map[string]json.RawMessage `json:"views"`
}

// ParseViewDefinition parses raw JSON into a ViewDefinition
func ParseViewDefinition(data json.RawMessage) (*ViewDefinition, error) {
	var view ViewDefinition
	if err := json.Unmarshal(data, &view); err != nil {
		return nil, err
	}
	return &view, nil
}
