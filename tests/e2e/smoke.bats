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
    
    # Ensure we have the built binary
    if [[ ! -f "dist/opennotes" ]]; then
        mise run build
    fi
    
    # Add dist to PATH for this test
    export PATH="$(pwd)/dist:$PATH"
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
    
    # Check config file exists in HOME directory (not test directory)
    [[ -f "$HOME_BACKUP/.config/opennotes/config.json" ]]
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

# Test 6: SQL querying functionality
@test "SQL querying with DuckDB" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "sql-test")
    
    # Create some notes
    create_test_note "$notebook_dir" "task1.md" "# Task 1\n\nPriority: High\nStatus: TODO"
    create_test_note "$notebook_dir" "task2.md" "# Task 2\n\nPriority: Low\nStatus: DONE"
    
    # Test basic SQL query
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "task1.md" ]]
    [[ "$output" =~ "task2.md" ]]
    
    # Test markdown content querying
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%High%'"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "task1.md" ]]
    [[ ! "$output" =~ "task2.md" ]]
}

# Test 7: SQL security (path traversal protection)
@test "SQL security prevents path traversal" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "security-test")
    create_test_note "$notebook_dir" "safe.md" "# Safe Note"
    
    # These should be blocked by path traversal protection
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT content FROM read_markdown('../../../etc/passwd')"
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "blocked due to path traversal" ]]
    
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT content FROM read_markdown('/etc/passwd')"
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "blocked due to path traversal" ]]
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
    run opennotes notebook register "info-test" "$notebook_dir"
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
    # Test without initialization
    run opennotes notebook list
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "config" || "$output" =~ "not found" || "$output" =~ "initialize" ]]
    
    # Test with invalid notebook
    opennotes init
    run opennotes --notebook "/nonexistent/path" notes list
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "notebook not found" || "$output" =~ "error" ]]
    
    # Test invalid SQL
    notebook_dir=$(create_test_notebook "error-test")
    run opennotes --notebook "$notebook_dir" notes list --sql "INVALID SQL SYNTAX"
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "error" || "$output" =~ "syntax" ]]
}

# Test 11: Complex SQL with Common Table Expressions (CTEs)
@test "Advanced SQL features work correctly" {
    # Setup
    opennotes init
    notebook_dir=$(create_test_notebook "advanced-sql-test")
    
    # Create notes with frontmatter
    cat > "$notebook_dir/project1.md" << 'EOF'
---
title: Project Alpha
priority: high
status: active
tags: [work, important]
---
# Project Alpha
This is a high priority project.
EOF

    cat > "$notebook_dir/project2.md" << 'EOF'
---
title: Project Beta
priority: low
status: completed
tags: [work, archive]
---
# Project Beta
This project is now completed.
EOF

    # Test CTE query
    run opennotes --notebook "$notebook_dir" notes list --sql "
    WITH high_priority AS (
        SELECT filename, frontmatter
        FROM notes
        WHERE frontmatter LIKE '%priority: high%'
    )
    SELECT filename FROM high_priority
    "
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
    run opennotes notebook create "workflow-test" "$TEST_DIR/workflow-test"
    [[ "$status" -eq 0 ]]
    
    # Add multiple notes
    notebook_dir="$TEST_DIR/workflow-test"
    run opennotes --notebook "$notebook_dir" notes add "meeting-notes.md"
    [[ "$status" -eq 0 ]]
    run opennotes --notebook "$notebook_dir" notes add "project-plan.md"  
    [[ "$status" -eq 0 ]]
    
    # List all notes
    run opennotes --notebook "$notebook_dir" notes list
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "meeting-notes.md" ]]
    [[ "$output" =~ "project-plan.md" ]]
    
    # Search notes
    # Add content to one note first
    echo "# Meeting Notes\nDiscussed project timeline." > "$notebook_dir/meeting-notes.md"
    
    run opennotes --notebook "$notebook_dir" notes search "timeline"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "meeting-notes.md" ]]
    
    # SQL query to count notes
    run opennotes --notebook "$notebook_dir" notes list --sql "SELECT COUNT(*) as total FROM notes"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "2" ]]
    
    # Get notebook info
    run opennotes --notebook "$notebook_dir" notebook
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "workflow-test" ]]
}