#!/usr/bin/env bats
# OpenNotes Core Smoke Tests - Essential functionality verification

setup() {
    export TEST_DIR="$(mktemp -d)"
    export HOME_BACKUP="$HOME"
    export HOME="$TEST_DIR"
    export OPENNOTES_CONFIG="$TEST_DIR/.config/opennotes/config.json"
    
    # Ensure binary exists (check from project root)
    if [[ ! -f "../../dist/opennotes" ]]; then
        cd ../.. && mise run build && cd tests/e2e
    fi
    
    export PATH="$(pwd)/../../dist:$PATH"
}

teardown() {
    rm -rf "$TEST_DIR"
    export HOME="$HOME_BACKUP"
}

@test "Core workflow: init → create notebook → add note → list → search → SQL query" {
    # Initialize OpenNotes
    run opennotes init
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "OpenNotes initialized" ]]
    
    # Create a test notebook
    notebook_dir="$TEST_DIR/work-notes"
    run opennotes notebook create "$notebook_dir" --name "Work Notes"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Created notebook" ]]
    [[ -d "$notebook_dir" ]]
    [[ -f "$notebook_dir/.opennotes.json" ]]
    
    # Add a note with content
    run opennotes --notebook "$notebook_dir" notes add "project-alpha.md" --title "Project Alpha"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Created note" ]]
    [[ -f "$notebook_dir/.notes/project-alpha.md" ]]
    
    # Add content to the note for testing
    echo -e "# Project Alpha\n\nThis is a **high priority** project.\n\nTasks:\n- Design phase\n- Implementation\n- Testing" > "$notebook_dir/.notes/project-alpha.md"
    
    # Add another note
    run opennotes --notebook "$notebook_dir" notes add "meeting-notes.md" --title "Weekly Meeting"
    [[ "$status" -eq 0 ]]
    echo -e "# Weekly Meeting\n\nDiscussed project timeline and deliverables.\n\nAction items:\n- Review requirements\n- Schedule follow-up" > "$notebook_dir/.notes/meeting-notes.md"
    
    # List notes
    run opennotes --notebook "$notebook_dir" notes list
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "project-alpha.md" ]]
    [[ "$output" =~ "meeting-notes.md" ]]
    
    # Search for specific content
    run opennotes --notebook "$notebook_dir" notes search "priority"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "project-alpha.md" ]]
    [[ ! "$output" =~ "meeting-notes.md" ]]
    
    # Basic SQL query to find all markdown files
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "project-alpha.md" ]]
    [[ "$output" =~ "meeting-notes.md" ]]
    
    # SQL query with content filtering
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%timeline%'"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "meeting-notes.md" ]]
    [[ ! "$output" =~ "project-alpha.md" ]]
    
    # Remove a note
    run opennotes --notebook "$notebook_dir" notes remove "meeting-notes.md" --force
    [[ "$status" -eq 0 ]]
    [[ ! -f "$notebook_dir/.notes/meeting-notes.md" ]]
    
    # Verify only one note remains
    run opennotes --notebook "$notebook_dir" notes list
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "project-alpha.md" ]]
    [[ ! "$output" =~ "meeting-notes.md" ]]
}

@test "CLI help system provides useful information" {
    # Main help
    run opennotes --help
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "OpenNotes" ]]
    [[ "$output" =~ "markdown-based notes" ]]
    [[ "$output" =~ "Available Commands" ]]
    
    # Notebook subcommands help
    run opennotes notebook --help
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "managing notebooks" ]]
    [[ "$output" =~ "create" ]]
    [[ "$output" =~ "list" ]]
    
    # Notes subcommands help
    run opennotes notes --help
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "managing notes" ]]
    [[ "$output" =~ "add" ]]
    [[ "$output" =~ "search" ]]
    [[ "$output" =~ "remove" ]]
    
    # SQL help in search command
    run opennotes notes search --help
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "SQL" ]]
    [[ "$output" =~ "DuckDB" ]]
    [[ "$output" =~ "read_markdown" ]]
    [[ "$output" =~ "Path traversal" ]]
}

@test "Error handling works correctly" {
    # After init, should work
    opennotes init
    
    run opennotes notebook list
    [[ "$status" -eq 0 ]]
    
    # Error with invalid notebook path
    run opennotes --notebook "/nonexistent/path" notes list
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "no such file or directory" || "$output" =~ "error" || "$output" =~ "does not exist" ]]
    
    # Error with invalid SQL (malformed syntax)
    notebook_dir="$TEST_DIR/error-test"
    opennotes notebook create "$notebook_dir" --name "Error Test"
    
    run opennotes --notebook "$notebook_dir" notes search --sql "INVALID SQL SYNTAX"
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "error" || "$output" =~ "failed" ]]
}

@test "SQL security features work" {
    opennotes init
    notebook_dir="$TEST_DIR/security-test"
    opennotes notebook create "$notebook_dir" --name "Security Test"
    echo "# Safe Note" > "$notebook_dir/.notes/safe.md"
    
    # Test that path traversal attempts are handled (should error, not succeed)
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT content FROM read_markdown('../../../etc/passwd')"
    [[ "$status" -ne 0 ]]
    # Error message indicates the path was rejected
    [[ "$output" =~ "error" || "$output" =~ "Error" ]]
    
    # Valid query within notebook should work
    run opennotes --notebook "$notebook_dir" notes search --sql "SELECT file_path FROM read_markdown('*.md', include_filepath:=true)"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "safe.md" ]]
}