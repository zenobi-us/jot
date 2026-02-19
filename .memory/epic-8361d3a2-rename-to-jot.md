---
id: 8361d3a2
title: Rename Project from OpenNotes to Jot
created_at: 2026-02-18T19:41:00+10:30
updated_at: 2026-02-18T19:41:00+10:30
status: planning
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

1. **Phase 1: Discovery** — Identify all locations requiring changes (in-repo and external)
   - Task: [task-b66d2263](task-b66d2263-identify-rename-locations.md)

2. **Phase 2: In-Repo Changes** — Update all code, configs, and documentation
   - TBD after Phase 1 complete

3. **Phase 3: GitHub Rename** — Rename repository and update remotes
   - TBD after Phase 2 complete

4. **Phase 4: External Updates** — Update any external references (package managers, etc.)
   - TBD after Phase 3 complete

## Dependencies

- Current feature branch should be merged or rebased before major rename
- Consider completing DSL views research first (epic-f661c068 residual)

## Notes

- Name change rationale: "Jot" is short (3 letters), action-oriented (verb), memorable, and perfectly describes the CLI note-taking use case
- Migration path needed for existing users (config file location, notebook files)
