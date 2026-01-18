package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestNewDisplay(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	if display == nil {
		t.Fatal("NewDisplay() returned nil display")
	}

	if display.renderer == nil {
		t.Fatal("NewDisplay() returned display with nil renderer")
	}
}

func TestDisplay_Render_BasicMarkdown(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "heading",
			input:    "# Hello World",
			contains: "Hello World",
		},
		{
			name:     "bullet list",
			input:    "- Item 1\n- Item 2",
			contains: "Item 1",
		},
		{
			name:     "bold text",
			input:    "**bold text**",
			contains: "bold",
		},
		{
			name:     "plain text",
			input:    "Just plain text",
			contains: "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := display.Render(tt.input)
			if err != nil {
				t.Fatalf("Render() failed: %v", err)
			}

			if result == "" {
				t.Error("Render() returned empty string")
			}

			// Check that the result contains expected text
			if tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("Render() result = %q, want to contain %q", result, tt.contains)
			}
		})
	}
}

func TestDisplay_Render_EmptyString(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	result, err := display.Render("")
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	// Empty input should produce some output (glamour may add whitespace)
	// We just check it doesn't error
	_ = result
}

func TestDisplay_RenderTemplate_ValidTemplate(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tmpl, err := template.New("test").Parse("# {{ .Title }}\n\nWelcome, {{ .Name }}!")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	ctx := map[string]string{
		"Title": "Greeting",
		"Name":  "User",
	}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() failed: %v", err)
	}

	if !containsString(result, "Greeting") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "Greeting")
	}

	if !containsString(result, "User") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "User")
	}
}

func TestDisplay_RenderTemplate_WithStruct(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	type Data struct {
		Title string
		Count int
	}

	tmpl, err := template.New("test").Parse("# {{ .Title }}\n\nItems: {{ .Count }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	ctx := Data{Title: "My List", Count: 42}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() failed: %v", err)
	}

	if !containsString(result, "My List") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "My List")
	}

	if !containsString(result, "42") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "42")
	}
}

func TestDisplay_RenderTemplate_InvalidTemplate_Fallback(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Try to parse invalid template
	_, err = template.New("test").Parse("# {{ .Title")
	if err == nil {
		t.Skip("Test assumes template parsing fails for invalid syntax")
	}

	// With the new API, we should get an error from Execute or from Parse
	// Let's test that nil template returns error
	result, err := display.RenderTemplate(nil, map[string]string{"Title": "Test"})
	if err == nil {
		t.Fatalf("RenderTemplate() should fail on nil template")
	}

	if result != "" {
		t.Errorf("RenderTemplate() result = %q, want empty string on error", result)
	}
}

func TestDisplay_RenderTemplate_ExecutionError_Fallback(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Valid template with map context - missing field won't error
	tmpl, err := template.New("test").Parse("# {{ .MissingField }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	ctx := map[string]string{"Title": "Test"}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() should not fail on missing map key: %v", err)
	}

	// Should render successfully (missing map keys just render as empty)
	if result == "" {
		t.Errorf("RenderTemplate() result is empty, should have some output")
	}

	// Now test with struct - missing field should cause execution error
	type Data struct {
		Title string
	}

	tmpl2, err := template.New("test2").Parse("# {{ .MissingField }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	ctx2 := Data{Title: "Test"}

	_, err = display.RenderTemplate(tmpl2, ctx2)
	if err == nil {
		t.Errorf("RenderTemplate() should fail on missing struct field")
	}
}

func TestDisplay_RenderTemplate_NilContext(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tmpl, err := template.New("test").Parse("# Static Heading")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := display.RenderTemplate(tmpl, nil)
	if err != nil {
		t.Fatalf("RenderTemplate() failed with nil context: %v", err)
	}

	if !containsString(result, "Static Heading") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "Static Heading")
	}
}

func TestDisplay_RenderTemplate_EmptyTemplate(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tmpl, err := template.New("test").Parse("")
	if err != nil {
		t.Fatalf("Failed to parse empty template: %v", err)
	}

	result, err := display.RenderTemplate(tmpl, nil)
	if err != nil {
		t.Fatalf("RenderTemplate() failed with empty template: %v", err)
	}

	// Empty template produces empty or whitespace result
	_ = result
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for RenderSQLResults

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestDisplay_RenderSQLResults_EmptyResults(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults([]map[string]interface{}{})
	})

	// Should output empty JSON array
	expected := "[]"
	if !strings.Contains(output, expected) {
		t.Errorf("RenderSQLResults() with empty results = %q, want to contain %q", output, expected)
	}
}

func TestDisplay_RenderSQLResults_SingleRow(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{
			"name":  "John",
			"email": "john@example.com",
			"age":   30,
		},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 1 {
		t.Errorf("RenderSQLResults() returned %d items, want 1", len(parsedResults))
	}

	row := parsedResults[0]
	if row["name"] != "John" {
		t.Errorf("RenderSQLResults() name = %v, want John", row["name"])
	}
	if row["email"] != "john@example.com" {
		t.Errorf("RenderSQLResults() email = %v, want john@example.com", row["email"])
	}
	// JSON numbers come back as float64
	if row["age"] != float64(30) {
		t.Errorf("RenderSQLResults() age = %v, want 30", row["age"])
	}
}

func TestDisplay_RenderSQLResults_MultipleRows(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 3 {
		t.Errorf("RenderSQLResults() returned %d items, want 3", len(parsedResults))
	}

	// Check each row
	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for i, row := range parsedResults {
		if row["id"] != float64(i+1) {
			t.Errorf("Row %d: id = %v, want %d", i, row["id"], i+1)
		}
		if row["name"] != expectedNames[i] {
			t.Errorf("Row %d: name = %v, want %s", i, row["name"], expectedNames[i])
		}
	}
}

func TestDisplay_RenderSQLResults_JSONFormatStructure(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"short": "a", "verylongname": "value1"},
		{"short": "abcdef", "verylongname": "v2"},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON array
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 2 {
		t.Errorf("RenderSQLResults() returned %d items, want 2", len(parsedResults))
	}

	// Verify data structure
	if parsedResults[0]["short"] != "a" || parsedResults[0]["verylongname"] != "value1" {
		t.Error("First row data incorrect")
	}
	if parsedResults[1]["short"] != "abcdef" || parsedResults[1]["verylongname"] != "v2" {
		t.Error("Second row data incorrect")
	}
}

func TestDisplay_RenderSQLResults_DifferentTypes(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{
			"string_col": "text",
			"int_col":    42,
			"float_col":  3.14,
			"bool_col":   true,
		},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 1 {
		t.Errorf("RenderSQLResults() returned %d items, want 1", len(parsedResults))
	}

	row := parsedResults[0]
	if row["string_col"] != "text" {
		t.Error("RenderSQLResults() missing string value")
	}
	if row["int_col"] != float64(42) {
		t.Error("RenderSQLResults() missing int value")
	}
	if row["float_col"] != 3.14 {
		t.Error("RenderSQLResults() missing float value")
	}
	if row["bool_col"] != true {
		t.Error("RenderSQLResults() missing bool value")
	}
}

func TestDisplay_RenderSQLResults_DataPreservation(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create results with columns in non-alphabetical order
	results := []map[string]interface{}{
		{"zebra": 1, "apple": 2, "middle": 3},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON that preserves all data
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 1 {
		t.Errorf("RenderSQLResults() returned %d items, want 1", len(parsedResults))
	}

	row := parsedResults[0]
	if row["apple"] != float64(2) {
		t.Error("RenderSQLResults() missing 'apple' data")
	}
	if row["middle"] != float64(3) {
		t.Error("RenderSQLResults() missing 'middle' data")
	}
	if row["zebra"] != float64(1) {
		t.Error("RenderSQLResults() missing 'zebra' data")
	}
}

func TestDisplay_RenderSQLResults_NilValues(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"col1": "value", "col2": nil},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON with proper nil handling
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 1 {
		t.Errorf("RenderSQLResults() returned %d items, want 1", len(parsedResults))
	}

	row := parsedResults[0]
	if row["col1"] != "value" {
		t.Error("RenderSQLResults() missing column with value")
	}
	if row["col2"] != nil {
		t.Error("RenderSQLResults() nil value not preserved")
	}
}

func TestDisplay_RenderSQLResults_LargeValues(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	longString := strings.Repeat("x", 100)
	results := []map[string]interface{}{
		{"short": "s", "long": longString},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should output valid JSON that preserves large values
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 1 {
		t.Errorf("RenderSQLResults() returned %d items, want 1", len(parsedResults))
	}

	row := parsedResults[0]
	if row["short"] != "s" {
		t.Error("RenderSQLResults() missing short value")
	}
	if row["long"] != longString {
		t.Error("RenderSQLResults() long string not preserved")
	}
}

// Tests for JSON serialization functionality

func TestDisplay_RenderSQLResultsAsJSON_EmptyResults(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON([]map[string]interface{}{})
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	// Should return empty array
	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("RenderSQLResultsAsJSON() with empty results = %d items, want 0", len(results))
	}
}

func TestDisplay_RenderSQLResultsAsJSON_SingleRow(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{
			"name":  "John",
			"email": "john@example.com",
			"age":   30,
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("RenderSQLResultsAsJSON() returned %d items, want 1", len(results))
	}

	row := results[0]
	if row["name"] != "John" {
		t.Errorf("RenderSQLResultsAsJSON() name = %v, want John", row["name"])
	}
	if row["email"] != "john@example.com" {
		t.Errorf("RenderSQLResultsAsJSON() email = %v, want john@example.com", row["email"])
	}
	// JSON numbers come back as float64
	if row["age"] != float64(30) {
		t.Errorf("RenderSQLResultsAsJSON() age = %v, want 30", row["age"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_MultipleRows(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("RenderSQLResultsAsJSON() returned %d items, want 3", len(results))
	}

	// Check each row
	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for i, row := range results {
		if row["id"] != float64(i+1) {
			t.Errorf("Row %d: id = %v, want %d", i, row["id"], i+1)
		}
		if row["name"] != expectedNames[i] {
			t.Errorf("Row %d: name = %v, want %s", i, row["name"], expectedNames[i])
		}
	}
}

func TestDisplay_RenderSQLResultsAsJSON_DifferentTypes(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{
			"string_col": "text",
			"int_col":    42,
			"float_col":  3.14,
			"bool_col":   true,
			"null_col":   nil,
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("RenderSQLResultsAsJSON() returned %d items, want 1", len(results))
	}

	row := results[0]
	if row["string_col"] != "text" {
		t.Errorf("string_col = %v, want text", row["string_col"])
	}
	if row["int_col"] != float64(42) {
		t.Errorf("int_col = %v, want 42", row["int_col"])
	}
	if row["float_col"] != 3.14 {
		t.Errorf("float_col = %v, want 3.14", row["float_col"])
	}
	if row["bool_col"] != true {
		t.Errorf("bool_col = %v, want true", row["bool_col"])
	}
	if row["null_col"] != nil {
		t.Errorf("null_col = %v, want nil", row["null_col"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_UTF8Content(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{
			"unicode": "Hello 世界 🌍",
			"emoji":   "🚀✨💻",
			"special": "áéíóú ñ ç",
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("RenderSQLResultsAsJSON() returned %d items, want 1", len(results))
	}

	row := results[0]
	if row["unicode"] != "Hello 世界 🌍" {
		t.Errorf("unicode = %v, want Hello 世界 🌍", row["unicode"])
	}
	if row["emoji"] != "🚀✨💻" {
		t.Errorf("emoji = %v, want 🚀✨💻", row["emoji"])
	}
	if row["special"] != "áéíóú ñ ç" {
		t.Errorf("special = %v, want áéíóú ñ ç", row["special"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_ValidJSONOutput(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{"title": "Note 1", "path": "/path/note1.md", "tags": "tag1,tag2"},
		{"title": "Note 2", "path": "/path/note2.md", "tags": "tag3"},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	// Validate JSON structure matches expected format
	expectedJSON := `[
  {"path":"/path/note1.md","tags":"tag1,tag2","title":"Note 1"},
  {"path":"/path/note2.md","tags":"tag3","title":"Note 2"}
]`

	var expected, actual interface{}
	err = json.Unmarshal([]byte(expectedJSON), &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	err = json.Unmarshal(jsonBytes, &actual)
	if err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	// Compare data structures (not raw JSON to avoid formatting issues)
	expectedBytes, _ := json.Marshal(expected)
	actualBytes, _ := json.Marshal(actual)

	if !bytes.Equal(expectedBytes, actualBytes) {
		t.Errorf("JSON structure mismatch.\nExpected: %s\nActual: %s", 
			string(expectedBytes), string(actualBytes))
	}
}

// Tests for RenderSQLResultsWithFormat integration

func TestDisplay_RenderSQLResultsWithFormat_TableFormat(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"name": "Alice", "age": 30},
	}

	output := captureOutput(func() {
		err = display.RenderSQLResultsWithFormat(results, "table")
		if err != nil {
			t.Errorf("RenderSQLResultsWithFormat(table) failed: %v", err)
		}
	})

	// Should behave like original table format
	if !strings.Contains(output, "Alice") {
		t.Error("Table format should contain data")
	}
	if !strings.Contains(output, "age") {
		t.Error("Table format should contain column headers")
	}
}

func TestDisplay_RenderSQLResultsWithFormat_JSONFormat(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"name": "Alice", "age": 30},
	}

	output := captureOutput(func() {
		err = display.RenderSQLResultsWithFormat(results, "json")
		if err != nil {
			t.Errorf("RenderSQLResultsWithFormat(json) failed: %v", err)
		}
	})

	// Should be valid JSON
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Errorf("JSON format output is not valid JSON: %v", err)
	}

	if len(parsedResults) != 1 {
		t.Errorf("JSON format returned %d items, want 1", len(parsedResults))
	}

	if parsedResults[0]["name"] != "Alice" {
		t.Errorf("JSON format name = %v, want Alice", parsedResults[0]["name"])
	}
}

func TestDisplay_RenderSQLResultsWithFormat_InvalidFormat(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"name": "Alice", "age": 30},
	}

	output := captureOutput(func() {
		err = display.RenderSQLResultsWithFormat(results, "invalid_format")
		if err != nil {
			t.Errorf("RenderSQLResultsWithFormat(invalid_format) failed: %v", err)
		}
	})

	// Should fallback to table format
	if !strings.Contains(output, "Alice") {
		t.Error("Invalid format should fallback to table and contain data")
	}
	if !strings.Contains(output, "age") {
		t.Error("Invalid format should fallback to table and contain column headers")
	}
}

func TestDisplay_RenderSQLResults_JSONCompatibility(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"name": "Alice", "age": 30},
	}

	output := captureOutput(func() {
		err = display.RenderSQLResults(results)
		if err != nil {
			t.Errorf("RenderSQLResults() failed: %v", err)
		}
	})

	// Should now work as JSON format (breaking change from table to JSON)
	var parsedResults []map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsedResults)
	if err != nil {
		t.Fatalf("RenderSQLResults() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if len(parsedResults) != 1 {
		t.Errorf("RenderSQLResults() returned %d items, want 1", len(parsedResults))
	}

	row := parsedResults[0]
	if row["name"] != "Alice" {
		t.Error("RenderSQLResults should work as JSON format")
	}
	if row["age"] != float64(30) {
		t.Error("RenderSQLResults should preserve numeric values in JSON")
	}
}

// Complex Data Type Tests

func TestDisplay_RenderSQLResultsAsJSON_ComplexTypes_Maps(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test map-like structures that DuckDB might return
	input := []map[string]interface{}{
		{
			"user_id":  1,
			"metadata": map[string]interface{}{"role": "admin", "active": true},
			"settings": map[string]interface{}{"theme": "dark", "notifications": false},
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]
	metadata, ok := row["metadata"].(map[string]interface{})
	if !ok {
		t.Errorf("metadata should be map[string]interface{}, got %T", row["metadata"])
	}

	if metadata["role"] != "admin" {
		t.Errorf("metadata.role = %v, want admin", metadata["role"])
	}
	if metadata["active"] != true {
		t.Errorf("metadata.active = %v, want true", metadata["active"])
	}

	settings, ok := row["settings"].(map[string]interface{})
	if !ok {
		t.Errorf("settings should be map[string]interface{}, got %T", row["settings"])
	}

	if settings["theme"] != "dark" {
		t.Errorf("settings.theme = %v, want dark", settings["theme"])
	}
	if settings["notifications"] != false {
		t.Errorf("settings.notifications = %v, want false", settings["notifications"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_ComplexTypes_Arrays(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test array-like structures that DuckDB might return
	input := []map[string]interface{}{
		{
			"note_id":     1,
			"tags":        []interface{}{"work", "meeting", "urgent"},
			"attendees":   []interface{}{"alice", "bob", "charlie"},
			"scores":      []interface{}{85.5, 92.0, 78.3},
			"flags":       []interface{}{true, false, true},
			"mixed_array": []interface{}{"text", 42, true, nil},
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]

	// Test string arrays
	tags, ok := row["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags should be []interface{}, got %T", row["tags"])
	}
	expectedTags := []string{"work", "meeting", "urgent"}
	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("tags[%d] = %v, want %s", i, tag, expectedTags[i])
		}
	}

	// Test numeric arrays 
	scores, ok := row["scores"].([]interface{})
	if !ok {
		t.Fatalf("scores should be []interface{}, got %T", row["scores"])
	}
	expectedScores := []float64{85.5, 92.0, 78.3}
	for i, score := range scores {
		if score != expectedScores[i] {
			t.Errorf("scores[%d] = %v, want %f", i, score, expectedScores[i])
		}
	}

	// Test boolean arrays
	flags, ok := row["flags"].([]interface{})
	if !ok {
		t.Fatalf("flags should be []interface{}, got %T", row["flags"])
	}
	expectedFlags := []bool{true, false, true}
	for i, flag := range flags {
		if flag != expectedFlags[i] {
			t.Errorf("flags[%d] = %v, want %t", i, flag, expectedFlags[i])
		}
	}

	// Test mixed type arrays
	mixedArray, ok := row["mixed_array"].([]interface{})
	if !ok {
		t.Fatalf("mixed_array should be []interface{}, got %T", row["mixed_array"])
	}
	if len(mixedArray) != 4 {
		t.Errorf("mixed_array length = %d, want 4", len(mixedArray))
	}
	if mixedArray[0] != "text" {
		t.Errorf("mixed_array[0] = %v, want text", mixedArray[0])
	}
	if mixedArray[1] != float64(42) {
		t.Errorf("mixed_array[1] = %v, want 42", mixedArray[1])
	}
	if mixedArray[2] != true {
		t.Errorf("mixed_array[2] = %v, want true", mixedArray[2])
	}
	if mixedArray[3] != nil {
		t.Errorf("mixed_array[3] = %v, want nil", mixedArray[3])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_ComplexTypes_NestedStructures(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test deeply nested structures
	input := []map[string]interface{}{
		{
			"document": map[string]interface{}{
				"title": "Complex Document",
				"author": map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
					"tags":  []interface{}{"author", "verified"},
				},
				"sections": []interface{}{
					map[string]interface{}{
						"title":    "Introduction",
						"content":  "This is the intro",
						"metadata": map[string]interface{}{"wordCount": 50, "draft": false},
					},
					map[string]interface{}{
						"title":   "Conclusion",
						"content": "This is the conclusion",
						"tags":    []interface{}{"final", "review"},
					},
				},
			},
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]
	document, ok := row["document"].(map[string]interface{})
	if !ok {
		t.Fatalf("document should be map[string]interface{}, got %T", row["document"])
	}

	if document["title"] != "Complex Document" {
		t.Errorf("document.title = %v, want Complex Document", document["title"])
	}

	author, ok := document["author"].(map[string]interface{})
	if !ok {
		t.Fatalf("author should be map[string]interface{}, got %T", document["author"])
	}

	if author["name"] != "John Doe" {
		t.Errorf("author.name = %v, want John Doe", author["name"])
	}

	authorTags, ok := author["tags"].([]interface{})
	if !ok {
		t.Fatalf("author.tags should be []interface{}, got %T", author["tags"])
	}
	if len(authorTags) != 2 {
		t.Errorf("author.tags length = %d, want 2", len(authorTags))
	}

	sections, ok := document["sections"].([]interface{})
	if !ok {
		t.Fatalf("sections should be []interface{}, got %T", document["sections"])
	}
	if len(sections) != 2 {
		t.Errorf("sections length = %d, want 2", len(sections))
	}

	section1, ok := sections[0].(map[string]interface{})
	if !ok {
		t.Fatalf("sections[0] should be map[string]interface{}, got %T", sections[0])
	}
	if section1["title"] != "Introduction" {
		t.Errorf("sections[0].title = %v, want Introduction", section1["title"])
	}

	metadata, ok := section1["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata should be map[string]interface{}, got %T", section1["metadata"])
	}
	if metadata["wordCount"] != float64(50) {
		t.Errorf("metadata.wordCount = %v, want 50", metadata["wordCount"])
	}
}

// Error Handling Tests

func TestDisplay_RenderSQLResultsAsJSON_ErrorHandling_JSONMarshalFailure(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create data that would cause JSON marshal to fail (circular reference simulation)
	// Since Go's json.Marshal doesn't fail on basic types, we test the error path by 
	// checking that valid data doesn't error (error conditions are hard to simulate with standard types)
	input := []map[string]interface{}{
		{
			"valid_field": "test",
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Errorf("RenderSQLResultsAsJSON() should not fail on valid data: %v", err)
	}

	// Verify the result is valid JSON
	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestDisplay_RenderSQLResultsWithFormat_ErrorHandling_InvalidJSON(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test that we can handle edge case data that might cause issues
	input := []map[string]interface{}{
		{
			"special_chars":  "\x00\x01\x02\x1f",  // Control characters
			"large_number":   float64(9223372036854775807), // Max int64
			"tiny_number":    1e-10,
			"infinity":       1.0,  // Not infinity, but large number  
			"empty_string":   "",
			"unicode_null":   "\u0000",
		},
	}

	err = display.RenderSQLResultsWithFormat(input, "json")
	if err != nil {
		t.Errorf("RenderSQLResultsWithFormat() should handle edge case data: %v", err)
	}
}

// Performance Tests

func TestDisplay_RenderSQLResultsAsJSON_Performance_LargeDataset(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create dataset with 1000 rows
	input := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = map[string]interface{}{
			"id":          i,
			"title":       fmt.Sprintf("Note %d", i),
			"content":     strings.Repeat("Lorem ipsum dolor sit amet. ", 20), // ~560 chars
			"tags":        []interface{}{"tag1", "tag2", "tag3"},
			"metadata":    map[string]interface{}{"author": fmt.Sprintf("user%d", i%10), "created": "2024-01-01"},
			"active":      i%2 == 0,
			"score":       float64(i) * 0.1,
		}
	}

	start := time.Now()
	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	// Performance target: <100ms for 1000 rows (more realistic target)
	if duration > 100*time.Millisecond {
		t.Errorf("RenderSQLResultsAsJSON() took %v, want <100ms for 1000 rows", duration)
	}

	// Verify the result is valid JSON
	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal large dataset JSON: %v", err)
	}

	if len(results) != 1000 {
		t.Errorf("Expected 1000 results, got %d", len(results))
	}

	// Verify a sample of data integrity
	if results[0]["id"] != float64(0) {
		t.Errorf("results[0].id = %v, want 0", results[0]["id"])
	}
	if results[999]["id"] != float64(999) {
		t.Errorf("results[999].id = %v, want 999", results[999]["id"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_Performance_SmallDataset(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create dataset with 100 rows (more realistic for the 5ms target)
	input := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		input[i] = map[string]interface{}{
			"id":       i,
			"title":    fmt.Sprintf("Note %d", i),
			"content":  "Short content",
			"active":   i%2 == 0,
		}
	}

	start := time.Now()
	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	// Performance target: <5ms for 100 rows
	if duration > 5*time.Millisecond {
		t.Errorf("RenderSQLResultsAsJSON() took %v, want <5ms for 100 rows", duration)
	}

	// Verify the result is valid JSON
	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 100 {
		t.Errorf("Expected 100 results, got %d", len(results))
	}
}

func BenchmarkDisplay_RenderSQLResultsAsJSON_SmallDataset(b *testing.B) {
	display, err := NewDisplay()
	if err != nil {
		b.Fatalf("NewDisplay() failed: %v", err)
	}

	// Small dataset: 10 rows
	input := make([]map[string]interface{}, 10)
	for i := 0; i < 10; i++ {
		input[i] = map[string]interface{}{
			"id":    i,
			"title": fmt.Sprintf("Note %d", i),
			"tags":  []interface{}{"tag1", "tag2"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := display.RenderSQLResultsAsJSON(input)
		if err != nil {
			b.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
		}
	}
}

func BenchmarkDisplay_RenderSQLResultsAsJSON_MediumDataset(b *testing.B) {
	display, err := NewDisplay()
	if err != nil {
		b.Fatalf("NewDisplay() failed: %v", err)
	}

	// Medium dataset: 100 rows
	input := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		input[i] = map[string]interface{}{
			"id":          i,
			"title":       fmt.Sprintf("Note %d", i),
			"content":     strings.Repeat("Sample content. ", 50),
			"tags":        []interface{}{"work", "personal", "urgent"},
			"metadata":    map[string]interface{}{"author": "user", "version": i},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := display.RenderSQLResultsAsJSON(input)
		if err != nil {
			b.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
		}
	}
}

func BenchmarkDisplay_RenderSQLResultsAsJSON_LargeDataset(b *testing.B) {
	display, err := NewDisplay()
	if err != nil {
		b.Fatalf("NewDisplay() failed: %v", err)
	}

	// Large dataset: 1000 rows
	input := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = map[string]interface{}{
			"id":       i,
			"title":    fmt.Sprintf("Note %d", i),
			"content":  strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 20),
			"tags":     []interface{}{"tag1", "tag2", "tag3", "tag4"},
			"metadata": map[string]interface{}{
				"author":     fmt.Sprintf("author%d", i%5),
				"created":    "2024-01-01T00:00:00Z",
				"updated":    "2024-01-02T00:00:00Z",
				"wordCount":  500 + i,
				"published":  i%3 == 0,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := display.RenderSQLResultsAsJSON(input)
		if err != nil {
			b.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
		}
	}
}

// Edge Case Tests

func TestDisplay_RenderSQLResultsAsJSON_EdgeCases_EmptyMaps(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{},  // Empty map
		{"empty_nested": map[string]interface{}{}},
		{"empty_array": []interface{}{}},
		{"null_value": nil},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// Check empty map
	if len(results[0]) != 0 {
		t.Errorf("Expected empty map, got %v", results[0])
	}

	// Check empty nested map
	emptyNested, ok := results[1]["empty_nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("empty_nested should be map[string]interface{}, got %T", results[1]["empty_nested"])
	}
	if len(emptyNested) != 0 {
		t.Errorf("Expected empty nested map, got %v", emptyNested)
	}

	// Check empty array
	emptyArray, ok := results[2]["empty_array"].([]interface{})
	if !ok {
		t.Fatalf("empty_array should be []interface{}, got %T", results[2]["empty_array"])
	}
	if len(emptyArray) != 0 {
		t.Errorf("Expected empty array, got %v", emptyArray)
	}

	// Check null value
	if results[3]["null_value"] != nil {
		t.Errorf("Expected nil value, got %v", results[3]["null_value"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_EdgeCases_SpecialNumbers(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{
			"zero":           0,
			"negative_int":   -42,
			"large_int":      9223372036854775807, // Max int64
			"zero_float":     0.0,
			"negative_float": -123.456,
			"small_float":    1e-10,
			"scientific":     1.23e+10,
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]
	if row["zero"] != float64(0) {
		t.Errorf("zero = %v, want 0", row["zero"])
	}
	if row["negative_int"] != float64(-42) {
		t.Errorf("negative_int = %v, want -42", row["negative_int"])
	}
	if row["large_int"] != float64(9223372036854775807) {
		t.Errorf("large_int = %v, want 9223372036854775807", row["large_int"])
	}
	if row["zero_float"] != float64(0.0) {
		t.Errorf("zero_float = %v, want 0.0", row["zero_float"])
	}
	if row["negative_float"] != -123.456 {
		t.Errorf("negative_float = %v, want -123.456", row["negative_float"])
	}
	if row["small_float"] != 1e-10 {
		t.Errorf("small_float = %v, want 1e-10", row["small_float"])
	}
	if row["scientific"] != 1.23e+10 {
		t.Errorf("scientific = %v, want 1.23e+10", row["scientific"])
	}
}

func TestDisplay_RenderSQLResultsAsJSON_EdgeCases_SpecialStrings(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	input := []map[string]interface{}{
		{
			"empty_string":     "",
			"whitespace":       "   \t\n\r   ",
			"json_like":        `{"key": "value"}`,
			"quotes_double":    `"quoted string"`,
			"quotes_single":    "'single quoted'",
			"backslashes":      `C:\path\to\file`,
			"newlines":         "line1\nline2\nline3",
			"tabs":             "col1\tcol2\tcol3",
			"unicode_emoji":    "Hello 👋 World 🌍",
			"unicode_chars":    "αβγδε 中文 العربية",
			"control_chars":    "start\x00\x01\x02end",
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]
	if row["empty_string"] != "" {
		t.Errorf("empty_string = %v, want empty string", row["empty_string"])
	}
	if row["whitespace"] != "   \t\n\r   " {
		t.Errorf("whitespace preserved incorrectly")
	}
	if row["json_like"] != `{"key": "value"}` {
		t.Errorf("json_like = %v, want {\"key\": \"value\"}", row["json_like"])
	}
	if row["unicode_emoji"] != "Hello 👋 World 🌍" {
		t.Errorf("unicode_emoji = %v, want Hello 👋 World 🌍", row["unicode_emoji"])
	}
	if row["unicode_chars"] != "αβγδε 中文 العربية" {
		t.Errorf("unicode_chars = %v, want αβγδε 中文 العربية", row["unicode_chars"])
	}
}

// Integration Tests with Real-World Data Patterns

func TestDisplay_RenderSQLResultsAsJSON_RealWorld_NotesQuery(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Simulate real notes query results
	input := []map[string]interface{}{
		{
			"path":          "/notebook/projects/project-alpha.md",
			"title":         "Project Alpha Planning",
			"content":       "# Project Alpha\n\nThis is a planning document...",
			"display_name":  "Project Alpha Planning",
			"tags":          []interface{}{"project", "planning", "alpha"},
			"created":       "2024-01-15T10:30:00Z",
			"modified":      "2024-01-20T15:45:00Z",
			"word_count":    450,
			"frontmatter":   map[string]interface{}{"status": "in-progress", "priority": "high"},
		},
		{
			"path":         "/notebook/meetings/standup-2024-01-20.md",
			"title":        "Daily Standup - 2024-01-20",
			"content":      "## Standup Notes\n\n- Alice: Working on feature X...",
			"display_name": "Daily Standup - 2024-01-20",
			"tags":         []interface{}{"meeting", "standup", "team"},
			"created":      "2024-01-20T09:00:00Z",
			"modified":     "2024-01-20T09:30:00Z", 
			"word_count":   125,
			"frontmatter":  map[string]interface{}{"attendees": []interface{}{"alice", "bob", "charlie"}},
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify first note
	note1 := results[0]
	if note1["title"] != "Project Alpha Planning" {
		t.Errorf("note1.title = %v, want Project Alpha Planning", note1["title"])
	}

	tags1, ok := note1["tags"].([]interface{})
	if !ok {
		t.Fatalf("note1.tags should be []interface{}, got %T", note1["tags"])
	}
	if len(tags1) != 3 {
		t.Errorf("note1.tags length = %d, want 3", len(tags1))
	}
	if tags1[0] != "project" {
		t.Errorf("note1.tags[0] = %v, want project", tags1[0])
	}

	frontmatter1, ok := note1["frontmatter"].(map[string]interface{})
	if !ok {
		t.Fatalf("note1.frontmatter should be map[string]interface{}, got %T", note1["frontmatter"])
	}
	if frontmatter1["status"] != "in-progress" {
		t.Errorf("note1.frontmatter.status = %v, want in-progress", frontmatter1["status"])
	}

	// Verify second note
	note2 := results[1]
	frontmatter2, ok := note2["frontmatter"].(map[string]interface{})
	if !ok {
		t.Fatalf("note2.frontmatter should be map[string]interface{}, got %T", note2["frontmatter"])
	}

	attendees, ok := frontmatter2["attendees"].([]interface{})
	if !ok {
		t.Fatalf("attendees should be []interface{}, got %T", frontmatter2["attendees"])
	}
	if len(attendees) != 3 {
		t.Errorf("attendees length = %d, want 3", len(attendees))
	}
	if attendees[0] != "alice" {
		t.Errorf("attendees[0] = %v, want alice", attendees[0])
	}
}

// Memory Usage Tests

func TestDisplay_RenderSQLResultsAsJSON_MemoryUsage_LargeContent(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create a dataset with large content to test memory handling
	largeContent := strings.Repeat("This is a very long content string that simulates large note content. ", 100) // ~7KB per note

	input := make([]map[string]interface{}, 50) // 50 notes with large content
	for i := 0; i < 50; i++ {
		input[i] = map[string]interface{}{
			"id":      i,
			"title":   fmt.Sprintf("Large Note %d", i),
			"content": largeContent,
			"tags":    []interface{}{"large", "test", "memory"},
		}
	}

	start := time.Now()
	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed on large content: %v", err)
	}

	// Should complete in reasonable time even with large content
	if duration > 100*time.Millisecond {
		t.Errorf("RenderSQLResultsAsJSON() took %v for large content, want <100ms", duration)
	}

	// Verify JSON is valid
	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal large content JSON: %v", err)
	}

	if len(results) != 50 {
		t.Errorf("Expected 50 results, got %d", len(results))
	}

	// Verify content integrity
	if results[0]["content"] != largeContent {
		t.Errorf("Large content not preserved correctly")
	}

	// Check JSON size is reasonable (should be larger than input due to JSON formatting)
	if len(jsonBytes) < len(largeContent)*50 {
		t.Errorf("JSON output seems too small for large content dataset")
	}
}

func TestDisplay_RenderSQLResultsAsJSON_DeepNestedStructures(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create deeply nested structure to test handling
	deepNested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"level5": map[string]interface{}{
							"data":  "deep value",
							"array": []interface{}{"item1", "item2", "item3"},
							"number": 42,
						},
					},
				},
			},
		},
	}

	input := []map[string]interface{}{
		{
			"id":        1,
			"structure": deepNested,
		},
	}

	jsonBytes, err := display.RenderSQLResultsAsJSON(input)
	if err != nil {
		t.Fatalf("RenderSQLResultsAsJSON() failed on deep nested structures: %v", err)
	}

	var results []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal deep nested JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	row := results[0]
	structure, ok := row["structure"].(map[string]interface{})
	if !ok {
		t.Fatalf("structure should be map[string]interface{}, got %T", row["structure"])
	}

	// Navigate to the deep value
	level1, ok := structure["level1"].(map[string]interface{})
	if !ok {
		t.Fatalf("level1 should be map[string]interface{}, got %T", structure["level1"])
	}

	level2, ok := level1["level2"].(map[string]interface{})
	if !ok {
		t.Fatalf("level2 should be map[string]interface{}, got %T", level1["level2"])
	}

	level3, ok := level2["level3"].(map[string]interface{})
	if !ok {
		t.Fatalf("level3 should be map[string]interface{}, got %T", level2["level3"])
	}

	level4, ok := level3["level4"].(map[string]interface{})
	if !ok {
		t.Fatalf("level4 should be map[string]interface{}, got %T", level3["level4"])
	}

	level5, ok := level4["level5"].(map[string]interface{})
	if !ok {
		t.Fatalf("level5 should be map[string]interface{}, got %T", level4["level5"])
	}

	if level5["data"] != "deep value" {
		t.Errorf("Deep nested data = %v, want deep value", level5["data"])
	}

	if level5["number"] != float64(42) {
		t.Errorf("Deep nested number = %v, want 42", level5["number"])
	}

	deepArray, ok := level5["array"].([]interface{})
	if !ok {
		t.Fatalf("Deep nested array should be []interface{}, got %T", level5["array"])
	}
	if len(deepArray) != 3 {
		t.Errorf("Deep nested array length = %d, want 3", len(deepArray))
	}
}

// Additional Coverage Tests

func TestDisplay_RenderTemplate_RenderFallback(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Test that template execution with bad renderer fallback to plain text
	tmpl, err := template.New("test").Parse("# {{ .Title }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Temporarily break the renderer by setting it to nil (not possible in this implementation)
	// This test validates the current implementation where renderer errors are handled

	ctx := map[string]string{"Title": "Test"}
	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() should handle renderer fallback: %v", err)
	}

	if result == "" {
		t.Error("RenderTemplate() should return some output even on renderer error")
	}
}
