---
id: ca68615f-01
title: Task 1 - Core Implementation (Title, Data Flags, Path Resolution)
created_at: 2026-01-24T23:45:00+10:30
updated_at: 2026-01-24T23:45:00+10:30
status: in-progress
epic_id: 3e01c563
phase_id: ca68615f
assigned_to: claude-20260124-session2
---

# Task 1: Core Implementation

## Objective

Implement the core functionality for enhanced note creation:
1. Positional `<title>` argument parsing
2. `--data field=value` flag parsing (repeatable)
3. Path resolution auto-detection
4. Title slugification
5. Frontmatter generation

## Steps

### Step 1: Update Command Signature ✅
- [x] Modify `cmd/notes_add.go` to accept positional args
- [x] Change `Args: cobra.ExactArgs(0)` to `Args: cobra.MaximumNArgs(2)`
- [x] Update command `Use` string to show new syntax

### Step 2: Add --data Flag ✅ COMPLETE
- [x] Add `StringArrayVar` for `--data` flag
- [x] Make flag repeatable
- [x] Add help text with examples

### Step 3: Implement Argument Parsing ✅ COMPLETE
- [x] Create `parseArguments()` function in command
- [x] Handle both old style (`--title`) and new style (positional)
- [x] Return title and path separately
- [x] Add error handling for conflicting arguments
- [x] Use `cmd.Flags().Changed("title")` to detect flag usage

### Step 4: Implement Data Flag Parser ✅ COMPLETE
- [x] Create `ParseDataFlags()` in `internal/services/note.go` (exported)
- [x] Parse `field=value` format
- [x] Support repeated fields (create arrays)
- [x] Validate field names (check for empty)
- [x] Return `map[string]interface{}`

### Step 5: Title Slugification ✅ COMPLETE (already existed)
- [x] Use existing `core.Slugify()` function
- [x] Converts to lowercase
- [x] Replaces spaces with hyphens
- [x] Removes special characters
- [x] Keeps alphanumeric and hyphens only

### Step 6: Implement Path Resolution ✅ COMPLETE
- [x] Create `ResolvePath()` in `internal/services/note.go` (exported)
- [x] Handle no path (default to root + slugified title)
- [x] Handle folder paths (ends with `/`)
- [x] Handle full filepaths with `.md`
- [x] Handle filepaths without extension (auto-add `.md`)
- [x] Path validation handled by notebook service

### Step 7: Implement Frontmatter Generation ✅ COMPLETE
- [x] Create `generateFrontmatter()` in command
- [x] Start with `created` timestamp
- [x] Add `title` field (if not empty)
- [x] Merge custom `--data` fields
- [x] Handle repeated fields (arrays)
- [x] Serialize to YAML

### Step 8: Write Tests ✅ COMPLETE
- [x] Test `ParseDataFlags()` with valid input (10 test cases)
- [x] Test `ParseDataFlags()` with invalid input
- [x] Test `ParseDataFlags()` with repeated fields
- [x] Test `ResolvePath()` with all scenarios (5 test cases)
- [x] All existing tests pass (161+ tests)
- [x] Argument parsing tested via e2e tests

## Expected Outcome

After completing this task:
- ✅ Positional title argument working
- ✅ `--data` flags parsing correctly
- ✅ Path resolution handling all cases
- ✅ Title slugification producing safe names
- ✅ Frontmatter generation with custom fields
- ✅ All unit tests passing
- ✅ No regressions in existing tests

## Actual Outcome

**Status**: ✅ COMPLETE  
**Started**: 2026-01-24T23:45:00+10:30  
**Completed**: 2026-01-24T23:54:00+10:30  
**Duration**: ~1 hour

### Implementation Progress

#### All Steps Completed ✅

**Files Modified**:
1. `cmd/notes_add.go` - Complete rewrite of argument parsing and note creation logic
2. `internal/services/note.go` - Added `ParseDataFlags()` and `ResolvePath()` functions
3. `internal/services/note_test.go` - Added 15 new test cases

**Features Implemented**:
- ✅ Positional `<title> [path]` syntax (new style)
- ✅ Backward compatibility with `--title` flag (deprecated)
- ✅ `--data field=value` flags for frontmatter (repeatable)
- ✅ Repeated fields create arrays
- ✅ Path resolution auto-detection (file vs folder)
- ✅ Stdin content integration (highest priority)
- ✅ Template support (second priority)
- ✅ Default content generation (third priority)
- ✅ Deprecation warnings for old syntax

**Test Results**:
- ✅ All 161+ existing tests pass (zero regressions)
- ✅ 15 new unit tests added (ParseDataFlags, ResolvePath)
- ✅ 4 e2e tests pass (empty title, existing file, directory, long filename)
- ✅ Manual testing confirms all features work

### Challenges Encountered

1. **Challenge**: E2e test failure for empty title
   - **Cause**: Argument parsing logic didn't differentiate between `--title` flag not provided vs `--title ""` provided
   - **Solution**: Used `cmd.Flags().Changed("title")` to detect if flag was actually provided
   
2. **Challenge**: Path resolution created wrong filenames
   - **Cause**: When using `--title`, args[0] was being treated as title instead of path
   - **Solution**: Simplified logic - if `titleFlagProvided`, then args[0] is always path; otherwise args[0] is title

3. **Challenge**: Multiple test failures for file operations
   - **Cause**: Validation logic was too strict, rejecting valid old-style syntax
   - **Solution**: Removed conflicting title check when using `--title` flag

### Lessons Learned

1. **Use Cobra's `Changed()` method**: Essential for differentiating between "flag not provided" vs "flag provided with empty value"
2. **Test-driven development works**: Writing tests first caught the argument parsing issues early
3. **Backward compatibility is tricky**: Need to carefully consider how old and new syntax interact
4. **Reuse existing functions**: The `core.Slugify()` function already existed and worked perfectly
5. **Simple is better**: The final parseArguments() function is much simpler than initial attempts

## Notes

- Following TDD approach: Write tests before implementation where possible
- All business logic goes in `internal/services/note.go`
- Command file is thin orchestration only
- Using established patterns from existing codebase
- Referencing `.memory/spec-ca68615f-note-creation-enhancement.md` for details

---

**Status**: 🔄 IN PROGRESS  
**Next Step**: Implement --data flag parsing  
**Estimated Completion**: 2-3 hours from start
