# Jot TODO

## Active Epic
- [epic-8361d3a2](epic-8361d3a2-rename-to-jot.md) - **Rename Project to "Jot"** (in-progress)

## Phase 2 — In-Repo Changes ✅ COMPLETE

- [x] **Step 1**: Go module path + all imports (47 files)
- [x] **Step 2**: Constants in config.go and storage.go
- [x] **Step 3**: Env var prefix OPENNOTES_ → JOT_
- [x] **Step 4**: CLI command names & help text (10 cmd/ files)
- [x] **Step 5**: Rename opennotes.schema.json → jot.schema.json
- [x] **Step 6**: Build config (.goreleaser.yaml, Containerfiles, mise tasks)
- [x] **Step 7**: Tests (bats, Go e2e, unit tests)
- [x] **Step 8**: Documentation (docs/, pkgs/docs/, README, AGENTS.md, CHANGELOG.md)
- [x] **Step 9**: Pi extension package (pkgs/pi-opennotes/ → pkgs/pi-jot/)
- [x] **Step 10**: Build & test — all pass, 0 lint issues

## Remaining Rename Work

- [ ] **[NEEDS-HUMAN]** Phase 3: GitHub repo rename (`zenobi-us/opennotes` → `zenobi-us/jot`)
- [ ] **[NEEDS-HUMAN]** Phase 4: External updates (pkg.go.dev, any published packages)
- [ ] [task-8281af6b](task-8281af6b-notebook-migrate-versioned-framework.md) - Plan versioned `jot notebook migrate` framework (refined v1/v2 architecture proposal drafted)
- [ ] Phase 4: Write migration guide for existing users (config path changes)
- [ ] Regenerate `docs/notebook-discovery.svg` (contains embedded "opennotes" text)

## Parked Tasks
- [ ] [task-9c4a2f8d](task-9c4a2f8d-github-actions-moonrepo-releases.md) - GitHub Actions CI/CD
- [ ] [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) - DSL views implementation (10 tasks)

## Recently Completed
- ✅ Phase 1: Discovery (task-b66d2263) — Comprehensive inventory
- ✅ Phase 2: In-Repo Changes — All 10 steps complete, 2 commits on `feat/rename-to-jot`
