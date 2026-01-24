package services

import (
	"encoding/json"
	"testing"
)

// Integration tests specifically for DuckDB type conversion in display service

func TestDisplay_RenderSQLResultsAsJSON_Integration_DuckDBConverter(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test that complex DuckDB types are properly converted through the integration
	input := []map[string]interface{}{
		{
			"note_id": 1,
			// Simulate DuckDB MAP type - map[interface{}]interface{}
			"metadata": map[interface{}]interface{}{
				"title":    "Test Note",
				"priority": 1,
				"tags":     []interface{}{"important", "work"},
				"settings": map[string]interface{}{
					"public": false,
					"draft":  true,
				},
			},
			// Simulate DuckDB ARRAY type - []interface{}
			"categories": []interface{}{"personal", "project", "urgent"},
			// Nested structure
			"audit": map[string]interface{}{
				"created": "2024-01-15T10:30:00Z",
				"history": []interface{}{
					map[string]interface{}{"action": "created", "user": "alice"},
					map[string]interface{}{"action": "updated", "user": "bob"},
				},
			},
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	// Parse the result back to verify proper conversion
	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]

	// Verify metadata conversion (map[interface{}]interface{} -> map[string]interface{})
	metadata, ok := row["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata should be converted to map[string]interface{}, got %T", row["metadata"])
	}

	if metadata["title"] != "Test Note" {
		t.Errorf("metadata.title = %v, want Test Note", metadata["title"])
	}

	if metadata["priority"] != float64(1) { // JSON numbers become float64
		t.Errorf("metadata.priority = %v, want 1", metadata["priority"])
	}

	// Verify nested array in metadata
	tags, ok := metadata["tags"].([]interface{})
	if !ok {
		t.Fatalf("metadata.tags should be []interface{}, got %T", metadata["tags"])
	}
	if len(tags) != 2 || tags[0] != "important" || tags[1] != "work" {
		t.Errorf("metadata.tags = %v, want [important work]", tags)
	}

	// Verify nested map in metadata
	settings, ok := metadata["settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata.settings should be map[string]interface{}, got %T", metadata["settings"])
	}
	if settings["public"] != false || settings["draft"] != true {
		t.Errorf("metadata.settings = %v, want {public: false, draft: true}", settings)
	}

	// Verify categories array conversion
	categories, ok := row["categories"].([]interface{})
	if !ok {
		t.Fatalf("categories should be []interface{}, got %T", row["categories"])
	}
	if len(categories) != 3 || categories[0] != "personal" || categories[1] != "project" || categories[2] != "urgent" {
		t.Errorf("categories = %v, want [personal project urgent]", categories)
	}

	// Verify nested audit structure
	audit, ok := row["audit"].(map[string]interface{})
	if !ok {
		t.Fatalf("audit should be map[string]interface{}, got %T", row["audit"])
	}

	history, ok := audit["history"].([]interface{})
	if !ok {
		t.Fatalf("audit.history should be []interface{}, got %T", audit["history"])
	}
	if len(history) != 2 {
		t.Errorf("audit.history length = %d, want 2", len(history))
	}

	entry1, ok := history[0].(map[string]interface{})
	if !ok {
		t.Fatalf("history[0] should be map[string]interface{}, got %T", history[0])
	}
	if entry1["action"] != "created" || entry1["user"] != "alice" {
		t.Errorf("history[0] = %v, want {action: created, user: alice}", entry1)
	}
}

func TestDisplay_RenderSQLResultsWithFormat_Table_ComplexTypes(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test table format with complex types
	input := []map[string]interface{}{
		{
			"id":   1,
			"name": "Test Note",
			// Complex map should be formatted as compact JSON
			"metadata": map[string]interface{}{
				"status": "active",
				"tags":   []interface{}{"work", "important"},
			},
			// Array should be formatted as JSON array
			"categories": []interface{}{"personal", "project"},
		},
	}

	output := captureOutput(func() {
		err = display.RenderSQLResultsWithFormat(input, "table")
		if err != nil {
			t.Errorf("RenderSQLResultsWithFormat(table) failed: %v", err)
		}
	})

	// Should not contain ugly Go formatting like "map[" or raw interface{} output
	if containsString(output, "map[") {
		t.Errorf("Table output should not contain Go map formatting: %s", output)
	}

	if containsString(output, "interface{}") {
		t.Errorf("Table output should not contain interface{} formatting: %s", output)
	}

	// Should contain JSON-formatted data
	if !containsString(output, "Test Note") {
		t.Errorf("Table output should contain note name: %s", output)
	}

	// Should contain compact JSON representations
	if !containsString(output, "{") || !containsString(output, "[") {
		t.Errorf("Table output should contain JSON-formatted complex types: %s", output)
	}
}

func TestDisplay_formatValueForTable_ComplexTypes(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tests := []struct {
		name        string
		input       interface{}
		contains    []string // strings that should be in the output
		notContains []string // strings that should NOT be in the output
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"status": "active",
				"count":  5,
			},
			contains:    []string{"{", "}", "status", "active"},
			notContains: []string{"map[", "interface{}"},
		},
		{
			name:        "simple array",
			input:       []interface{}{"work", "urgent", "project"},
			contains:    []string{"[", "]", "work", "urgent"},
			notContains: []string{"interface{}"},
		},
		{
			name:        "empty map",
			input:       map[string]interface{}{},
			contains:    []string{"{}"},
			notContains: []string{"map["},
		},
		{
			name:        "empty array",
			input:       []interface{}{},
			contains:    []string{"[]"},
			notContains: []string{"interface{}"},
		},
		{
			name:        "nil value",
			input:       nil,
			contains:    []string{"NULL"},
			notContains: []string{"<nil>"},
		},
		{
			name:        "long string truncation",
			input:       "This is a very long string that should be truncated when displayed in table format because it exceeds the 50 character limit",
			contains:    []string{"This is a very long string that should be tru", "..."},
			notContains: []string{"limit"}, // end of string should be cut off
		},
		{
			name: "complex nested structure truncation",
			input: map[string]interface{}{
				"verylongkey1": "verylongvalue1",
				"verylongkey2": "verylongvalue2",
				"verylongkey3": "verylongvalue3",
				"verylongkey4": "verylongvalue4",
			},
			contains:    []string{"{", "..."},
			notContains: []string{"verylongkey4"}, // Should be truncated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := display.formatValueForTable(tt.input)

			// Check that result contains expected strings
			for _, should := range tt.contains {
				if !containsString(result, should) {
					t.Errorf("formatValueForTable() result = %q should contain %q", result, should)
				}
			}

			// Check that result doesn't contain unwanted strings
			for _, shouldNot := range tt.notContains {
				if containsString(result, shouldNot) {
					t.Errorf("formatValueForTable() result = %q should NOT contain %q", result, shouldNot)
				}
			}
		})
	}
}

func TestDisplay_Integration_TableAndJSON_Consistency(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test that both table and JSON formats use the same conversion logic
	input := []map[string]interface{}{
		{
			"id": 1,
			"metadata": map[interface{}]interface{}{
				"title": "Integration Test",
				"tags":  []interface{}{"test", "integration"},
			},
		},
	}

	// Test JSON format
	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var jsonResults []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonResults)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Test table format (capture output and verify it doesn't have ugly formatting)
	tableOutput := captureOutput(func() {
		err = display.RenderSQLResultsWithFormat(input, "table")
		if err != nil {
			t.Errorf("RenderSQLResultsWithFormat(table) failed: %v", err)
		}
	})

	// Verify JSON has proper structure
	if len(jsonResults) != 1 {
		t.Errorf("JSON results length = %d, want 1", len(jsonResults))
	}

	metadata, ok := jsonResults[0]["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("JSON metadata should be map[string]interface{}, got %T", jsonResults[0]["metadata"])
	}

	if metadata["title"] != "Integration Test" {
		t.Errorf("JSON metadata.title = %v, want Integration Test", metadata["title"])
	}

	// Verify table doesn't have ugly Go formatting
	if containsString(tableOutput, "map[") {
		t.Errorf("Table output contains Go map formatting: %s", tableOutput)
	}

	if containsString(tableOutput, "interface{}") {
		t.Errorf("Table output contains interface{} formatting: %s", tableOutput)
	}

	// Both should handle the conversion properly
	if !containsString(tableOutput, "Integra") { // Partial match since it's truncated
		t.Errorf("Table output should contain converted title: %s", tableOutput)
	}
}
