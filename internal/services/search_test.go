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
