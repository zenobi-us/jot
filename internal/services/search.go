package services

import (
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/sahilm/fuzzy"
)

// SearchService provides search operations for notes.
type SearchService struct {
	log zerolog.Logger
}

// NewSearchService creates a new search service.
func NewSearchService() *SearchService {
	return &SearchService{
		log: Log("SearchService"),
	}
}

// fuzzyMatch represents a note with its fuzzy match score.
type fuzzyMatch struct {
	note  Note
	score int
}

// FuzzySearch performs fuzzy matching on notes and returns them ranked by score.
// If query is empty, returns all notes unsorted.
// Title matches are weighted 2x higher than body matches.
func (s *SearchService) FuzzySearch(query string, notes []Note) []Note {
	if len(notes) == 0 {
		return nil
	}

	// Empty query - return all notes
	if query == "" {
		return notes
	}

	var matches []fuzzyMatch

	for _, note := range notes {
		titleScore := 0
		bodyScore := 0

		// Try fuzzy matching on title
		title := note.DisplayName()
		if title != "" {
			titleMatches := fuzzy.Find(query, []string{title})
			if len(titleMatches) > 0 {
				// Title matches are weighted 2x higher
				titleScore = titleMatches[0].Score * 2
			}
		}

		// Try fuzzy matching on body preview (first 500 chars for performance)
		bodyPreview := note.Content
		if len(bodyPreview) > 500 {
			bodyPreview = bodyPreview[:500]
		}
		if bodyPreview != "" {
			bodyMatches := fuzzy.Find(query, []string{bodyPreview})
			if len(bodyMatches) > 0 {
				bodyScore = bodyMatches[0].Score
			}
		}

		// Take the best score
		score := titleScore
		if bodyScore > score {
			score = bodyScore
		}

		// Only include if there's a match
		if score > 0 {
			matches = append(matches, fuzzyMatch{
				note:  note,
				score: score,
			})
		}
	}

	// Sort by score descending (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	// Extract sorted notes
	result := make([]Note, len(matches))
	for i, match := range matches {
		result[i] = match.note
	}

	s.log.Debug().
		Str("query", query).
		Int("total_notes", len(notes)).
		Int("matches", len(result)).
		Msg("fuzzy search completed")

	return result
}

// TextSearch performs exact text matching on notes.
// Searches both content and filepath (case-insensitive).
func (s *SearchService) TextSearch(query string, notes []Note) []Note {
	if query == "" {
		return notes
	}

	var matches []Note
	queryLower := strings.ToLower(query)

	for _, note := range notes {
		// Check content
		if strings.Contains(strings.ToLower(note.Content), queryLower) {
			matches = append(matches, note)
			continue
		}

		// Check filepath
		if strings.Contains(strings.ToLower(note.File.Filepath), queryLower) {
			matches = append(matches, note)
			continue
		}
	}

	s.log.Debug().
		Str("query", query).
		Int("total_notes", len(notes)).
		Int("matches", len(matches)).
		Msg("text search completed")

	return matches
}
