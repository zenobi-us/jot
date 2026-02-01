package search

// Query represents a parsed search query.
//
// The Query is the root of the AST produced by parsing a query string.
// It contains a tree of expressions that represent the search criteria.
//
// Example query string -> AST:
//
//	"tag:work title:meeting -archived"
//
//	Query{
//	    Expressions: []Expr{
//	        FieldExpr{Field: "tag", Value: "work"},
//	        FieldExpr{Field: "title", Value: "meeting"},
//	        NotExpr{Expr: TermExpr{Value: "archived"}},
//	    },
//	}
type Query struct {
	// Expressions is the list of query expressions (implicit AND)
	Expressions []Expr

	// Raw is the original query string
	Raw string
}

// IsEmpty returns true if the query has no expressions.
func (q *Query) IsEmpty() bool {
	return q == nil || len(q.Expressions) == 0
}

// Expr is the interface for all query expression types.
type Expr interface {
	// exprNode is a marker method to ensure only defined types implement Expr.
	exprNode()
}

// TermExpr represents a simple text search term.
//
// Example: "meeting" searches for "meeting" in all text fields.
type TermExpr struct {
	// Value is the search term
	Value string
}

func (TermExpr) exprNode() {}

// FieldExpr represents a field-qualified search.
//
// Example: "tag:work" searches for notes with tag "work".
type FieldExpr struct {
	// Field is the field name (tag, title, path, etc.)
	Field string

	// Op is the comparison operator (defaults to Equals)
	Op CompareOp

	// Value is the field value to match
	Value string
}

func (FieldExpr) exprNode() {}

// CompareOp is a comparison operator for field expressions.
type CompareOp string

const (
	OpEquals CompareOp = "="  // Exact match or contains
	OpPrefix CompareOp = "^"  // Starts with
	OpSuffix CompareOp = "$"  // Ends with
	OpGt     CompareOp = ">"  // Greater than (for dates/numbers)
	OpGte    CompareOp = ">=" // Greater than or equal
	OpLt     CompareOp = "<"  // Less than
	OpLte    CompareOp = "<=" // Less than or equal
)

// NotExpr represents a negated expression.
//
// Example: "-archived" excludes notes matching "archived".
type NotExpr struct {
	// Expr is the expression to negate
	Expr Expr
}

func (NotExpr) exprNode() {}

// OrExpr represents an OR combination of expressions.
//
// Example: "(tag:work OR tag:personal)" matches either.
// Note: OR requires explicit parentheses to avoid ambiguity.
type OrExpr struct {
	// Left is the left operand
	Left Expr

	// Right is the right operand
	Right Expr
}

func (OrExpr) exprNode() {}

// DateExpr represents a date-based filter.
//
// Example: "created:>2024-01-01" matches notes created after Jan 1, 2024.
type DateExpr struct {
	// Field is the date field (created, modified)
	Field string

	// Op is the comparison operator
	Op CompareOp

	// Value is the date value (parsed from string)
	Value string

	// Note: The actual time.Time value is computed during query execution
	// since it may involve relative dates like "yesterday" or "this-week".
}

func (DateExpr) exprNode() {}

// RangeExpr represents a range query.
//
// Example: "created:2024-01..2024-06" matches notes in that range.
type RangeExpr struct {
	// Field is the field name
	Field string

	// Start is the range start (inclusive)
	Start string

	// End is the range end (inclusive)
	End string
}

func (RangeExpr) exprNode() {}

// WildcardExpr represents a prefix/suffix wildcard search.
//
// Example: "title:java*" matches titles starting with "java".
type WildcardExpr struct {
	// Field is the field name (empty for full-text)
	Field string

	// Pattern is the wildcard pattern (e.g., "java*")
	Pattern string

	// Type indicates the wildcard type
	Type WildcardType
}

func (WildcardExpr) exprNode() {}

// WildcardType specifies the type of wildcard match.
type WildcardType string

const (
	WildcardPrefix WildcardType = "prefix" // "java*"
	WildcardSuffix WildcardType = "suffix" // "*java"
	WildcardBoth   WildcardType = "both"   // "*java*"
)
