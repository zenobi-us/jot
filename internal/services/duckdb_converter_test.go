package services

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestDuckDBConverter_ConvertValue_Primitives(t *testing.T) {
	converter := NewDuckDBConverter()

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"nil", nil, nil},
		{"string", "hello", "hello"},
		{"int", 42, 42},
		{"float64", 3.14, 3.14},
		{"bool true", true, true},
		{"bool false", false, false},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.convertValue(tt.input)
			if err != nil {
				t.Errorf("convertValue() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("convertValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDuckDBConverter_ConvertValue_Time(t *testing.T) {
	converter := NewDuckDBConverter()

	testTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	expectedISO := "2024-01-15T10:30:45Z"

	result, err := converter.convertValue(testTime)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}

	if result != expectedISO {
		t.Errorf("convertValue() = %v, want %v", result, expectedISO)
	}
}

func TestDuckDBConverter_ConvertValue_SimpleMap(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test map[string]interface{} - common DuckDB MAP type
	input := map[string]interface{}{
		"name":   "John",
		"age":    30,
		"active": true,
	}

	result, err := converter.convertValue(input)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("convertValue() result should be map[string]interface{}, got %T", result)
	}

	if resultMap["name"] != "John" {
		t.Errorf("map name = %v, want John", resultMap["name"])
	}
	if resultMap["age"] != 30 {
		t.Errorf("map age = %v, want 30", resultMap["age"])
	}
	if resultMap["active"] != true {
		t.Errorf("map active = %v, want true", resultMap["active"])
	}
}

func TestDuckDBConverter_ConvertValue_MapWithNonStringKeys(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test map[interface{}]interface{} - DuckDB might return this
	input := map[interface{}]interface{}{
		"stringKey": "value1",
		42:          "value2",
		true:        "value3",
	}

	result, err := converter.convertValue(input)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("convertValue() result should be map[string]interface{}, got %T", result)
	}

	// All keys should be converted to strings
	if resultMap["stringKey"] != "value1" {
		t.Errorf("stringKey = %v, want value1", resultMap["stringKey"])
	}
	if resultMap["42"] != "value2" {
		t.Errorf("42 key = %v, want value2", resultMap["42"])
	}
	if resultMap["true"] != "value3" {
		t.Errorf("true key = %v, want value3", resultMap["true"])
	}
}

func TestDuckDBConverter_ConvertValue_SimpleArray(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test []interface{} - common DuckDB ARRAY type
	input := []interface{}{"apple", "banana", "cherry"}

	result, err := converter.convertValue(input)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok {
		t.Fatalf("convertValue() result should be []interface{}, got %T", result)
	}

	if len(resultSlice) != 3 {
		t.Errorf("array length = %d, want 3", len(resultSlice))
	}

	expected := []string{"apple", "banana", "cherry"}
	for i, exp := range expected {
		if resultSlice[i] != exp {
			t.Errorf("array[%d] = %v, want %v", i, resultSlice[i], exp)
		}
	}
}

func TestDuckDBConverter_ConvertValue_MixedArray(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test array with mixed types
	input := []interface{}{"text", 42, true, nil, 3.14}

	result, err := converter.convertValue(input)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok {
		t.Fatalf("convertValue() result should be []interface{}, got %T", result)
	}

	if len(resultSlice) != 5 {
		t.Errorf("array length = %d, want 5", len(resultSlice))
	}

	if resultSlice[0] != "text" {
		t.Errorf("array[0] = %v, want text", resultSlice[0])
	}
	if resultSlice[1] != 42 {
		t.Errorf("array[1] = %v, want 42", resultSlice[1])
	}
	if resultSlice[2] != true {
		t.Errorf("array[2] = %v, want true", resultSlice[2])
	}
	if resultSlice[3] != nil {
		t.Errorf("array[3] = %v, want nil", resultSlice[3])
	}
	if resultSlice[4] != 3.14 {
		t.Errorf("array[4] = %v, want 3.14", resultSlice[4])
	}
}

func TestDuckDBConverter_ConvertValue_NestedStructures(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test deeply nested map and array structures
	input := map[string]interface{}{
		"document": map[string]interface{}{
			"title":  "Complex Doc",
			"author": "John Doe",
			"tags":   []interface{}{"work", "important"},
			"metadata": map[string]interface{}{
				"created": "2024-01-15",
				"version": 1,
				"settings": map[string]interface{}{
					"public": false,
					"themes": []interface{}{"dark", "light"},
				},
			},
		},
		"sections": []interface{}{
			map[string]interface{}{
				"title":    "Introduction",
				"content":  "Hello world",
				"metadata": map[string]interface{}{"wordCount": 50},
			},
			map[string]interface{}{
				"title":   "Conclusion",
				"content": "The end",
				"tags":    []interface{}{"final"},
			},
		},
	}

	result, err := converter.convertValue(input)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("convertValue() result should be map[string]interface{}, got %T", result)
	}

	// Navigate to nested structure
	document, ok := resultMap["document"].(map[string]interface{})
	if !ok {
		t.Fatalf("document should be map[string]interface{}, got %T", resultMap["document"])
	}

	if document["title"] != "Complex Doc" {
		t.Errorf("document.title = %v, want Complex Doc", document["title"])
	}

	tags, ok := document["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags should be []interface{}, got %T", document["tags"])
	}
	if len(tags) != 2 || tags[0] != "work" || tags[1] != "important" {
		t.Errorf("tags = %v, want [work important]", tags)
	}

	metadata, ok := document["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata should be map[string]interface{}, got %T", document["metadata"])
	}

	settings, ok := metadata["settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("settings should be map[string]interface{}, got %T", metadata["settings"])
	}

	if settings["public"] != false {
		t.Errorf("settings.public = %v, want false", settings["public"])
	}

	themes, ok := settings["themes"].([]interface{})
	if !ok {
		t.Fatalf("themes should be []interface{}, got %T", settings["themes"])
	}
	if len(themes) != 2 || themes[0] != "dark" || themes[1] != "light" {
		t.Errorf("themes = %v, want [dark light]", themes)
	}

	// Test sections array
	sections, ok := resultMap["sections"].([]interface{})
	if !ok {
		t.Fatalf("sections should be []interface{}, got %T", resultMap["sections"])
	}
	if len(sections) != 2 {
		t.Errorf("sections length = %d, want 2", len(sections))
	}

	section1, ok := sections[0].(map[string]interface{})
	if !ok {
		t.Fatalf("sections[0] should be map[string]interface{}, got %T", sections[0])
	}
	if section1["title"] != "Introduction" {
		t.Errorf("section1.title = %v, want Introduction", section1["title"])
	}
}

func TestDuckDBConverter_ConvertValue_EmptyContainers(t *testing.T) {
	converter := NewDuckDBConverter()

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"empty map", map[string]interface{}{}, map[string]interface{}{}},
		{"empty slice", []interface{}{}, []interface{}{}},
		{"empty array", [0]interface{}{}, []interface{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.convertValue(tt.input)
			if err != nil {
				t.Errorf("convertValue() error = %v", err)
				return
			}

			// Use deep equal for container types
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("convertValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDuckDBConverter_ConvertValue_Pointers(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test nil pointer
	var nilPtr *string
	result, err := converter.convertValue(nilPtr)
	if err != nil {
		t.Errorf("convertValue() error = %v", err)
	}
	if result != nil {
		t.Errorf("convertValue() = %v, want nil", result)
	}

	// Test valid pointer
	value := "hello"
	ptr := &value
	result, err = converter.convertValue(ptr)
	if err != nil {
		t.Errorf("convertValue() error = %v", err)
	}
	if result != "hello" {
		t.Errorf("convertValue() = %v, want hello", result)
	}
}

// Row-level conversion tests

func TestDuckDBConverter_ConvertRow_SimpleRow(t *testing.T) {
	converter := NewDuckDBConverter()

	input := map[string]interface{}{
		"id":       1,
		"name":     "John",
		"metadata": map[string]interface{}{"role": "admin"},
		"tags":     []interface{}{"user", "active"},
	}

	result, err := converter.ConvertRow(input)
	if err != nil {
		t.Fatalf("ConvertRow() error = %v", err)
	}

	if result["id"] != 1 {
		t.Errorf("id = %v, want 1", result["id"])
	}
	if result["name"] != "John" {
		t.Errorf("name = %v, want John", result["name"])
	}

	metadata, ok := result["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata should be map[string]interface{}, got %T", result["metadata"])
	}
	if metadata["role"] != "admin" {
		t.Errorf("metadata.role = %v, want admin", metadata["role"])
	}

	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags should be []interface{}, got %T", result["tags"])
	}
	if len(tags) != 2 || tags[0] != "user" || tags[1] != "active" {
		t.Errorf("tags = %v, want [user active]", tags)
	}
}

func TestDuckDBConverter_ConvertRow_ErrorHandling(t *testing.T) {
	converter := NewDuckDBConverter()

	// Test row with values that should trigger fallback behavior
	input := map[string]interface{}{
		"good_value": "normal",
		"channel":    make(chan int), // channels can't be JSON marshaled, but won't error in conversion
	}

	result, err := converter.ConvertRow(input)
	if err != nil {
		t.Fatalf("ConvertRow() error = %v", err)
	}

	// Should have fallback behavior for problematic types
	if result["good_value"] != "normal" {
		t.Errorf("good_value = %v, want normal", result["good_value"])
	}

	// Channel should pass through (since reflection doesn't error on it)
	// The actual error would occur during JSON marshaling later
	channelValue := result["channel"]
	if channelValue == nil {
		t.Errorf("channel value should not be nil")
	}
}

// Results-level conversion tests

func TestDuckDBConverter_ConvertResults_EmptyResults(t *testing.T) {
	converter := NewDuckDBConverter()

	input := []map[string]interface{}{}
	result, err := converter.ConvertResults(input)
	if err != nil {
		t.Fatalf("ConvertResults() error = %v", err)
	}

	if len(result) != 0 {
		t.Errorf("ConvertResults() length = %d, want 0", len(result))
	}
}

func TestDuckDBConverter_ConvertResults_MultipleRows(t *testing.T) {
	converter := NewDuckDBConverter()

	input := []map[string]interface{}{
		{
			"id":       1,
			"name":     "Alice",
			"metadata": map[string]interface{}{"department": "engineering"},
			"skills":   []interface{}{"go", "sql"},
		},
		{
			"id":       2,
			"name":     "Bob",
			"metadata": map[string]interface{}{"department": "design"},
			"skills":   []interface{}{"figma", "css"},
		},
	}

	result, err := converter.ConvertResults(input)
	if err != nil {
		t.Fatalf("ConvertResults() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ConvertResults() length = %d, want 2", len(result))
	}

	// Check first row
	row1 := result[0]
	if row1["name"] != "Alice" {
		t.Errorf("row1.name = %v, want Alice", row1["name"])
	}

	metadata1, ok := row1["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("row1.metadata should be map[string]interface{}, got %T", row1["metadata"])
	}
	if metadata1["department"] != "engineering" {
		t.Errorf("row1.metadata.department = %v, want engineering", metadata1["department"])
	}

	skills1, ok := row1["skills"].([]interface{})
	if !ok {
		t.Fatalf("row1.skills should be []interface{}, got %T", row1["skills"])
	}
	if len(skills1) != 2 || skills1[0] != "go" || skills1[1] != "sql" {
		t.Errorf("row1.skills = %v, want [go sql]", skills1)
	}

	// Check second row
	row2 := result[1]
	if row2["name"] != "Bob" {
		t.Errorf("row2.name = %v, want Bob", row2["name"])
	}

	skills2, ok := row2["skills"].([]interface{})
	if !ok {
		t.Fatalf("row2.skills should be []interface{}, got %T", row2["skills"])
	}
	if len(skills2) != 2 || skills2[0] != "figma" || skills2[1] != "css" {
		t.Errorf("row2.skills = %v, want [figma css]", skills2)
	}
}

func TestDuckDBConverter_ConvertResults_RealWorldNoteStructure(t *testing.T) {
	converter := NewDuckDBConverter()

	// Simulate actual DuckDB markdown extension results
	input := []map[string]interface{}{
		{
			"filepath": "/notebook/project.md",
			"content":  "# Project Alpha\n\nPlanning document...",
			"metadata": map[interface{}]interface{}{ // DuckDB might return map[interface{}]interface{}
				"title":    "Project Alpha",
				"status":   "active",
				"priority": 1,
				"tags":     []interface{}{"project", "planning"},
				"team": map[string]interface{}{
					"lead":    "alice",
					"members": []interface{}{"bob", "charlie"},
				},
			},
		},
		{
			"filepath": "/notebook/meeting.md",
			"content":  "# Meeting Notes\n\nDiscussed next steps...",
			"metadata": map[string]interface{}{
				"title": "Weekly Standup",
				"date":  "2024-01-15",
				"attendees": []interface{}{
					map[string]interface{}{"name": "Alice", "role": "lead"},
					map[string]interface{}{"name": "Bob", "role": "dev"},
				},
			},
		},
	}

	result, err := converter.ConvertResults(input)
	if err != nil {
		t.Fatalf("ConvertResults() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ConvertResults() length = %d, want 2", len(result))
	}

	// Check first note (project.md)
	note1 := result[0]
	if note1["filepath"] != "/notebook/project.md" {
		t.Errorf("note1.filepath = %v, want /notebook/project.md", note1["filepath"])
	}

	metadata1, ok := note1["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("note1.metadata should be map[string]interface{}, got %T", note1["metadata"])
	}

	if metadata1["title"] != "Project Alpha" {
		t.Errorf("note1.metadata.title = %v, want Project Alpha", metadata1["title"])
	}

	tags1, ok := metadata1["tags"].([]interface{})
	if !ok {
		t.Fatalf("note1.metadata.tags should be []interface{}, got %T", metadata1["tags"])
	}
	if len(tags1) != 2 || tags1[0] != "project" || tags1[1] != "planning" {
		t.Errorf("note1.metadata.tags = %v, want [project planning]", tags1)
	}

	team1, ok := metadata1["team"].(map[string]interface{})
	if !ok {
		t.Fatalf("note1.metadata.team should be map[string]interface{}, got %T", metadata1["team"])
	}

	if team1["lead"] != "alice" {
		t.Errorf("note1.metadata.team.lead = %v, want alice", team1["lead"])
	}

	members1, ok := team1["members"].([]interface{})
	if !ok {
		t.Fatalf("note1.metadata.team.members should be []interface{}, got %T", team1["members"])
	}
	if len(members1) != 2 || members1[0] != "bob" || members1[1] != "charlie" {
		t.Errorf("note1.metadata.team.members = %v, want [bob charlie]", members1)
	}

	// Check second note (meeting.md)
	note2 := result[1]
	metadata2, ok := note2["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("note2.metadata should be map[string]interface{}, got %T", note2["metadata"])
	}

	attendees2, ok := metadata2["attendees"].([]interface{})
	if !ok {
		t.Fatalf("note2.metadata.attendees should be []interface{}, got %T", metadata2["attendees"])
	}
	if len(attendees2) != 2 {
		t.Errorf("note2.metadata.attendees length = %d, want 2", len(attendees2))
	}

	alice, ok := attendees2[0].(map[string]interface{})
	if !ok {
		t.Fatalf("attendees[0] should be map[string]interface{}, got %T", attendees2[0])
	}
	if alice["name"] != "Alice" || alice["role"] != "lead" {
		t.Errorf("alice = %v, want {name: Alice, role: lead}", alice)
	}

	bob, ok := attendees2[1].(map[string]interface{})
	if !ok {
		t.Fatalf("attendees[1] should be map[string]interface{}, got %T", attendees2[1])
	}
	if bob["name"] != "Bob" || bob["role"] != "dev" {
		t.Errorf("bob = %v, want {name: Bob, role: dev}", bob)
	}
}

func TestDuckDBConverter_ConvertResults_PerformanceLargeDataset(t *testing.T) {
	converter := NewDuckDBConverter()

	// Create a large dataset to test performance
	input := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = map[string]interface{}{
			"id":       i,
			"title":    fmt.Sprintf("Note %d", i),
			"metadata": map[string]interface{}{"author": "user", "version": i},
			"tags":     []interface{}{"work", "important"},
		}
	}

	start := time.Now()
	result, err := converter.ConvertResults(input)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("ConvertResults() error = %v", err)
	}

	// Performance target: <50ms for 1000 rows
	if duration > 50*time.Millisecond {
		t.Errorf("ConvertResults() took %v, want <50ms for 1000 rows", duration)
	}

	if len(result) != 1000 {
		t.Errorf("ConvertResults() length = %d, want 1000", len(result))
	}

	// Verify a sample of the data
	if result[0]["id"] != 0 {
		t.Errorf("result[0].id = %v, want 0", result[0]["id"])
	}
	if result[999]["id"] != 999 {
		t.Errorf("result[999].id = %v, want 999", result[999]["id"])
	}
}

func BenchmarkDuckDBConverter_ConvertValue_SimpleMap(b *testing.B) {
	converter := NewDuckDBConverter()
	input := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.convertValue(input)
	}
}

func BenchmarkDuckDBConverter_ConvertValue_SimpleArray(b *testing.B) {
	converter := NewDuckDBConverter()
	input := []interface{}{"item1", "item2", "item3", 42, true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.convertValue(input)
	}
}

func BenchmarkDuckDBConverter_ConvertResults_MediumDataset(b *testing.B) {
	converter := NewDuckDBConverter()

	// Medium dataset: 100 rows
	input := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		input[i] = map[string]interface{}{
			"id":       i,
			"title":    fmt.Sprintf("Note %d", i),
			"metadata": map[string]interface{}{"author": "user", "version": i},
			"tags":     []interface{}{"work", "important"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.ConvertResults(input)
	}
}
