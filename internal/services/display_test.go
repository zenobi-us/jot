package services

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"text/template"
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
