package search

// Parser defines the interface for parsing query strings into AST.
//
// The parser converts Gmail-style query strings into structured Query objects.
// Different implementations can use different parsing strategies (Participle,
// hand-written recursive descent, etc.).
//
// Example:
//
//	parser := participle.NewParser()
//	query, err := parser.Parse("tag:work title:meeting -archived")
//	// query.Expressions contains:
//	// - FieldExpr{Field: "tag", Value: "work"}
//	// - FieldExpr{Field: "title", Value: "meeting"}
//	// - NotExpr{Expr: TermExpr{Value: "archived"}}
type Parser interface {
	// Parse parses a query string into a Query AST.
	// Returns a ParseError if the query is malformed.
	Parse(input string) (*Query, error)

	// Validate checks if a query string is syntactically valid.
	// This is faster than Parse when you don't need the AST.
	Validate(input string) error

	// Help returns syntax help for the query language.
	// Used to display help text to users.
	Help() string
}

// ParseError represents an error during query parsing.
type ParseError struct {
	// Message is a human-readable error description
	Message string

	// Position is the character offset where the error occurred
	Position int

	// Line is the line number (1-indexed) for multi-line queries
	Line int

	// Column is the column number (1-indexed)
	Column int

	// Input is the original query string
	Input string

	// Suggestion is an optional fix suggestion
	Suggestion string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.Suggestion != "" {
		return e.Message + ". Did you mean: " + e.Suggestion + "?"
	}
	return e.Message
}

// SupportedFields returns the list of fields recognized by the query language.
// This is used for validation and autocompletion.
func SupportedFields() []FieldSpec {
	return []FieldSpec{
		{Name: "tag", Description: "Filter by tag", Example: "tag:work"},
		{Name: "title", Description: "Search in title", Example: "title:meeting"},
		{Name: "path", Description: "Filter by path prefix", Example: "path:projects/"},
		{Name: "created", Description: "Filter by creation date", Example: "created:>2024-01-01"},
		{Name: "modified", Description: "Filter by modification date", Example: "modified:<2024-06-30"},
		{Name: "body", Description: "Search in body only", Example: "body:important"},
	}
}

// FieldSpec describes a supported query field.
type FieldSpec struct {
	// Name is the field name used in queries
	Name string

	// Description is a human-readable description
	Description string

	// Example shows how to use the field
	Example string

	// SupportsWildcard indicates if wildcards are allowed
	SupportsWildcard bool

	// SupportsRange indicates if range queries are allowed
	SupportsRange bool

	// SupportsComparison indicates if comparison operators are allowed
	SupportsComparison bool
}
