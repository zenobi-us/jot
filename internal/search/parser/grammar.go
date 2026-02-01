package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Grammar AST types for Participle parser.
// These are intermediate types that get converted to search.Expr types.

// queryAST is the root grammar node.
type queryAST struct {
	Expressions []*expressionAST `parser:"@@*"`
}

// expressionAST represents a single expression in the query.
type expressionAST struct {
	Not   *notExprAST   `parser:"  @@"`
	Field *fieldExprAST `parser:"| @@"`
	Term  *termAST      `parser:"| @@"`
}

// notExprAST represents a negated expression: -term or -field:value
type notExprAST struct {
	Field *fieldExprAST `parser:"'-' ( @@"`
	Term  *termAST      `parser:"    | @@ )"`
}

// fieldExprAST represents a field-qualified expression: field:value or field:>value
type fieldExprAST struct {
	Field    string `parser:"@Field ':'"`
	Operator string `parser:"@( '>''=' | '<''=' | '>' | '<' )?"`
	Value    string `parser:"( @String | @Date | @Word )"`
}

// termAST represents a simple search term.
type termAST struct {
	Value string `parser:"@String | @Word"`
}

// queryLexer defines the token types for the query language.
var queryLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Field", Pattern: `(tag|title|path|created|modified|body|status)`},
	{Name: "String", Pattern: `"[^"]*"`},
	// Date patterns must come before Word to capture dates properly
	{Name: "Date", Pattern: `\d{4}-\d{2}-\d{2}`},
	{Name: "Word", Pattern: `[^\s:"\-><>=]+`},
	{Name: "Punct", Pattern: `[:\-><>=]`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// queryParser is the Participle parser instance.
var queryParser = participle.MustBuild[queryAST](
	participle.Lexer(queryLexer),
	participle.Elide("Whitespace"),
	participle.UseLookahead(2),
)
