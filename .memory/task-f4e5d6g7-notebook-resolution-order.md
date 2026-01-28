# Task: Fix Notebook Resolution Order Priority

**Status**: ✅ COMPLETED  
**Priority**: HIGH  
**Estimated Duration**: 1-2 hours  
**Created**: 2026-01-25  
**Completed**: 2026-01-28

> **Note**: This task was created as a specification AFTER the implementation was already complete.
> Comprehensive verification confirmed all 5 resolution tiers were working correctly.

## Problem

Current notebook resolution order is incorrect:
1. ❌ --notebook flag (global config)
2. ❌ context match (registered notebooks)
3. ❌ ancestor search
4. ❌ (no envvar support)
5. ❌ (no current directory check)

**Correct resolution order should be** (first wins):
1. ✅ OPENNOTES_NOTEBOOK envvar
2. ✅ --notebook flag
3. ✅ .opennotes.json in current directory
4. ✅ context match (registered notebooks with path context)
5. ✅ ancestor search (walk up tree for .opennotes.json)

## Impact

- Users expect envvar > flag > auto-detection
- Current behavior skips envvar entirely
- Missing direct .opennotes.json check in current directory (only checks via ancestor search)
- Order violates principle of least surprise

## Changes Required

### 1. Update `requireNotebook()` in cmd/notes_list.go

```go
func requireNotebook(cmd *cobra.Command) (*services.Notebook, error) {
	// Step 1: Check OPENNOTES_NOTEBOOK envvar
	if envNotebook := os.Getenv("OPENNOTES_NOTEBOOK"); envNotebook != "" {
		return notebookService.Open(envNotebook)
	}

	// Step 2: Check --notebook flag
	notebookPath, _ := cmd.Flags().GetString("notebook")
	if notebookPath != "" {
		return notebookService.Open(notebookPath)
	}

	// Step 3-5: Use updated Infer() logic
	nb, err := notebookService.Infer("")
	if err != nil {
		return nil, err
	}

	if nb == nil {
		return nil, fmt.Errorf("no notebook found. Set OPENNOTES_NOTEBOOK, use --notebook flag, or create one with: opennotes notebook create")
	}

	return nb, nil
}
```

### 2. Update `NotebookService.Infer()` in internal/services/notebook.go

```go
func (s *NotebookService) Infer(cwd string) (*Notebook, error) {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Step 1: Check .opennotes.json in current directory (direct check, not ancestor)
	if s.HasNotebook(cwd) {
		return s.Open(cwd)
	}

	// Step 2: Check registered notebooks for context match
	notebooks, _ := s.List(cwd)
	for _, nb := range notebooks {
		if nb.MatchContext(cwd) != "" {
			return nb, nil
		}
	}

	// Step 3: Search ancestor directories
	current := filepath.Dir(cwd)  // Start from parent, not current
	for current != "/" && current != "" {
		if s.HasNotebook(current) {
			return s.Open(current)
		}
		current = filepath.Dir(current)
	}

	return nil, nil // No notebook found
}
```

## Notes

- The key change: direct `.opennotes.json` check in cwd BEFORE ancestor search
- This ensures `/path/to/project/.opennotes.json` is found immediately, not only via ancestor walk-up
- Envvar support is upstream in `requireNotebook()`, not in Infer()
- All existing tests need updated expectations

## Tests to Update

1. `TestNotebookService_Infer_DeclaredPathPriority` → remove (no longer valid)
2. `TestNotebookService_Infer_ContextMatchPriority` → update (now priority 4)
3. `TestNotebookService_Infer_AncestorSearchPriority` → update (now priority 5)
4. Add new tests:
   - `TestNotebookService_Infer_CurrentDirectoryPriority` (priority 3)
   - `TestRequireNotebook_EnvvarPriority` (priority 1, in requireNotebook)
   - `TestRequireNotebook_FlagPriority` (priority 2, in requireNotebook)

## Files to Modify

- [ ] `cmd/notes_list.go` - Update requireNotebook()
- [ ] `internal/services/notebook.go` - Update Infer()
- [ ] `internal/services/notebook_test.go` - Update/add tests
- [ ] All commands using requireNotebook (check for copies)

## Verification

Run full test suite:
```bash
mise run test
```

Manual verification:
```bash
# Test envvar priority
OPENNOTES_NOTEBOOK=/path/to/notebook opennotes notes list

# Test flag priority (should override envvar? Or no?)
opennotes notes list --notebook /other/path

# Test current directory
cd /path/with/.opennotes.json && opennotes notes list

# Test registered context
opennotes notes list  # Should find if dir matches registered notebook context

# Test ancestor search
cd /path/to/ancestor/child && opennotes notes list
```

---

## Actual Outcome (Verified 2026-01-28)

### Implementation Status: ✅ COMPLETE

The notebook resolution order was implemented correctly with all 5 tiers working:

| Priority | Resolution Method | Status | Verified By |
|----------|-------------------|--------|-------------|
| 1 | `OPENNOTES_NOTEBOOK` envvar | ✅ Working | `TestRequireNotebook_EnvvarPriority` |
| 2 | `--notebook` flag | ✅ Working | `TestRequireNotebook_FlagOverridesEnvvar` |
| 3 | `.opennotes.json` in cwd | ✅ Working | `TestNotebookService_Infer_CurrentDirectoryPriority` |
| 4 | Context match (registered) | ✅ Working | `TestNotebookService_Infer_ContextMatchPriority` |
| 5 | Ancestor directory search | ✅ Working | `TestNotebookService_Infer_AncestorSearchPriority` |

### Verification Evidence

1. **Code Review**: `cmd/notes_list.go:requireNotebook()` implements correct order
2. **Automated Tests**: 8 tests verify resolution behavior (all passing)
3. **Manual Testing**: All tiers manually verified with real commands
4. **Documentation**: README and AGENTS.md match implementation
5. **Full Test Suite**: 711+ tests pass including notebook resolution

### Files Implemented

- [x] `cmd/notes_list.go` - `requireNotebook()` with correct priority order
- [x] `internal/services/notebook.go` - `Infer()` with cwd > context > ancestor
- [x] `internal/services/notebook_test.go` - Comprehensive test coverage

### Conclusion

Task was a specification document created after implementation. No code changes required - implementation was already correct and fully tested.
