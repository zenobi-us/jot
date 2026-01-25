package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

// SpecialViewExecutor handles execution of views that require special logic beyond SQL
type SpecialViewExecutor struct {
	noteService *NoteService
	log         zerolog.Logger
}

// NewSpecialViewExecutor creates a new executor for special views
func NewSpecialViewExecutor(noteService *NoteService) *SpecialViewExecutor {
	return &SpecialViewExecutor{
		noteService: noteService,
		log:         Log("SpecialViewExecutor"),
	}
}

// ExecuteBrokenLinksView executes the broken-links view
// Finds notes containing links to non-existent files
func (sve *SpecialViewExecutor) ExecuteBrokenLinksView(ctx context.Context) ([]map[string]interface{}, error) {
	// Get all notes in the notebook
	notes, err := sve.noteService.getAllNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all notes: %w", err)
	}

	// Build a map of all existing note paths for quick lookup
	existingPaths := make(map[string]bool)
	for _, note := range notes {
		existingPaths[note.File.Relative] = true
	}

	// Track notes with broken links
	notesWithBrokenLinks := make([]map[string]interface{}, 0)

	for _, note := range notes {
		brokenLinks := sve.findBrokenLinks(&note, existingPaths)
		if len(brokenLinks) > 0 {
			result := map[string]interface{}{
				"file_path":     note.File.Filepath,
				"relative_path": note.File.Relative,
				"title":         note.DisplayName(),
				"broken_links":  brokenLinks,
				"link_count":    len(brokenLinks),
				"updated":       note.Metadata["updated"],
				"body":          strings.TrimSpace(note.Content),
			}
			notesWithBrokenLinks = append(notesWithBrokenLinks, result)
		}
	}

	// Sort by updated time (most recent first)
	// (Sorting would require importing sort package and custom comparator)

	return notesWithBrokenLinks, nil
}

// findBrokenLinks extracts all broken links from a note
// Checks both frontmatter links and markdown body links
func (sve *SpecialViewExecutor) findBrokenLinks(note *Note, existingPaths map[string]bool) []string {
	brokenLinks := make([]string, 0)
	foundLinks := make(map[string]bool) // Deduplicate

	// 1. Check frontmatter links (data.links array)
	if metadata, ok := note.Metadata["links"]; ok {
		if linksArray, ok := metadata.([]interface{}); ok {
			for _, linkItem := range linksArray {
				if linkStr, ok := linkItem.(string); ok {
					if !existingPaths[linkStr] && !foundLinks[linkStr] {
						brokenLinks = append(brokenLinks, linkStr)
						foundLinks[linkStr] = true
					}
				}
			}
		}
	}

	// 2. Check markdown body links: [text](path) syntax
	mdLinkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	for _, match := range mdLinkPattern.FindAllStringSubmatch(note.Content, -1) {
		if len(match) >= 3 {
			linkPath := strings.TrimSpace(match[2])
			// Skip external URLs
			if !strings.HasPrefix(linkPath, "http://") && !strings.HasPrefix(linkPath, "https://") && !strings.HasPrefix(linkPath, "#") {
				// Normalize path (remove .md if present, paths are usually without extension in frontmatter)
				normalizedPath := strings.TrimSuffix(linkPath, ".md")
				if !existingPaths[linkPath] && !existingPaths[normalizedPath] && !foundLinks[linkPath] {
					brokenLinks = append(brokenLinks, linkPath)
					foundLinks[linkPath] = true
				}
			}
		}
	}

	// 3. Check wiki-style links: [[wikilink]] syntax
	wikiLinkPattern := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	for _, match := range wikiLinkPattern.FindAllStringSubmatch(note.Content, -1) {
		if len(match) >= 2 {
			linkPath := strings.TrimSpace(match[1])
			// Add .md extension for wiki links if not present
			mdPath := linkPath
			if !strings.HasSuffix(mdPath, ".md") {
				mdPath = linkPath + ".md"
			}
			if !existingPaths[linkPath] && !existingPaths[mdPath] && !foundLinks[linkPath] {
				brokenLinks = append(brokenLinks, linkPath)
				foundLinks[linkPath] = true
			}
		}
	}

	return brokenLinks
}

// ExecuteOrphansView executes the orphans view
// Finds notes with no incoming links (isolated nodes in the knowledge graph)
func (sve *SpecialViewExecutor) ExecuteOrphansView(ctx context.Context, definitionType string) ([]map[string]interface{}, error) {
	// Get all notes in the notebook
	notes, err := sve.noteService.getAllNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all notes: %w", err)
	}

	// Build a map of all outgoing links from each note
	allOutgoingLinks := make(map[string]map[string]bool) // note path -> set of links
	for _, note := range notes {
		allOutgoingLinks[note.File.Relative] = sve.extractAllLinks(&note)
	}

	// Find orphan notes based on definition type
	orphans := make([]map[string]interface{}, 0)

	for _, note := range notes {
		isOrphan := false

		switch definitionType {
		case "no-incoming":
			// Node with no incoming links
			isOrphan = sve.hasNoIncomingLinks(&note, allOutgoingLinks)
		case "no-links":
			// Node with no links at all (incoming or outgoing)
			isOrphan = sve.hasNoLinksAtAll(&note, allOutgoingLinks)
		case "isolated":
			fallthrough
		default:
			// Isolated node: no links AND not tagged/categorized
			isOrphan = sve.isIsolatedNode(&note, allOutgoingLinks)
		}

		if isOrphan {
			result := map[string]interface{}{
				"file_path":     note.File.Filepath,
				"relative_path": note.File.Relative,
				"title":         note.DisplayName(),
				"created":       note.Metadata["created"],
				"updated":       note.Metadata["updated"],
				"tags":          note.Metadata["tags"],
				"body":          strings.TrimSpace(note.Content),
			}
			orphans = append(orphans, result)
		}
	}

	return orphans, nil
}

// extractAllLinks extracts all outgoing links from a note
func (sve *SpecialViewExecutor) extractAllLinks(note *Note) map[string]bool {
	links := make(map[string]bool)

	// 1. Frontmatter links
	if metadata, ok := note.Metadata["links"]; ok {
		if linksArray, ok := metadata.([]interface{}); ok {
			for _, linkItem := range linksArray {
				if linkStr, ok := linkItem.(string); ok {
					links[linkStr] = true
				}
			}
		}
	}

	// 2. Markdown body links
	mdLinkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	for _, match := range mdLinkPattern.FindAllStringSubmatch(note.Content, -1) {
		if len(match) >= 3 {
			linkPath := strings.TrimSpace(match[2])
			if !strings.HasPrefix(linkPath, "http://") && !strings.HasPrefix(linkPath, "https://") && !strings.HasPrefix(linkPath, "#") {
				links[linkPath] = true
				// Also add without .md extension
				links[strings.TrimSuffix(linkPath, ".md")] = true
			}
		}
	}

	// 3. Wiki links
	wikiLinkPattern := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	for _, match := range wikiLinkPattern.FindAllStringSubmatch(note.Content, -1) {
		if len(match) >= 2 {
			linkPath := strings.TrimSpace(match[1])
			links[linkPath] = true
			links[linkPath+".md"] = true
		}
	}

	return links
}

// hasNoIncomingLinks checks if a note has no incoming links (no other notes link to it)
func (sve *SpecialViewExecutor) hasNoIncomingLinks(targetNote *Note, allOutgoingLinks map[string]map[string]bool) bool {
	targetPath := targetNote.File.Relative
	targetPathMd := strings.TrimSuffix(targetPath, ".md")

	for sourcePath, outgoingLinks := range allOutgoingLinks {
		if sourcePath == targetPath {
			continue // Skip self-references
		}
		if outgoingLinks[targetPath] || outgoingLinks[targetPathMd] {
			return false // Found an incoming link
		}
	}

	return true
}

// hasNoLinksAtAll checks if a note has no links (incoming or outgoing)
func (sve *SpecialViewExecutor) hasNoLinksAtAll(note *Note, allOutgoingLinks map[string]map[string]bool) bool {
	// Check outgoing links
	if len(allOutgoingLinks[note.File.Relative]) > 0 {
		return false
	}

	// Check incoming links
	return sve.hasNoIncomingLinks(note, allOutgoingLinks)
}

// isIsolatedNode checks if a note is isolated: no links AND not tagged/categorized
func (sve *SpecialViewExecutor) isIsolatedNode(note *Note, allOutgoingLinks map[string]map[string]bool) bool {
	// Must have no links at all
	if !sve.hasNoLinksAtAll(note, allOutgoingLinks) {
		return false
	}

	// Must not be tagged or have categories
	if tags, ok := note.Metadata["tags"]; ok && tags != nil {
		if tagsList, ok := tags.([]interface{}); ok && len(tagsList) > 0 {
			return false
		}
	}

	if tag, ok := note.Metadata["tag"]; ok && tag != nil {
		if tagStr, ok := tag.(string); ok && tagStr != "" {
			return false
		}
	}

	// Check for category
	if category, ok := note.Metadata["category"]; ok && category != nil {
		if catStr, ok := category.(string); ok && catStr != "" {
			return false
		}
	}

	return true
}
