package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/zenobi-us/opennotes/internal/search"
	"github.com/zenobi-us/opennotes/internal/search/bleve"
	"gopkg.in/yaml.v3"
)

// NotebookGroup defines a group of notes with shared properties.
type NotebookGroup struct {
	Name     string         `json:"name"`
	Globs    []string       `json:"globs"`
	Metadata map[string]any `json:"metadata"`
	Template string         `json:"template,omitempty"`
}

// StoredNotebookConfig is what's stored in .opennotes.json.
type StoredNotebookConfig struct {
	Root      string            `json:"root"`
	Name      string            `json:"name"`
	Contexts  []string          `json:"contexts,omitempty"`
	Templates map[string]string `json:"templates,omitempty"`
	Groups    []NotebookGroup   `json:"groups,omitempty"`
}

// NotebookConfig includes runtime-resolved paths.
type NotebookConfig struct {
	StoredNotebookConfig
	Path string `json:"-"` // Path to the config file (not stored)
}

// Notebook represents a loaded notebook with its services.
type Notebook struct {
	Config NotebookConfig
	Notes  *NoteService
}

// NotebookService manages notebook operations.
type NotebookService struct {
	configService *ConfigService
	log           zerolog.Logger
}

// NewNotebookService creates a notebook service.
func NewNotebookService(cfg *ConfigService) *NotebookService {
	return &NotebookService{
		configService: cfg,
		log:           Log("NotebookService"),
	}
}

// configFilePath returns the config file path for a notebook directory.
func configFilePath(notebookDir string) string {
	return filepath.Join(notebookDir, NotebookConfigFile)
}

// HasNotebook checks if a directory contains a notebook.
func (s *NotebookService) HasNotebook(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(configFilePath(path))
	return err == nil
}

// LoadConfig loads notebook configuration from a directory.
func (s *NotebookService) LoadConfig(path string) (*NotebookConfig, error) {
	configPath := configFilePath(path)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var stored StoredNotebookConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("invalid notebook config: %w", err)
	}

	// Resolve root path relative to config location
	rootPath := filepath.Join(path, stored.Root)
	if _, err := os.Stat(rootPath); err != nil {
		// Create root directory if it doesn't exist
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(rootPath, 0755); mkErr != nil {
				return nil, fmt.Errorf("notes path not found and could not create: %s", rootPath)
			}
		} else {
			return nil, fmt.Errorf("notes path error: %w", err)
		}
	}

	return &NotebookConfig{
		StoredNotebookConfig: StoredNotebookConfig{
			Root:      rootPath, // Now absolute
			Name:      stored.Name,
			Contexts:  stored.Contexts,
			Templates: stored.Templates,
			Groups:    stored.Groups,
		},
		Path: configPath,
	}, nil
}

// Open loads a notebook from the given path.
func (s *NotebookService) Open(notebookPath string) (*Notebook, error) {
	config, err := s.LoadConfig(notebookPath)
	if err != nil {
		return nil, err
	}

	// Create Bleve index for this notebook
	idx, err := s.createIndex(config.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to create search index: %w", err)
	}

	noteService := NewNoteService(s.configService, idx, config.Root)

	semanticIdx, err := s.createSemanticIndex(config.Root)
	if err != nil {
		s.log.Warn().Err(err).Msg("failed to initialize semantic backend; using noop fallback")
		semanticIdx = NewNoopSemanticIndex()
	}
	noteService.SetSemanticIndex(semanticIdx)

	return &Notebook{
		Config: *config,
		Notes:  noteService,
	}, nil
}

// createIndex creates and populates a Bleve index for the notebook
func (s *NotebookService) createIndex(notebookRoot string) (search.Index, error) {
	// For now, use in-memory index
	// TODO: Consider persistent index for large notebooks
	storage := bleve.MemStorage()
	idx, err := bleve.NewIndex(storage, bleve.Options{InMemory: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// Index all markdown files in the notebook
	fs := afero.NewOsFs()
	err = afero.Walk(fs, notebookRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Get relative path from notebook root
		relPath, err := filepath.Rel(notebookRoot, path)
		if err != nil {
			s.log.Warn().Err(err).Str("path", path).Msg("failed to get relative path")
			return nil
		}

		// Read file content
		content, err := afero.ReadFile(fs, path)
		if err != nil {
			s.log.Warn().Err(err).Str("path", path).Msg("failed to read file")
			return nil
		}

		// Parse frontmatter and extract metadata
		metadata, body := parseFrontmatter(content)

		// Create document
		doc := search.Document{
			Path:     relPath,
			Title:    extractTitle(metadata),
			Body:     body,
			Lead:     extractLead(body),
			Tags:     extractTags(metadata),
			Metadata: metadata,
			Created:  extractTime(metadata, "created", info.ModTime()),
			Modified: extractTime(metadata, "modified", info.ModTime()),
		}

		// Add to index
		ctx := context.Background()
		if err := idx.Add(ctx, doc); err != nil {
			s.log.Warn().Err(err).Str("path", relPath).Msg("failed to index document")
		}

		return nil
	})

	if err != nil {
		_ = idx.Close()
		return nil, fmt.Errorf("failed to index notebook: %w", err)
	}

	return idx, nil
}

// createSemanticIndex initializes semantic retrieval backend for a notebook.
// Phase 3 starts with a safe noop backend and can be swapped with a real
// semantic backend implementation without changing callers.
func (s *NotebookService) createSemanticIndex(notebookRoot string) (SemanticIndex, error) {
	s.log.Debug().Str("notebookRoot", notebookRoot).Msg("semantic backend not configured; using noop fallback")
	return NewNoopSemanticIndex(), nil
}

// Helper functions for extracting metadata

func extractTitle(metadata map[string]any) string {
	if title, ok := metadata["title"].(string); ok && title != "" {
		return title
	}
	return ""
}

func extractTags(metadata map[string]any) []string {
	// Handle both "tag" and "tags" fields
	if tag, ok := metadata["tag"].(string); ok && tag != "" {
		return []string{tag}
	}
	if tags, ok := metadata["tags"].([]any); ok {
		result := make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if tags, ok := metadata["tags"].([]string); ok {
		return tags
	}
	return nil
}

func extractTime(metadata map[string]any, field string, defaultTime time.Time) time.Time {
	if t, ok := metadata[field].(time.Time); ok {
		return t
	}
	if s, ok := metadata[field].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			return parsed
		}
	}
	return defaultTime
}

func parseFrontmatter(content []byte) (map[string]any, string) {
	// Check for frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return make(map[string]any), string(content)
	}

	// Find the end of frontmatter
	rest := content[4:] // Skip first "---\n"
	endIdx := bytes.Index(rest, []byte("\n---\n"))
	if endIdx == -1 {
		// No closing delimiter, treat as no frontmatter
		return make(map[string]any), string(content)
	}

	// Extract frontmatter and body
	frontmatterBytes := rest[:endIdx]
	bodyBytes := rest[endIdx+5:] // Skip "\n---\n"

	// Parse YAML frontmatter
	var metadata map[string]any
	if err := yaml.Unmarshal(frontmatterBytes, &metadata); err != nil {
		// Failed to parse, return empty metadata
		return make(map[string]any), string(content)
	}

	return metadata, string(bodyBytes)
}

func extractLead(body string) string {
	lines := strings.Split(body, "\n")
	var lead strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines at the start
		if lead.Len() == 0 && line == "" {
			continue
		}

		// Skip headings
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Stop at first empty line after content
		if lead.Len() > 0 && line == "" {
			break
		}

		// Add line to lead
		if line != "" {
			if lead.Len() > 0 {
				lead.WriteString(" ")
			}
			lead.WriteString(line)
		}
	}

	result := lead.String()
	if len(result) > 200 {
		return result[:200] + "..."
	}
	return result
}

// Create creates a new notebook.
func (s *NotebookService) Create(name, path string, register bool) (*Notebook, error) {
	if path == "" {
		path, _ = os.Getwd()
	}

	// If the directory exists, use "." as root (use existing notes)
	// If it doesn't exist, create a ".notes" subdirectory for new notes
	var notesDir string
	if _, err := os.Stat(path); err == nil {
		// Directory exists - use it as the root
		notesDir = path
	} else {
		// Directory doesn't exist - create ".notes" subdirectory
		notesDir = filepath.Join(path, ".notes")
	}

	config := NotebookConfig{
		StoredNotebookConfig: StoredNotebookConfig{
			Root:     notesDir, // Store absolute path; SaveConfig will convert to relative
			Name:     name,
			Contexts: []string{path},
			Groups: []NotebookGroup{
				{
					Name:     "Default",
					Globs:    []string{"**/*.md"},
					Metadata: map[string]any{},
				},
			},
			Templates: map[string]string{},
		},
		Path: configFilePath(path),
	}

	// Create notes directory if it doesn't exist
	// (it already exists for existing notebook directories)
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, err
	}

	// Create notebook directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	// Create Bleve index for this notebook
	idx, err := s.createIndex(notesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create search index: %w", err)
	}

	noteService := NewNoteService(s.configService, idx, notesDir)

	semanticIdx, semErr := s.createSemanticIndex(notesDir)
	if semErr != nil {
		s.log.Warn().Err(semErr).Msg("failed to initialize semantic backend; using noop fallback")
		semanticIdx = NewNoopSemanticIndex()
	}
	noteService.SetSemanticIndex(semanticIdx)

	notebook := &Notebook{
		Config: config,
		Notes:  noteService,
	}

	// Save config
	if err := notebook.SaveConfig(register, s.configService); err != nil {
		return nil, err
	}

	return notebook, nil
}

// Infer discovers notebook from current context (auto-detection only).
// Note: OPENNOTES_NOTEBOOK envvar and --notebook flag are handled upstream in requireNotebook().
// Infer() handles auto-detection priority:
// 1. .opennotes.json in current directory
// 2. Context matching (registered notebooks with path context)
// 3. Ancestor search (walk up tree for .opennotes.json)
func (s *NotebookService) Infer(cwd string) (*Notebook, error) {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Step 1: Check .opennotes.json in current directory (direct check)
	if s.HasNotebook(cwd) {
		return s.Open(cwd)
	}

	// Step 2: Check registered notebooks for context match
	notebooks, _ := s.List(cwd)
	for _, nb := range notebooks {
		if nb.MatchContext(cwd) != "" {
			return nb, nil
		}
	}

	// Step 3: Search ancestor directories (start from parent, not current)
	current := filepath.Dir(cwd)
	for current != "/" && current != "" {
		if s.HasNotebook(current) {
			return s.Open(current)
		}
		current = filepath.Dir(current)
	}

	return nil, nil // No notebook found
}

// List returns all known notebooks.
func (s *NotebookService) List(cwd string) ([]*Notebook, error) {
	var notebooks []*Notebook

	// From registered paths
	for _, path := range s.configService.Store.Notebooks {
		if s.HasNotebook(path) {
			if nb, err := s.Open(path); err == nil {
				notebooks = append(notebooks, nb)
			}
		}
	}

	// From ancestor directories
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	current := cwd
	for current != "/" && current != "" {
		if s.HasNotebook(current) {
			if nb, err := s.Open(current); err == nil {
				// Avoid duplicates
				found := false
				for _, existing := range notebooks {
					if existing.Config.Path == nb.Config.Path {
						found = true
						break
					}
				}
				if !found {
					notebooks = append(notebooks, nb)
				}
			}
		}
		current = filepath.Dir(current)
	}

	return notebooks, nil
}

// Notebook methods

// MatchContext checks if a path matches any notebook context.
func (n *Notebook) MatchContext(path string) string {
	for _, ctx := range n.Config.Contexts {
		if strings.HasPrefix(path, ctx) {
			return ctx
		}
	}
	return ""
}

// AddContext adds a context path to the notebook.
func (n *Notebook) AddContext(contextPath string, configService *ConfigService) error {
	if contextPath == "" {
		contextPath, _ = os.Getwd()
	}

	// Check if already exists
	if slices.Contains(n.Config.Contexts, contextPath) {
		return nil // Already exists
	}

	n.Config.Contexts = append(n.Config.Contexts, contextPath)
	return n.SaveConfig(false, configService)
}

// SaveConfig writes the notebook config to disk.
func (n *Notebook) SaveConfig(register bool, configService *ConfigService) error {
	configDir := filepath.Dir(n.Config.Path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Calculate relative root for storage
	relRoot, _ := filepath.Rel(configDir, n.Config.Root)
	if relRoot == "" {
		relRoot = "."
	}

	stored := StoredNotebookConfig{
		Root:      relRoot,
		Name:      n.Config.Name,
		Contexts:  n.Config.Contexts,
		Templates: n.Config.Templates,
		Groups:    n.Config.Groups,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(n.Config.Path, data, 0644); err != nil {
		return err
	}

	// Register globally if requested
	if register {
		notebooks := configService.Store.Notebooks
		notebookDir := filepath.Dir(n.Config.Path)
		if slices.Contains(notebooks, notebookDir) {
			return nil // Already registered
		}
		configService.Store.Notebooks = append(notebooks, notebookDir)
		return configService.Write(configService.Store)
	}

	return nil
}

// GetViews returns all views defined in a notebook's .opennotes.json
// Returns an empty map if no views are defined
func (s *NotebookService) GetViews(notebookPath string) (map[string]json.RawMessage, error) {
	if notebookPath == "" {
		return map[string]json.RawMessage{}, nil
	}

	configPath := configFilePath(notebookPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, fmt.Errorf("failed to read notebook config: %w", err)
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse notebook config: %w", err)
	}

	views, ok := configData["views"].(map[string]interface{})
	if !ok {
		return map[string]json.RawMessage{}, nil
	}

	// Convert to json.RawMessage for consistency
	result := make(map[string]json.RawMessage)
	for name, viewData := range views {
		rawData, err := json.Marshal(viewData)
		if err != nil {
			s.log.Warn().Str("view", name).Err(err).Msg("failed to marshal view definition")
			continue
		}
		result[name] = rawData
	}

	return result, nil
}
