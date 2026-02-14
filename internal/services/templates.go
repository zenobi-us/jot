package services

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.gotmpl
var templateFiles embed.FS

var loadedTemplates map[string]*template.Template

// init loads all templates on package initialization.
func init() {
	loadedTemplates = make(map[string]*template.Template)

	templateNames := []string{"note-list", "note-detail", "notebook-info", "notebook-list", "note-search-semantic"}
	for _, name := range templateNames {
		tmpl, err := loadTemplate(name)
		if err != nil {
			// Log warning but don't fail - templates may be optional
			fmt.Printf("warning: failed to load template %s: %v\n", name, err)
			continue
		}
		loadedTemplates[name] = tmpl
	}
}

// loadTemplate loads a template by name from the embedded filesystem.
func loadTemplate(name string) (*template.Template, error) {
	content, err := templateFiles.ReadFile(fmt.Sprintf("templates/%s.gotmpl", name))
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// TuiRender is a convenience function to render a template by name with glamour.
func TuiRender(name string, ctx any) (string, error) {
	// Get the pre-loaded template
	tmpl, ok := loadedTemplates[name]
	if !ok {
		return "", fmt.Errorf("template %q not found", name)
	}

	display, err := NewDisplay()
	if err != nil {
		// Fallback without glamour rendering
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, ctx); err != nil {
			return "", err
		}

		return buf.String(), nil
	}

	return display.RenderTemplate(tmpl, ctx)
}
