package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// NotebookConfig represents a notebook configuration file.
type NotebookConfig struct {
	Name     string   `json:"name"`
	Root     string   `json:"root"`
	Contexts []string `json:"contexts,omitempty"`
}

// CreateTestNotebook creates a notebook directory structure for testing.
// Returns the path to the notebook root directory.
func CreateTestNotebook(t *testing.T, dir, name string) string {
	t.Helper()

	notebookDir := filepath.Join(dir, name)

	// Create notebook directory
	if err := os.MkdirAll(notebookDir, 0755); err != nil {
		t.Fatalf("failed to create notebook directory: %v", err)
	}

	// Create notes directory
	notesDir := filepath.Join(notebookDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("failed to create notes directory: %v", err)
	}

	// Create .opennotes.json config
	config := NotebookConfig{
		Name:     name,
		Root:     notebookDir,
		Contexts: []string{},
	}

	configPath := filepath.Join(notebookDir, ".opennotes.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal notebook config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write notebook config: %v", err)
	}

	return notebookDir
}

// CreateTestNote creates a test markdown note file.
// Returns the path to the note file.
func CreateTestNote(t *testing.T, notebookDir, filename, content string) string {
	t.Helper()

	notesDir := filepath.Join(notebookDir, "notes")
	notePath := filepath.Join(notesDir, filename)

	// Ensure notes directory exists
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("failed to create notes directory: %v", err)
	}

	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write note file: %v", err)
	}

	return notePath
}

// CreateTestNoteWithFrontmatter creates a note with YAML frontmatter.
func CreateTestNoteWithFrontmatter(t *testing.T, notebookDir, filename string, frontmatter map[string]interface{}, body string) string {
	t.Helper()

	// Build frontmatter YAML
	var content string
	if len(frontmatter) > 0 {
		content = "---\n"
		for key, value := range frontmatter {
			content += key + ": " + formatValue(value) + "\n"
		}
		content += "---\n\n"
	}
	content += body

	return CreateTestNote(t, notebookDir, filename, content)
}

// formatValue formats a value for YAML output.
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []string:
		result := "["
		for i, s := range val {
			if i > 0 {
				result += ", "
			}
			result += s
		}
		return result + "]"
	default:
		data, _ := json.Marshal(val)
		return string(data)
	}
}

// CreateInvalidNotebookConfig creates a notebook with invalid JSON config.
func CreateInvalidNotebookConfig(t *testing.T, dir, name string) string {
	t.Helper()

	notebookDir := filepath.Join(dir, name)

	if err := os.MkdirAll(notebookDir, 0755); err != nil {
		t.Fatalf("failed to create notebook directory: %v", err)
	}

	configPath := filepath.Join(notebookDir, ".opennotes.json")
	if err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	return notebookDir
}
