---
id: b66d2263
title: Identify All Rename Locations
created_at: 2026-02-18T19:41:00+10:30
updated_at: 2026-02-19T09:12:00+10:30
status: completed
epic_id: 8361d3a2
phase_id: phase-1-discovery
assigned_to: cleanup-2026-02-19
---

# Identify All Rename Locations

## Objective

Comprehensively identify every location — both inside the repository and external to it — where "opennotes" or "OpenNotes" appears and will need to be changed to "jot" or "Jot".

## Steps

### In-Repository Locations

- [x] **Go Module Path**: `go.mod` — `github.com/zenobi-us/opennotes`
- [x] **Import Statements**: All `*.go` files importing the module (47 files)
- [x] **Binary Name**: `.goreleaser.yaml` — `project_name` and `binary` fields
- [x] **Version Ldflags**: `.goreleaser.yaml` — `-X github.com/zenobius/opennotes/cmd.*` paths
- [x] **Documentation**: `README.md`, `AGENTS.md`, `CHANGELOG.md`, and docs/
- [x] **Config Paths**: Default config directory and notebook config file
- [x] **CLI Help Text**: Command descriptions, usage examples
- [x] **Test Files**: Test fixtures, expected output strings
- [x] **Memory Files**: `.memory/` project management files (deferred — update when rename is done)
- [x] **GitHub Actions**: `.github/workflows/` — No opennotes references found (clean)
- [x] **Containerfiles**: `Containerfile` and `Containerfile.test`
- [x] **Schema File**: `opennotes.schema.json`
- [x] **Mise Config**: `mise.toml` — No opennotes references found (clean)

### External Locations

- [x] **GitHub Repository**: `github.com/zenobi-us/opennotes` → needs rename to `jot`
- [x] **GitHub Releases**: Existing release assets named "opennotes" (goreleaser URLs)
- [x] **Go Package Registry**: `pkg.go.dev/github.com/zenobi-us/opennotes`
- [x] **`github.com/zenobi-us/jot`**: Available ✅ (404 — not taken)

### Research: Naming Conflicts

- [x] **`jot` binary on system**: Not found in PATH ✅
- [x] **`jot` apt package**: Not found ✅
- [x] **NOTE**: BSD systems have a `jot(1)` utility (generates sequential/random data). This is a known conflict on macOS/FreeBSD but not Linux. Users on those systems would need to use full path or alias.

## Expected Outcome

A comprehensive inventory document listing all changes needed.

## Actual Outcome

### Complete Inventory

#### Category 1: Go Module & Imports (47 files)

**Pattern**: `github.com/zenobi-us/opennotes` → `github.com/zenobi-us/jot`

| Area | Files | Notes |
|------|-------|-------|
| `go.mod` | 1 | Module declaration |
| `main.go` | 1 | Root import |
| `cmd/*.go` | 12 | CLI command files |
| `internal/search/bleve/*.go` | 8 | Search backend |
| `internal/search/parser/*.go` | 4 | Query parser |
| `internal/services/*.go` | 18 | Service layer |
| `internal/testutil/*.go` | 2 | Test utilities |
| `tests/e2e/*.go` | 3 | End-to-end tests |
| `pkgs/md-process-blocks-cli/go.mod` | 1 | Sub-package module |

**Approach**: `go mod edit -module` + `find/sed` for imports, then `go mod tidy`.

#### Category 2: Constants & Config Paths (3 files)

| File | Constant/Value | Old | New |
|------|----------------|-----|-----|
| `internal/services/config.go:19` | `NotebookConfigFile` | `.opennotes.json` | `.jot.json` |
| `internal/services/config.go:47` | Config dir | `opennotes/config.json` | `jot/config.json` |
| `internal/search/bleve/storage.go:139` | `IndexDir` | `.opennotes/index` | `.jot/index` |

#### Category 3: Environment Variables (2 files + docs)

| Old | New | Files |
|-----|-----|-------|
| `OPENNOTES_CONFIG` | `JOT_CONFIG` | `internal/services/config.go`, `cmd/root.go`, docs |
| `OPENNOTES_NOTEBOOK` | `JOT_NOTEBOOK` | `cmd/notes_list.go`, docs |
| `OPENNOTES_` prefix | `JOT_` prefix | `internal/services/config.go:83` |

#### Category 4: CLI Command Names & Help Text (8 files)

| File | Changes |
|------|---------|
| `cmd/root.go` | `Use: "opennotes"` → `"jot"`, Long description, examples |
| `cmd/init.go` | Help text, examples |
| `cmd/notes.go` | Help text, examples |
| `cmd/notes_search.go` | ~20 example lines |
| `cmd/notes_view.go` | ~10 example lines, config path refs |
| `cmd/notes_list.go` | Comments, error message |
| `cmd/notebook.go` | Help text, examples |
| `cmd/notebook_create.go` | Help text |
| `cmd/notebook_list.go` | Help text |
| `cmd/notebook_register.go` | Help text |

#### Category 5: Build & Release Configuration

| File | Changes |
|------|---------|
| `.goreleaser.yaml` | `project_name`, `binary`, ldflags paths, release URLs, install instructions |
| `Containerfile` | Comment "Multi-stage Containerfile for OpenNotes" |
| `Containerfile.test` | Config path `~/.config/opennotes/` |
| `opennotes.schema.json` | Rename file → `jot.schema.json`, update description |

#### Category 6: Test Files

| File | Approx Changes |
|------|----------------|
| `tests/e2e/core-smoke.bats` | Config path, `.opennotes.json` refs |
| `tests/e2e/smoke.bats` | Config path, `.opennotes.json` refs |
| `tests/e2e/go_smoke_test.go` | ~6 `.opennotes.json` refs |
| `tests/e2e/filesystem_errors_test.go` | `.opennotes.json` ref |
| `tests/e2e/search_test.go` | `.opennotes.json` ref |
| `tests/e2e/stress_test.go` | ~4 `.opennotes.json` refs |
| `tests/e2e/command_errors_test.go` | `.opennotes.json` ref |
| `internal/testutil/notebook.go` | `.opennotes.json` config creation |
| `internal/services/notebook_test.go` | `.opennotes.json` refs |

#### Category 7: Documentation (many files)

| Directory | Files | Approx Refs |
|-----------|-------|-------------|
| `docs/` | 10+ files | ~100+ refs |
| `pkgs/docs/` | 5+ files | ~50+ refs |
| `pkgs/pi-opennotes/docs/` | 5+ files | ~50+ refs |
| `README.md` | 1 | ~5 refs |
| `AGENTS.md` | 1 | ~10 refs |
| `CHANGELOG.md` | 1 | ~5 refs |

#### Category 8: Pi-OpenNotes Extension Package

| Area | Notes |
|------|-------|
| `pkgs/pi-opennotes/` | **Entire directory** needs rename to `pkgs/pi-jot/` |
| Source files | `OPENNOTES_*` env vars, `.opennotes.json` refs |
| Docs | Configuration, troubleshooting, tool-usage guides |
| Tests | Setup, notebook tests |

#### Category 9: External / Post-Rename

| Item | Action |
|------|--------|
| GitHub repo name | Rename `zenobi-us/opennotes` → `zenobi-us/jot` |
| GitHub redirect | GitHub auto-redirects old URL |
| Go module proxy | New module path takes effect on next publish |
| pkg.go.dev | Will index new module path automatically |
| Existing users | Need migration guide for config paths |

### Recommended Order of Operations

1. **Update Go module path** (`go.mod`, all imports) — mechanical, use sed
2. **Update constants** (`config.go`, `storage.go`) — 3 edits
3. **Update env var prefix** (`OPENNOTES_` → `JOT_`) — few files
4. **Update CLI command names & help** — 10 files
5. **Rename schema file** (`opennotes.schema.json` → `jot.schema.json`)
6. **Update build config** (`.goreleaser.yaml`, Containerfiles)
7. **Update tests** (bats, Go e2e tests, unit tests)
8. **Update documentation** (docs/, README, AGENTS.md)
9. **Rename pi extension package** (`pkgs/pi-opennotes/` → `pkgs/pi-jot/`)
10. **Build & run full test suite** to verify
11. **GitHub repo rename** (post-merge, manual)
12. **Migration guide** for existing users (config path changes)

### Migration Concerns

- **Config directory**: `~/.config/opennotes/` → `~/.config/jot/` — need migration code or docs
- **Notebook config**: `.opennotes.json` → `.jot.json` — need migration or backward compat
- **Index directory**: `.opennotes/index/` → `.jot/index/` — will trigger re-index
- **BSD `jot` conflict**: macOS/FreeBSD have `jot(1)` — document in README

## Lessons Learned

- The codebase has extensive documentation and help text, making up the bulk of rename locations
- Go's import path renaming is mechanical but touches 47 files
- The pi extension package (`pkgs/pi-opennotes/`) is a significant second codebase to rename
- No naming conflicts found on Linux; BSD `jot(1)` is a known issue to document
- GitHub workflows and mise.toml are clean — no opennotes references
