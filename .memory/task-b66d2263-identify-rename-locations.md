---
id: b66d2263
title: Identify All Rename Locations
created_at: 2026-02-18T19:41:00+10:30
updated_at: 2026-02-18T19:41:00+10:30
status: todo
epic_id: 8361d3a2
phase_id: phase-1-discovery
assigned_to: unassigned
---

# Identify All Rename Locations

## Objective

Comprehensively identify every location — both inside the repository and external to it — where "opennotes" or "OpenNotes" appears and will need to be changed to "jot" or "Jot".

## Steps

### In-Repository Locations

- [ ] **Go Module Path**: `go.mod` — `github.com/zenobi-us/opennotes`
- [ ] **Import Statements**: All `*.go` files importing the module
- [ ] **Binary Name**: `.goreleaser.yaml` — `project_name` and `binary` fields
- [ ] **Version Ldflags**: `.goreleaser.yaml` — `-X github.com/zenobius/opennotes/cmd.*` paths
- [ ] **Documentation**: `README.md`, `AGENTS.md`, any other docs
- [ ] **Config Paths**: 
  - Default config directory: `~/.config/opennotes/`
  - Notebook config file: `.opennotes.json`
- [ ] **CLI Help Text**: Command descriptions, usage examples
- [ ] **Test Files**: Test fixtures, expected output strings
- [ ] **Memory Files**: `.memory/` project management files
- [ ] **GitHub Actions**: `.github/workflows/` if any CI/CD references

### External Locations

- [ ] **GitHub Repository**: `github.com/zenobi-us/opennotes`
  - Repo name
  - Repo description
  - Topics/tags
- [ ] **GitHub Releases**: Existing release assets named "opennotes"
- [ ] **Go Package Registry**: `pkg.go.dev/github.com/zenobi-us/opennotes`
- [ ] **Any published packages**: Homebrew, AUR, etc. (if applicable)
- [ ] **Personal/project websites**: Any references to the project
- [ ] **Social media/announcements**: If any exist

### Research Required

- [ ] Use `lynx` or web search to check if "jot" conflicts with existing popular CLI tools
- [ ] Verify `github.com/zenobi-us/jot` is available
- [ ] Check if `jot` package name exists on Go pkg.go.dev

## Expected Outcome

A comprehensive inventory document listing:
1. Every file in the repo that needs modification
2. Exact line numbers or patterns to change
3. All external services/platforms that need updating
4. Any naming conflicts or concerns discovered
5. Recommended order of operations for the rename

## Actual Outcome

_To be filled after task completion_

## Lessons Learned

_To be filled after task completion_
