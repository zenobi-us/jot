package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/search"
)

// TestSearchService_BuildQuery_SingleTag tests a single tag condition.
func TestSearchService_BuildQuery_SingleTag(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.NotNil(t, query)
	require.Len(t, query.Expressions, 1)

	// Check expression type
	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok, "expected FieldExpr")
	assert.Equal(t, "metadata.tag", fieldExpr.Field)
	assert.Equal(t, search.OpEquals, fieldExpr.Op)
	assert.Equal(t, "work", fieldExpr.Value)
}

// TestSearchService_BuildQuery_MultipleAnd tests multiple AND conditions.
func TestSearchService_BuildQuery_MultipleAnd(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
		{Type: "and", Field: "data.status", Operator: "=", Value: "active"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.Len(t, query.Expressions, 2)

	// Both should be FieldExpr
	for i, expr := range query.Expressions {
		_, ok := expr.(search.FieldExpr)
		assert.True(t, ok, "expected FieldExpr for AND condition %d", i)
	}
}

// TestSearchService_BuildQuery_MultipleOr tests multiple OR conditions.
func TestSearchService_BuildQuery_MultipleOr(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.Len(t, query.Expressions, 1)

	// Should be nested OrExpr
	orExpr, ok := query.Expressions[0].(search.OrExpr)
	require.True(t, ok, "expected OrExpr for OR conditions")

	// Check left and right
	leftField, ok := orExpr.Left.(search.FieldExpr)
	assert.True(t, ok)
	assert.Equal(t, "high", leftField.Value)

	rightField, ok := orExpr.Right.(search.FieldExpr)
	assert.True(t, ok)
	assert.Equal(t, "critical", rightField.Value)
}

// TestSearchService_BuildQuery_SingleOr tests a single OR condition.
func TestSearchService_BuildQuery_SingleOr(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "or", Field: "data.tag", Operator: "=", Value: "work"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.Len(t, query.Expressions, 1)

	// Single OR should be just a FieldExpr
	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok, "expected FieldExpr for single OR")
	assert.Equal(t, "metadata.tag", fieldExpr.Field)
	assert.Equal(t, "work", fieldExpr.Value)
}

// TestSearchService_BuildQuery_Not tests a NOT condition.
func TestSearchService_BuildQuery_Not(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "not", Field: "data.status", Operator: "=", Value: "archived"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.Len(t, query.Expressions, 1)

	// Should be NotExpr
	notExpr, ok := query.Expressions[0].(search.NotExpr)
	require.True(t, ok, "expected NotExpr")

	// Inner should be FieldExpr
	fieldExpr, ok := notExpr.Expr.(search.FieldExpr)
	require.True(t, ok)
	assert.Equal(t, "metadata.status", fieldExpr.Field)
	assert.Equal(t, "archived", fieldExpr.Value)
}

// TestSearchService_BuildQuery_PathPrefix tests path prefix optimization.
func TestSearchService_BuildQuery_PathPrefix(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "projects/*"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)

	// Should be FieldExpr with OpPrefix
	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok)
	assert.Equal(t, "path", fieldExpr.Field)
	assert.Equal(t, search.OpPrefix, fieldExpr.Op)
	assert.Equal(t, "projects/", fieldExpr.Value)
}

// TestSearchService_BuildQuery_PathWithTrailingSlash tests path with trailing slash.
func TestSearchService_BuildQuery_PathWithTrailingSlash(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "projects/"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)

	// Should be FieldExpr with OpPrefix
	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok)
	assert.Equal(t, "path", fieldExpr.Field)
	assert.Equal(t, search.OpPrefix, fieldExpr.Op)
	assert.Equal(t, "projects/", fieldExpr.Value)
}

// TestSearchService_BuildQuery_PathWildcard tests complex wildcard patterns.
func TestSearchService_BuildQuery_PathWildcard(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "**/tasks/*.md"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)

	// Should be WildcardExpr
	wildcardExpr, ok := query.Expressions[0].(search.WildcardExpr)
	require.True(t, ok)
	assert.Equal(t, "path", wildcardExpr.Field)
	assert.Equal(t, "**/tasks/*.md", wildcardExpr.Pattern)
}

// TestSearchService_BuildQuery_PathExact tests exact path match.
func TestSearchService_BuildQuery_PathExact(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "projects/epic1.md"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)

	// Should be FieldExpr with OpEquals
	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok)
	assert.Equal(t, "path", fieldExpr.Field)
	assert.Equal(t, search.OpEquals, fieldExpr.Op)
	assert.Equal(t, "projects/epic1.md", fieldExpr.Value)
}

// TestSearchService_BuildQuery_TitleField tests title field.
func TestSearchService_BuildQuery_TitleField(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "title", Operator: "=", Value: "Meeting Notes"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)

	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok)
	assert.Equal(t, "title", fieldExpr.Field)
	assert.Equal(t, search.OpEquals, fieldExpr.Op)
	assert.Equal(t, "Meeting Notes", fieldExpr.Value)
}

// TestSearchService_BuildQuery_EmptyConditions tests empty conditions.
func TestSearchService_BuildQuery_EmptyConditions(t *testing.T) {
	searchSvc := NewSearchService()

	query, err := searchSvc.BuildQuery(context.Background(), []QueryCondition{})
	require.NoError(t, err)
	assert.NotNil(t, query)
	assert.Len(t, query.Expressions, 0)
}

// TestSearchService_BuildQuery_LinksToError tests links-to returns error.
func TestSearchService_BuildQuery_LinksToError(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "links-to", Operator: "=", Value: "docs/*.md"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	assert.Error(t, err)
	assert.Nil(t, query)
	assert.Contains(t, err.Error(), "link queries are not yet supported")
	assert.Contains(t, err.Error(), "Phase 5.3")
}

// TestSearchService_BuildQuery_LinkedByError tests linked-by returns error.
func TestSearchService_BuildQuery_LinkedByError(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "linked-by", Operator: "=", Value: "plan.md"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	assert.Error(t, err)
	assert.Nil(t, query)
	assert.Contains(t, err.Error(), "link queries are not yet supported")
	assert.Contains(t, err.Error(), "Phase 5.3")
}

// TestSearchService_BuildQuery_UnknownField tests error for unknown fields.
func TestSearchService_BuildQuery_UnknownField(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "unknown-field", Operator: "=", Value: "value"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	assert.Error(t, err)
	assert.Nil(t, query)
	assert.Contains(t, err.Error(), "unsupported field")
}

// TestSearchService_BuildQuery_MixedConditions tests mixed AND/OR/NOT.
func TestSearchService_BuildQuery_MixedConditions(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
		{Type: "not", Field: "data.status", Operator: "=", Value: "done"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.Len(t, query.Expressions, 3)

	// Should have FieldExpr (AND), FieldExpr (single OR), NotExpr (NOT)
	_, hasField := query.Expressions[0].(search.FieldExpr)
	assert.True(t, hasField, "first should be FieldExpr (AND)")

	_, hasOr := query.Expressions[1].(search.FieldExpr)
	assert.True(t, hasOr, "second should be FieldExpr (single OR)")

	_, hasNot := query.Expressions[2].(search.NotExpr)
	assert.True(t, hasNot, "third should be NotExpr")
}

// TestSearchService_BuildQuery_TagsAlias tests data.tags is aliased to data.tag.
func TestSearchService_BuildQuery_TagsAlias(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "and", Field: "data.tags", Operator: "=", Value: "work"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	require.NoError(t, err)
	require.Len(t, query.Expressions, 1)

	// Should map data.tags -> metadata.tag
	fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
	require.True(t, ok)
	assert.Equal(t, "metadata.tag", fieldExpr.Field)
	assert.Equal(t, "work", fieldExpr.Value)
}

// TestSearchService_BuildQuery_AllMetadataFields tests all supported metadata fields.
func TestSearchService_BuildQuery_AllMetadataFields(t *testing.T) {
	searchSvc := NewSearchService()

	fields := []string{
		"data.tag", "data.status", "data.priority", "data.assignee",
		"data.author", "data.type", "data.category", "data.project", "data.sprint",
	}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			conditions := []QueryCondition{
				{Type: "and", Field: field, Operator: "=", Value: "test"},
			}

			query, err := searchSvc.BuildQuery(context.Background(), conditions)
			require.NoError(t, err)
			require.Len(t, query.Expressions, 1)

			fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
			require.True(t, ok)

			expectedField := "metadata." + field[5:] // Strip "data." prefix
			assert.Equal(t, expectedField, fieldExpr.Field)
			assert.Equal(t, "test", fieldExpr.Value)
		})
	}
}

// TestSearchService_BuildQuery_InvalidConditionType tests invalid condition type.
func TestSearchService_BuildQuery_InvalidConditionType(t *testing.T) {
	searchSvc := NewSearchService()

	conditions := []QueryCondition{
		{Type: "invalid", Field: "data.tag", Operator: "=", Value: "work"},
	}

	query, err := searchSvc.BuildQuery(context.Background(), conditions)
	assert.Error(t, err)
	assert.Nil(t, query)
	assert.Contains(t, err.Error(), "unsupported condition type")
}
