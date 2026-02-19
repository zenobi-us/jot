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

@test "Core workflow: init → create notebook → add note → list → search → query" {
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
    run opennotes --notebook "$notebook_dir" notes add "Project Alpha" "project-alpha.md"
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Created note" ]]
    [[ -f "$notebook_dir/.notes/project-alpha.md" ]]
    
    # Add content to the note for testing
    echo -e "# Project Alpha\n\nThis is a **high priority** project.\n\nTasks:\n- Design phase\n- Implementation\n- Testing" > "$notebook_dir/.notes/project-alpha.md"
    
    # Add another note
    run opennotes --notebook "$notebook_dir" notes add "Weekly Meeting" "meeting-notes.md"
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
    
    # Boolean query to find a specific note by path
    run opennotes --notebook "$notebook_dir" notes search query --and path=project-alpha.md
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "project-alpha.md" ]]
    [[ ! "$output" =~ "meeting-notes.md" ]]
    
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
    
    # Search help includes fuzzy + boolean query info
    run opennotes notes search --help
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "Fuzzy" ]]
    [[ "$output" =~ "BOOLEAN" ]]
    [[ "$output" =~ "query" ]]
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
    
    # Error with invalid query field
    notebook_dir="$TEST_DIR/error-test"
    opennotes notebook create "$notebook_dir" --name "Error Test"
    
    run opennotes --notebook "$notebook_dir" notes search query --and data.unknown=foo
    [[ "$status" -ne 0 ]]
    [[ "$output" =~ "invalid field" || "$output" =~ "allowed" ]]
}

@test "Path filtering works with boolean queries" {
    opennotes init
    notebook_dir="$TEST_DIR/security-test"
    opennotes notebook create "$notebook_dir" --name "Security Test"
    echo "# Safe Note" > "$notebook_dir/.notes/safe.md"
    echo "# Other Note" > "$notebook_dir/.notes/other.md"
    
    # Path query should return only matching note
    run opennotes --notebook "$notebook_dir" notes search query --and path=safe.md
    [[ "$status" -eq 0 ]]
    [[ "$output" =~ "safe.md" ]]
    [[ ! "$output" =~ "other.md" ]]
}