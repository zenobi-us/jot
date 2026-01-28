package core

import "encoding/json"

// ViewDefinition represents a named, reusable query preset
type ViewDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ViewParameter `json:"parameters,omitempty"`
	Query       ViewQuery       `json:"query"`
}

// ViewInfo represents view metadata for discovery/listing (includes origin)
type ViewInfo struct {
	Name        string          `json:"name"`
	Origin      string          `json:"origin"` // "built-in", "global", "notebook"
	Description string          `json:"description"`
	Parameters  []ViewParameter `json:"parameters,omitempty"`
}

// ViewParameter represents a dynamic parameter in a view
type ViewParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "list", "date", "bool"
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// ViewQuery represents the query logic for a view
type ViewQuery struct {
	Conditions       []ViewCondition   `json:"conditions,omitempty"`
	Distinct         bool              `json:"distinct,omitempty"`
	OrderBy          string            `json:"order_by,omitempty"`
	GroupBy          string            `json:"group_by,omitempty"`
	Having           []ViewCondition   `json:"having,omitempty"`
	SelectColumns    []string          `json:"select_columns,omitempty"`
	AggregateColumns map[string]string `json:"aggregate_columns,omitempty"`
	Limit            int               `json:"limit,omitempty"`
	Offset           int               `json:"offset,omitempty"`
}

// ViewCondition represents a single query condition
type ViewCondition struct {
	Logic    string `json:"logic,omitempty"` // "AND", "OR"
	Field    string `json:"field"`
	Operator string `json:"operator"` // "=", "!=", "<", ">", "<=", ">=", "LIKE", "IN", "IS NULL"
	Value    string `json:"value"`
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
