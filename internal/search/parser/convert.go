package parser

import (
	"strings"

	"github.com/zenobi-us/opennotes/internal/search"
)

// convert transforms the Participle AST into search.Query.
func convert(ast *queryAST) *search.Query {
	if ast == nil || len(ast.Expressions) == 0 {
		return &search.Query{}
	}

	exprs := make([]search.Expr, 0, len(ast.Expressions))
	for _, e := range ast.Expressions {
		if expr := convertExpr(e); expr != nil {
			exprs = append(exprs, expr)
		}
	}

	return &search.Query{
		Expressions: exprs,
	}
}

// convertExpr converts a single expression AST node.
func convertExpr(e *expressionAST) search.Expr {
	if e == nil {
		return nil
	}

	switch {
	case e.Not != nil:
		return convertNot(e.Not)
	case e.Field != nil:
		return convertField(e.Field)
	case e.Term != nil:
		return convertTerm(e.Term)
	default:
		return nil
	}
}

// convertNot converts a negation expression.
func convertNot(n *notExprAST) search.Expr {
	if n == nil {
		return nil
	}

	var inner search.Expr
	switch {
	case n.Field != nil:
		inner = convertField(n.Field)
	case n.Term != nil:
		inner = convertTerm(n.Term)
	default:
		return nil
	}

	return search.NotExpr{Expr: inner}
}

// convertField converts a field expression.
func convertField(f *fieldExprAST) search.Expr {
	if f == nil {
		return nil
	}

	value := unquote(f.Value)
	field := strings.ToLower(f.Field)
	op := normalizeOp(f.Operator)

	// Date fields get special handling
	if field == "created" || field == "modified" {
		return search.DateExpr{
			Field: field,
			Op:    op,
			Value: value,
		}
	}

	// Regular field expression
	return search.FieldExpr{
		Field: field,
		Op:    op,
		Value: value,
	}
}

// convertTerm converts a simple term.
func convertTerm(t *termAST) search.Expr {
	if t == nil {
		return nil
	}

	return search.TermExpr{
		Value: unquote(t.Value),
	}
}

// unquote removes surrounding quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// normalizeOp converts operator string to CompareOp.
func normalizeOp(op string) search.CompareOp {
	switch op {
	case ">":
		return search.OpGt
	case ">=":
		return search.OpGte
	case "<":
		return search.OpLt
	case "<=":
		return search.OpLte
	default:
		return search.OpEquals
	}
}
