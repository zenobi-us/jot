package services_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zenobi-us/opennotes/internal/services"
)

// Helper function to create test notes
func createTestNotes(count int) []services.Note {
	notes := make([]services.Note, count)
	for i := 0; i < count; i++ {
		notes[i] = services.Note{
			Content: "Test content",
			Metadata: map[string]any{
				"title": "Test Note",
			},
		}
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}
	return notes
}

func TestSearchService_FuzzySearch_BasicMatching(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "Team meeting notes",
			Metadata: map[string]any{
				"title": "Meeting Notes",
			},
		},
		{
			Content: "Daily standup",
			Metadata: map[string]any{
				"title": "Morning Standup",
			},
		},
		{
			Content: "Strategy discussion",
			Metadata: map[string]any{
				"title": "Project Planning",
			},
		},
	}

	// Set file paths
	for i := range notes {
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	// Test fuzzy matching for "meeting"
	results := svc.FuzzySearch("meeting", notes)

	// Should find at least the "Meeting Notes" note
	assert.GreaterOrEqual(t, len(results), 1)

	// First result should be the best match
	assert.Contains(t, results[0].DisplayName(), "Meeting")
}

func TestSearchService_FuzzySearch_Ranking(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "Some content",
			Metadata: map[string]any{
				"title": "project",
			},
		},
		{
			Content: "Some content",
			Metadata: map[string]any{
				"title": "big project ideas",
			},
		},
		{
			Content: "Some content",
			Metadata: map[string]any{
				"title": "project plan",
			},
		},
	}

	// Set file paths
	for i := range notes {
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	// Search for exact word "project"
	results := svc.FuzzySearch("project", notes)

	// All notes should match
	assert.Equal(t, 3, len(results))

	// Exact match should rank highest (first result should be just "project")
	assert.Equal(t, "project", results[0].DisplayName())
}

func TestSearchService_FuzzySearch_EmptyQuery(t *testing.T) {
	svc := services.NewSearchService()

	notes := createTestNotes(10)

	// Empty query returns all notes unsorted
	results := svc.FuzzySearch("", notes)

	assert.Equal(t, len(notes), len(results))
}

func TestSearchService_FuzzySearch_NoMatches(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "Apple content",
			Metadata: map[string]any{
				"title": "Apple Note",
			},
		},
		{
			Content: "Banana content",
			Metadata: map[string]any{
				"title": "Banana Note",
			},
		},
	}

	// Set file paths
	for i := range notes {
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	// Query that doesn't match anything
	results := svc.FuzzySearch("xyz123nonexistent", notes)

	assert.Equal(t, 0, len(results))
}

func TestSearchService_FuzzySearch_TitleVsBody(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "This note is about meetings and conferences",
			Metadata: map[string]any{
				"title": "Random Title",
			},
		},
		{
			Content: "This is just some random content",
			Metadata: map[string]any{
				"title": "Meeting Notes",
			},
		},
	}

	// Set file paths
	for i := range notes {
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	// Search for "meeting"
	results := svc.FuzzySearch("meeting", notes)

	// Both should match
	assert.Equal(t, 2, len(results))

	// Title match should rank higher (note with "Meeting Notes" title should be first)
	assert.Equal(t, "Meeting Notes", results[0].DisplayName())
}

func TestSearchService_FuzzySearch_EmptyNotes(t *testing.T) {
	svc := services.NewSearchService()

	var notes []services.Note

	results := svc.FuzzySearch("test", notes)

	assert.Nil(t, results)
}

func TestSearchService_FuzzySearch_LargeDataset(t *testing.T) {
	svc := services.NewSearchService()

	// Create 1000 notes
	notes := make([]services.Note, 1000)
	for i := 0; i < 1000; i++ {
		notes[i] = services.Note{
			Content: "Some content with various words and phrases",
			Metadata: map[string]any{
				"title": "Test Note " + string(rune(i)),
			},
		}
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	// Add a few notes with "meeting" keyword
	notes[100].Metadata["title"] = "Meeting Notes"
	notes[200].Content = "This is about a team meeting"
	notes[300].Metadata["title"] = "Conference Meeting"

	// Search should complete quickly
	results := svc.FuzzySearch("meeting", notes)

	// Should find at least the 3 notes we added
	assert.GreaterOrEqual(t, len(results), 3)
}

func TestSearchService_FuzzySearch_LongBodyContent(t *testing.T) {
	svc := services.NewSearchService()

	// Create note with very long body (> 500 chars)
	longContent := strings.Repeat("Some text about various topics. ", 50) + "meeting keyword here"

	notes := []services.Note{
		{
			Content: longContent,
			Metadata: map[string]any{
				"title": "Long Document",
			},
		},
	}
	notes[0].File.Filepath = "/test/note.md"
	notes[0].File.Relative = "note.md"

	// Should still work but only search first 500 chars
	results := svc.FuzzySearch("meeting", notes)

	// May or may not match depending on where "meeting" is in the 500-char window
	// Just ensure it doesn't crash
	assert.NotNil(t, results)
}

func TestSearchService_TextSearch_ExactMatch(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "This is about apples",
			Metadata: map[string]any{
				"title": "Apple Note",
			},
		},
		{
			Content: "This is about bananas",
			Metadata: map[string]any{
				"title": "Banana Note",
			},
		},
		{
			Content: "This is about cherries",
			Metadata: map[string]any{
				"title": "Cherry Note",
			},
		},
	}

	// Set file paths
	for i := range notes {
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	results := svc.TextSearch("apple", notes)

	assert.Equal(t, 1, len(results))
	assert.Contains(t, results[0].Content, "apples")
}

func TestSearchService_TextSearch_CaseInsensitive(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "This is about UPPERCASE content",
			Metadata: map[string]any{
				"title": "Test Note",
			},
		},
	}
	notes[0].File.Filepath = "/test/note.md"
	notes[0].File.Relative = "note.md"

	results := svc.TextSearch("uppercase", notes)

	assert.Equal(t, 1, len(results))
}

func TestSearchService_TextSearch_SearchInFilepath(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "Some content",
			Metadata: map[string]any{
				"title": "Test Note",
			},
		},
	}
	notes[0].File.Filepath = "/test/projects/myproject/note.md"
	notes[0].File.Relative = "note.md"

	results := svc.TextSearch("myproject", notes)

	assert.Equal(t, 1, len(results))
}

func TestSearchService_TextSearch_EmptyQuery(t *testing.T) {
	svc := services.NewSearchService()

	notes := createTestNotes(5)

	results := svc.TextSearch("", notes)

	// Empty query returns all notes
	assert.Equal(t, len(notes), len(results))
}

func TestSearchService_TextSearch_NoMatches(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: "Apple content",
			Metadata: map[string]any{
				"title": "Apple",
			},
		},
	}
	notes[0].File.Filepath = "/test/note.md"
	notes[0].File.Relative = "note.md"

	results := svc.TextSearch("banana", notes)

	assert.Equal(t, 0, len(results))
}

func TestSearchService_FuzzySearch_TitleWeighting(t *testing.T) {
	svc := services.NewSearchService()

	notes := []services.Note{
		{
			Content: strings.Repeat("project ", 100), // Many matches in body
			Metadata: map[string]any{
				"title": "Other Document",
			},
		},
		{
			Content: "Brief content",
			Metadata: map[string]any{
				"title": "project", // Single match in title
			},
		},
	}

	// Set file paths
	for i := range notes {
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	results := svc.FuzzySearch("project", notes)

	// Title match should rank higher even with fewer total occurrences
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "project", results[0].DisplayName())
}

// Benchmark tests
func BenchmarkSearchService_FuzzySearch_10kNotes(b *testing.B) {
	svc := services.NewSearchService()

	// Create 10,000 test notes
	notes := make([]services.Note, 10000)
	for i := 0; i < 10000; i++ {
		notes[i] = services.Note{
			Content: "Some content with various words and test phrases for searching",
			Metadata: map[string]any{
				"title": "Test Note " + strings.Repeat("x", i%100),
			},
		}
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	// Add some notes with "meeting" keyword
	for i := 0; i < 100; i++ {
		notes[i*100].Metadata["title"] = "Meeting Notes " + strings.Repeat("x", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := svc.FuzzySearch("meeting", notes)
		_ = results // Prevent optimization
	}
}

func BenchmarkSearchService_TextSearch_10kNotes(b *testing.B) {
	svc := services.NewSearchService()

	// Create 10,000 test notes
	notes := make([]services.Note, 10000)
	for i := 0; i < 10000; i++ {
		notes[i] = services.Note{
			Content: "Some content with various words and test phrases for searching",
			Metadata: map[string]any{
				"title": "Test Note " + strings.Repeat("x", i%100),
			},
		}
		notes[i].File.Filepath = "/test/note.md"
		notes[i].File.Relative = "note.md"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := svc.TextSearch("test", notes)
		_ = results // Prevent optimization
	}
}

// ============================================================================
// Boolean Query Tests - ParseConditions
// ============================================================================

func TestSearchService_ParseConditions_ValidSingleAnd(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{"data.tag=workflow"},
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "and", conditions[0].Type)
	assert.Equal(t, "data.tag", conditions[0].Field)
	assert.Equal(t, "=", conditions[0].Operator)
	assert.Equal(t, "workflow", conditions[0].Value)
}

func TestSearchService_ParseConditions_ValidMultipleAnd(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{"data.tag=workflow", "data.status=active"},
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 2)
	assert.Equal(t, "and", conditions[0].Type)
	assert.Equal(t, "and", conditions[1].Type)
	assert.Equal(t, "data.tag", conditions[0].Field)
	assert.Equal(t, "data.status", conditions[1].Field)
}

func TestSearchService_ParseConditions_ValidOrConditions(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{},
		[]string{"data.priority=high", "data.priority=critical"},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 2)
	assert.Equal(t, "or", conditions[0].Type)
	assert.Equal(t, "or", conditions[1].Type)
}

func TestSearchService_ParseConditions_ValidNotCondition(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{},
		[]string{},
		[]string{"data.status=archived"},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "not", conditions[0].Type)
}

func TestSearchService_ParseConditions_MixedConditions(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{"data.tag=epic"},
		[]string{"data.priority=high"},
		[]string{"data.status=archived"},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 3)

	// Verify order: and, or, not
	assert.Equal(t, "and", conditions[0].Type)
	assert.Equal(t, "or", conditions[1].Type)
	assert.Equal(t, "not", conditions[2].Type)
}

func TestSearchService_ParseConditions_AllValidFields(t *testing.T) {
	svc := services.NewSearchService()

	validFields := []string{
		"data.tag", "data.tags", "data.status", "data.priority",
		"data.assignee", "data.author", "data.type", "data.category",
		"data.project", "data.sprint", "path", "title",
		"links-to", "linked-by",
	}

	for _, field := range validFields {
		conditions, err := svc.ParseConditions(
			[]string{field + "=testvalue"},
			[]string{},
			[]string{},
		)
		assert.NoError(t, err, "field %s should be valid", field)
		assert.Len(t, conditions, 1)
		assert.Equal(t, field, conditions[0].Field)
	}
}

// ============================================================================
// Boolean Query Tests - Security: Invalid Field (Whitelist)
// ============================================================================

func TestSearchService_ParseConditions_InvalidField(t *testing.T) {
	svc := services.NewSearchService()

	invalidFields := []string{
		"content",           // not allowed - could be SQL injection vector
		"data.password",     // not allowed - not in whitelist
		"file_path",         // not allowed - use "path" instead
		"metadata",          // not allowed - too broad
		"SELECT",            // SQL keyword
		"DROP",              // SQL keyword
		"; DROP TABLE",      // SQL injection attempt
		"data.tag; DELETE",  // SQL injection attempt
	}

	for _, field := range invalidFields {
		_, err := svc.ParseConditions(
			[]string{field + "=value"},
			[]string{},
			[]string{},
		)
		assert.Error(t, err, "field %s should be invalid", field)
		assert.Contains(t, err.Error(), "invalid field")
	}
}

// ============================================================================
// Boolean Query Tests - Security: Value Validation
// ============================================================================

func TestSearchService_ParseConditions_ValueTooLong(t *testing.T) {
	svc := services.NewSearchService()

	// Create a value that exceeds MaxValueLength (1000)
	longValue := strings.Repeat("a", 1001)

	_, err := svc.ParseConditions(
		[]string{"data.tag=" + longValue},
		[]string{},
		[]string{},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestSearchService_ParseConditions_ValueAtMaxLength(t *testing.T) {
	svc := services.NewSearchService()

	// Create a value at exactly MaxValueLength (1000)
	maxValue := strings.Repeat("a", 1000)

	conditions, err := svc.ParseConditions(
		[]string{"data.tag=" + maxValue},
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, maxValue, conditions[0].Value)
}

func TestSearchService_ParseConditions_EmptyValue(t *testing.T) {
	svc := services.NewSearchService()

	_, err := svc.ParseConditions(
		[]string{"data.tag="},
		[]string{},
		[]string{},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// ============================================================================
// Boolean Query Tests - Format Validation
// ============================================================================

func TestSearchService_ParseConditions_InvalidFormat_NoEquals(t *testing.T) {
	svc := services.NewSearchService()

	_, err := svc.ParseConditions(
		[]string{"data.tag-workflow"},
		[]string{},
		[]string{},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected field=value")
}

func TestSearchService_ParseConditions_InvalidFormat_MultipleEquals(t *testing.T) {
	svc := services.NewSearchService()

	// Multiple equals should be handled - everything after first = is the value
	conditions, err := svc.ParseConditions(
		[]string{"data.tag=value=with=equals"},
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "value=with=equals", conditions[0].Value)
}

func TestSearchService_ParseConditions_WhitespaceHandling(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{" data.tag = workflow "},
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "data.tag", conditions[0].Field)
	assert.Equal(t, "workflow", conditions[0].Value)
}

// ============================================================================
// Boolean Query Tests - BuildWhereClause
// ============================================================================

func TestSearchService_BuildWhereClause_EmptyConditions(t *testing.T) {
	svc := services.NewSearchService()

	whereClause, params, err := svc.BuildWhereClause([]services.QueryCondition{})

	assert.NoError(t, err)
	assert.Equal(t, "", whereClause)
	assert.Len(t, params, 0)
}

func TestSearchService_BuildWhereClause_SingleAnd(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	assert.NotEmpty(t, whereClause)
	// Should use parameterized query with ? placeholder
	assert.Contains(t, whereClause, "?")
	// Params should contain the field name and value
	assert.Contains(t, params, "tag")
	assert.Contains(t, params, "workflow")
}

func TestSearchService_BuildWhereClause_MultipleAndConditions(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
		{Type: "and", Field: "data.status", Operator: "=", Value: "active"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Should have AND between conditions
	assert.Contains(t, whereClause, " AND ")
	// Should have 4 params (2 field names + 2 values)
	assert.Len(t, params, 4)
}

func TestSearchService_BuildWhereClause_OrConditions(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Should have OR between conditions in parentheses
	assert.Contains(t, whereClause, " OR ")
	assert.Contains(t, whereClause, "(")
	assert.Contains(t, whereClause, ")")
	assert.Len(t, params, 4)
}

func TestSearchService_BuildWhereClause_NotCondition(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "not", Field: "data.status", Operator: "=", Value: "archived"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Should have NOT keyword
	assert.Contains(t, whereClause, "NOT")
	assert.Len(t, params, 2)
}

func TestSearchService_BuildWhereClause_PathField(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "projects/*"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Should use LIKE for path with glob pattern converted
	assert.Contains(t, whereClause, "LIKE")
	assert.Contains(t, whereClause, "file_path")
	// Glob * should be converted to %
	assert.Contains(t, params, "projects/%")
}

func TestSearchService_BuildWhereClause_TitleField(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "title", Operator: "=", Value: "Meeting"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Should check both metadata title and filename
	assert.Contains(t, whereClause, "metadata")
	assert.Contains(t, whereClause, "file_path")
	assert.Contains(t, params, "Meeting")
}

func TestSearchService_BuildWhereClause_LinksToField(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "links-to", Operator: "=", Value: "target-note.md"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Should use EXISTS with subquery for links array
	assert.Contains(t, whereClause, "EXISTS")
	assert.Contains(t, params, "target-note.md")
}

// ============================================================================
// Boolean Query Tests - Security: SQL Injection Prevention
// ============================================================================

func TestSearchService_BuildWhereClause_SQLInjection_ValueWithQuotes(t *testing.T) {
	svc := services.NewSearchService()

	// Attempt SQL injection via value
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "'; DROP TABLE notes; --"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// The malicious value should be in params, NOT in the SQL string
	assert.NotContains(t, whereClause, "DROP")
	assert.NotContains(t, whereClause, "';")
	// The value should be passed as a parameter
	assert.Contains(t, params, "'; DROP TABLE notes; --")
}

func TestSearchService_BuildWhereClause_SQLInjection_ValueWithDash(t *testing.T) {
	svc := services.NewSearchService()

	// Attempt comment injection
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "test -- comment"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Comment characters should be in params, not affecting SQL
	assert.NotContains(t, whereClause, "--")
	assert.Contains(t, params, "test -- comment")
}

func TestSearchService_BuildWhereClause_SQLInjection_ValueWithSemicolon(t *testing.T) {
	svc := services.NewSearchService()

	// Attempt statement termination injection
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "test; DELETE FROM notes"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Semicolon and DELETE should be in params, not SQL
	assert.NotContains(t, whereClause, "DELETE")
	assert.NotContains(t, whereClause, ";")
	assert.Contains(t, params, "test; DELETE FROM notes")
}

func TestSearchService_BuildWhereClause_SQLInjection_UnicodeAttack(t *testing.T) {
	svc := services.NewSearchService()

	// Unicode characters that could be problematic
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "test\x00null\x1fcontrol"},
	}

	_, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Unicode should be safely parameterized
	assert.Contains(t, params, "test\x00null\x1fcontrol")
}

func TestSearchService_BuildWhereClause_Parameterized_CountPlaceholders(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "one"},
		{Type: "and", Field: "data.status", Operator: "=", Value: "two"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "three"},
	}

	whereClause, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// Count ? placeholders should match params count
	placeholderCount := strings.Count(whereClause, "?")
	assert.Equal(t, len(params), placeholderCount)
}

// ============================================================================
// Boolean Query Tests - Glob Pattern Conversion
// ============================================================================

func TestSearchService_BuildWhereClause_GlobConversion_Star(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "*.md"},
	}

	_, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// * should convert to %
	assert.Contains(t, params, "%.md")
}

func TestSearchService_BuildWhereClause_GlobConversion_DoubleStar(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "**/*.md"},
	}

	_, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// ** should convert to %
	assert.Contains(t, params, "%/%.md")
}

func TestSearchService_BuildWhereClause_GlobConversion_QuestionMark(t *testing.T) {
	svc := services.NewSearchService()

	conditions := []services.QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "note?.md"},
	}

	_, params, err := svc.BuildWhereClause(conditions)

	assert.NoError(t, err)
	// ? should convert to _
	assert.Contains(t, params, "note_.md")
}

// ============================================================================
// Boolean Query Tests - Empty Conditions
// ============================================================================

func TestSearchService_ParseConditions_AllEmpty(t *testing.T) {
	svc := services.NewSearchService()

	conditions, err := svc.ParseConditions(
		[]string{},
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.Len(t, conditions, 0)
}
