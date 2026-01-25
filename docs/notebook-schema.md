# Notebook Configuration Schema

This document describes the JSON schema for `.opennotes.json` notebook configuration files.

## Purpose

The schema provides:
1. **IDE Support**: Autocomplete and validation in editors (VSCode, IntelliJ, etc.)
2. **Documentation**: Clear reference for configuration structure
3. **Validation**: Catch configuration errors early

## Schema Location

The schema will be available at:
- Local: `./opennotes.schema.json` (generated on notebook creation)
- Remote: `https://opennotes.dev/schema/v1/notebook.json` (future)

## Configuration Structure

Based on `internal/services/notebook.go`:

```go
type StoredNotebookConfig struct {
	Root      string            `json:"root"`
	Name      string            `json:"name"`
	Contexts  []string          `json:"contexts,omitempty"`
	Templates map[string]string `json:"templates,omitempty"`
	Groups    []NotebookGroup   `json:"groups,omitempty"`
}

type NotebookGroup struct {
	Name     string         `json:"name"`
	Globs    []string       `json:"globs"`
	Metadata map[string]any `json:"metadata"`
	Template string         `json:"template,omitempty"`
}
```

## Current Schema Draft

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "OpenNotes Notebook Configuration",
  "description": "Configuration for an OpenNotes notebook",
  "type": "object",
  "required": ["root", "name"],
  "properties": {
    "$schema": {
      "type": "string",
      "description": "JSON Schema reference for IDE support"
    },
    "root": {
      "type": "string",
      "description": "Root directory for notes (relative to config file)",
      "examples": [".", ".notes", ".memory"]
    },
    "name": {
      "type": "string",
      "description": "Notebook name (used for display and identification)",
      "minLength": 1
    },
    "contexts": {
      "type": "array",
      "description": "Directories where this notebook is active (for auto-discovery)",
      "items": {
        "type": "string"
      },
      "uniqueItems": true
    },
    "templates": {
      "type": "object",
      "description": "Named templates mapping to file paths",
      "additionalProperties": {
        "type": "string"
      }
    },
    "groups": {
      "type": "array",
      "description": "Note groups with glob patterns and metadata",
      "items": {
        "type": "object",
        "required": ["name", "globs"],
        "properties": {
          "name": {
            "type": "string",
            "description": "Group name"
          },
          "globs": {
            "type": "array",
            "description": "Glob patterns for matching notes",
            "items": {
              "type": "string"
            },
            "minItems": 1
          },
          "metadata": {
            "type": "object",
            "description": "Default metadata for notes in this group"
          },
          "template": {
            "type": "string",
            "description": "Default template for notes in this group"
          }
        }
      }
    }
  }
}
```

## Usage

Once implemented, users can reference the schema in their `.opennotes.json`:

```json
{
  "$schema": "./opennotes.schema.json",
  "root": ".notes",
  "name": "My Notebook",
  "contexts": ["/path/to/workspace"],
  "groups": [
    {
      "name": "Default",
      "globs": ["**/*.md"],
      "metadata": {}
    }
  ]
}
```

## Implementation Plan

1. Create `opennotes.schema.json` file
2. Modify `opennotes notebook create` to generate schema file
3. Update existing `.opennotes.json` files to include `$schema` reference
4. Add validation command: `opennotes notebook validate`
5. Document in main README and guides
