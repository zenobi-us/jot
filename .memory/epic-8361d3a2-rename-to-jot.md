---
id: 8361d3a2
title: Rename Project from OpenNotes to Jot
created_at: 2026-02-18T19:41:00+10:30
updated_at: 2026-02-19T09:12:00+10:30
status: in-progress
---

# Rename Project from OpenNotes to Jot

## Vision/Goal

Rebrand the project from "OpenNotes" to "Jot" — a shorter, more memorable, action-oriented name that better captures the essence of quick note-taking from the command line.

**New Identity:**
- **Name**: Jot
- **Binary**: `jot`
- **Module**: `github.com/zenobi-us/jot`
- **Repo**: `github.com/zenobi-us/jot`

## Success Criteria

- [ ] All in-repo references updated (module path, imports, binary name, docs)
- [ ] GitHub repository renamed to `jot`
- [ ] Go module path updated and working
- [ ] Binary builds as `jot`
- [ ] README and documentation reflect new name
- [ ] Config directory updated (`.config/jot/` vs `.config/opennotes/`)
- [ ] Notebook config file renamed (`.jot.json` vs `.opennotes.json`)
- [ ] All tests pass with new naming
- [ ] Release/distribution updated (goreleaser, etc.)
- [ ] External references identified and migration plan documented

## Phases

1. **Phase 1: Discovery** — ✅ COMPLETE
   - Task: [task-b66d2263](task-b66d2263-identify-rename-locations.md)
   - Found: 47 Go files, 3 constants, 2 env vars, 10 CLI files, 10+ doc files, tests, build config
   - No naming conflicts on Linux; BSD `jot(1)` is documented concern

2. **Phase 2: In-Repo Changes** — READY
   - Step 1: Go module path + imports (47 files — mechanical sed)
   - Step 2: Constants (config.go, storage.go — 3 edits)
   - Step 3: Env var prefix (OPENNOTES_ → JOT_)
   - Step 4: CLI command names & help text (10 files)
   - Step 5: Rename schema file
   - Step 6: Build config (goreleaser, Containerfiles)
   - Step 7: Tests (bats, Go e2e, unit)
   - Step 8: Documentation (docs/, README, AGENTS.md)
   - Step 9: Pi extension package (pkgs/pi-opennotes/ → pkgs/pi-jot/)
   - Step 10: Build & run full test suite

3. **Phase 3: GitHub Rename** — TBD after Phase 2 complete
   - Rename repository
   - Update remotes

4. **Phase 4: External Updates** — TBD after Phase 3 complete
   - Package managers, documentation sites
   - Migration guide for existing users

## Dependencies

- Current feature branch should be merged or rebased before major rename
- Consider completing DSL views research first (epic-f661c068 residual)

## Notes

- Name change rationale: "Jot" is short (3 letters), action-oriented (verb), memorable, and perfectly describes the CLI note-taking use case
- Migration path needed for existing users (config file location, notebook files)
- BSD `jot(1)` conflict: document in README, not a blocker for Linux-primary tool
