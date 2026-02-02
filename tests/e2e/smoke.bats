#!/usr/bin/env bats
# OpenNotes E2E Smoke Tests
# Tests critical user journeys with real CLI binary

# Setup and teardown
setup() {
    # Create temporary directory for test
    export TEST_DIR="$(mktemp -d)"
    export HOME_BACKUP="$HOME"
    export HOME="$TEST_DIR"
    export OPENNOTES_CONFIG="$TEST_DIR/.config/opennotes/config.json"
    
    # Ensure we have the built binary (build from project root)
    if [[ ! -f "../../dist/opennotes" ]]; then
        cd ../.. && mise run build && cd tests/e2e
    fi
    
    # Add dist to PATH for this test
    export PATH="$(pwd)/../../dist:$PATH"
}

teardown() {
    # Clean up
    rm -rf "$TEST_DIR"
    export HOME="$HOME_BACKUP"
}

# Helper function to create a test notebook
create_test_notebook() {
    local name="$1"
    local dir="$TEST_DIR/$name"
    mkdir -p "$dir/.notes"
    echo '{"name":"'"$name"'","path":"'"$dir"'"}' > "$dir/.opennotes.json"
    echo "$dir"
}

# Helper function to create a test note
create_test_note() {
    local notebook_dir="$1"
    local filename="$2"
    local content="${3:-# Test Note\n\nThis is a test note.}"
    echo -e "$content" > "$notebook_dir/$filename"
}

# Test 1: Help shows correct information
@test "CLI shows help" {
    run opennotes --help
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "OpenNotes" ]]
    [[ "$output" =~ "CLI tool for managing your markdown-based notes" ]]
    [[ "$output" =~ "Available Commands" ]]
}

# Test 2: Initialize configuration
@test "Initialize configuration" {
    run opennotes init
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "OpenNotes initialized" ]]
    
    # Check config file exists in configured path
    [[ -f "$OPENNOTES_CONFIG" ]]
}

# Test 3: Create and list notebooks
@test "Create and list notebooks" {
    # Initialize first
    run opennotes init
    [[ "$status" -eq 0 ]]
    
    # Create a notebook
    run opennotes notebook create "$TEST_DIR/test-notebook" --name "test-notebook"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Created notebook" ]]
    
    # Check notebook directory and config
    [[ -d "$TEST_DIR/test-notebook" ]]
    [[ -f "$TEST_DIR/test-notebook/.opennotes.json" ]]
}

# Test 4: Add and list notes
@test "Add and list notes" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "notes-test")
    
    # Create a note
    run opennotes --notebook "$notebook_dir" notes add test-note.md --title "Test Note"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Created note" ]]
    
    # Check note exists (CLI creates in notebook root, not .notes subdirectory)
    [[ -f "$notebook_dir/test-note.md" ]]
    
    # List notes
    run opennotes --notebook "$notebook_dir" notes list
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "test-note.md" ]]
}

# Test 5: Search notes with content
@test "Search notes functionality" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "search-test")
    
    # Create notes with different content
    create_test_note "$notebook_dir" "note1.md" "# First Note\n\nThis contains the word example."
    create_test_note "$notebook_dir" "note2.md" "# Second Note\n\nThis is about testing."
    create_test_note "$notebook_dir" "note3.md" "# Third Note\n\nAnother example here."
    
    # Search for notes
    run opennotes --notebook "$notebook_dir" notes search "example"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "note1.md" ]]
    [[ "$output" =~ "note3.md" ]]
    [[ ! "$output" =~ "note2.md" ]]
}

# Test 6: Boolean query functionality
@test "Boolean query filtering" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "query-test")
    
    # Create some notes
    create_test_note "$notebook_dir" "task1.md" "# Task 1\n\nPriority: High\nStatus: TODO"
    create_test_note "$notebook_dir" "task2.md" "# Task 2\n\nPriority: Low\nStatus: DONE"
    
    # Exact path match
    run opennotes --notebook "$notebook_dir" notes search query --and path=task1.md
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "task1.md" ]]
    [[ ! "$output" =~ "task2.md" ]]
    
    # Wildcard path match
    run opennotes --notebook "$notebook_dir" notes search query --and path=task*.md
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "task1.md" ]]
    [[ "$output" =~ "task2.md" ]]
}

# Test 7: Fuzzy search basics
@test "Fuzzy search finds close matches" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "fuzzy-test")
    create_test_note "$notebook_dir" "meeting-notes.md" "# Meeting Notes\n\nDiscussed roadmap"
    
    run opennotes --notebook "$notebook_dir" notes search --fuzzy "metng"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "meeting-notes.md" ]]
}

# Test 8: Note removal
@test "Remove notes functionality" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "remove-test")
    create_test_note "$notebook_dir" "remove-me.md" "# Remove Me\n\nThis note will be deleted."
    
    # Verify note exists
    [[ -f "$notebook_dir/remove-me.md" ]]
    
    # Remove note (with force to skip confirmation)
    run opennotes --notebook "$notebook_dir" notes remove "remove-me.md" --force
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Removed" || "$output" =~ "removed" ]]
    
    # Verify note is gone
    [[ ! -f "$notebook_dir/remove-me.md" ]]
}

# Test 9: Notebook registration and info
@test "Notebook registration and info" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "info-test")
    
    # Register notebook
    run opennotes notebook register "$notebook_dir"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Registered notebook" ]]
    
    # Get notebook info
    run opennotes --notebook "$notebook_dir" notebook
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "info-test" ]]
    [[ "$output" =~ "$notebook_dir" ]]
}

# Test 10: Error handling
@test "Proper error handling" {
    # Ensure clean state - remove any existing config
    rm -rf "$(dirname "$OPENNOTES_CONFIG")"
    # Test without initialization
    run opennotes notebook list
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "config" || "$output" =~ "not found" || "$output" =~ "initialize" ]]
    
    # Test with invalid notebook
    opennotes init
    run opennotes --notebook "/nonexistent/path" notes list
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "notebook not found" || "$output" =~ "Error" || "$output" =~ "error" ]]
    
    # Test invalid query field
    notebook_dir=$(create_test_notebook "error-test")
    run opennotes --notebook "$notebook_dir" notes search query --and data.unknown=oops
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "invalid field" || "$output" =~ "allowed" ]]
}

# Test 11: Advanced boolean queries
@test "Advanced boolean queries work correctly" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "advanced-query-test")
    
    # Create notes
    create_test_note "$notebook_dir" "project1.md" "# Project Alpha\n\nHigh priority project."
    create_test_note "$notebook_dir" "project2.md" "# Project Beta\n\nLow priority project."
    
    # Use AND + NOT to filter
    run opennotes --notebook "$notebook_dir" notes search query --and path=project*.md --not path=project2.md
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "project1.md" ]]
    [[ ! "$output" =~ "project2.md" ]]
}

# Test 12: Full user journey
@test "Complete user workflow" {
    # Complete workflow test
    run opennotes init
    [[ "$status" -eq 0 ]]
    
    # Create notebook
    run opennotes notebook create "$TEST_DIR/workflow-test" --name "workflow-test"
    [[ "$status" -eq 0 ]]
    
    # Add multiple notes
    notebook_dir="$TEST_DIR/workflow-test"
    run opennotes --notebook "$notebook_dir" notes add "Meeting Notes" "meeting-notes.md"
    [[ "$status" -eq 0 ]]
    run opennotes --notebook "$notebook_dir" notes add "Project Plan" "project-plan.md"  
    [[ "$status" -eq 0 ]]
    
    # List all notes
    run opennotes --notebook "$notebook_dir" notes list
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Meeting Notes" ]]
    [[ "$output" =~ "Project Plan" ]]
    
    # Search notes
    # Add content to one note first  
    echo -e "# Meeting Notes\nDiscussed project timeline." > "$notebook_dir/.notes/meeting-notes.md"
    
    run opennotes --notebook "$notebook_dir" notes search "timeline"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "meeting-notes.md" ]]
    
    # List all notes via search output
    run opennotes --notebook "$notebook_dir" notes search
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Found 2 note(s)" ]]
    
    # Get notebook info
    run opennotes --notebook "$notebook_dir" notebook
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "workflow-test" ]]
}