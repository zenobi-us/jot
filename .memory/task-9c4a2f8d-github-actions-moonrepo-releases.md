---
id: 9c4a2f8d
title: GitHub Actions CI/CD with moonrepo, release-please, and GoReleaser
created_at: 2026-01-29T17:33:00+10:30
updated_at: 2026-01-29T17:37:50+10:30
status: todo
epic_id: null
phase_id: null
assigned_to: null
---

# GitHub Actions CI/CD with moonrepo and release-please

## Objective

Update GitHub Actions workflows to integrate moonrepo's affected command detection and release-please manifest mode for independent, safe package releases in a monorepo structure.

## Context

This task was deferred during Phase 4 (Polish) of the Go migration work. The original requirement was to "Update CI/CD for Go builds" but this has expanded to include:

1. Supporting multiple packages (Go and TypeScript/Bun) in the same repository
2. Using moonrepo's dependency graph to determine what needs testing
3. Using release-please for automatic version bumping and changelog generation
4. Only releasing packages that have actual changes

### Original Deferred Task Location

From `.memory/archive/01-migrate-to-golang/task-1k2l3m4n-polish.md`:
- Section 4.7 Build Configuration
- Reason deferred: Needed post-merge integration after Go rewrite was production-ready

## Research Findings

### Strategy: Implicit Detection + Dependency Graph (Recommended)

After evaluating multiple approaches, the recommended strategy is **"Strategy 3"** which combines:
- **release-please** with implicit detection (only bump versions for packages with commits in their scope)
- **moonrepo** as the dependency graph enforcer (ensures dependent packages are tested)

This approach provides:
- ✅ Clean, independent version bumps (no linked versions)
- ✅ Safety through dependency-aware testing
- ✅ Prevents releases if dependent packages break
- ✅ Simple configuration and maintenance

### moonrepo Integration

**Purpose**: Dependency graph enforcement and affected test detection

**Key Commands**:
- `moon ci --base main` - Run tests for affected projects based on git diff
- This ensures that when `go-api` changes, `web-app` tests also run (if web-app depends on go-api)

**Configuration** (`moon.yml`):
```yaml
# In apps/web-app/moon.yml
dependsOn:
  - services/go-api
```

**Benefits**:
- Automatically discovers what needs testing based on changes
- Respects dependency relationships
- Prevents breaking changes from being released

### release-please Configuration

**Purpose**: Automatic version bumping and changelog generation for packages with actual changes

**Mode**: Manifest mode with separate package definitions

**Key Configuration** (`release-please-config.json`):
```json
{
  "packages": {
    "services/go-api": {
      "component": "go-api",
      "release-type": "go",
      "changelog-path": "CHANGELOG.md"
    },
    "apps/web-app": {
      "component": "web-app", 
      "release-type": "node",
      "changelog-path": "CHANGELOG.md"
    }
  }
}
```

**Important**: Do NOT use `linked-versions` - this would force all packages to bump together, defeating the purpose of independent releases.

**Behavior**:
- Scans commit messages following Conventional Commits
- Only bumps versions for packages with commits in their scope (e.g., `fix(go-api): ...`)
- Creates separate PRs for each package that needs releasing
- Updates version files automatically:
  - Node/Bun: `package.json`
  - Go: `version.go` or similar (needs setup)

### CI Workflow Structure

**Recommended Flow**:

```yaml
jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Need full history for moonrepo
      
      - name: Setup moonrepo
        uses: moonrepo/setup-toolchain@v0
      
      - name: Run affected tests
        run: moon ci --base main
  
  release:
    needs: quality-gate
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: google-github-actions/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

**Key Points**:
1. Quality gate runs first using moonrepo
2. Release job only runs after tests pass
3. Release job only triggers on main branch
4. release-please creates PRs for version bumps
5. Merging the release PR triggers actual release creation

## Steps

### 1. Setup moonrepo Configuration

- [ ] Verify `moon.yml` exists in project root
- [ ] Define dependencies between packages in individual `moon.yml` files
  ```yaml
  # Example: apps/web-app/moon.yml
  dependsOn:
    - services/go-api
  ```
- [ ] Test locally: `moon ci --base main`

### 2. Setup release-please Configuration

- [ ] Create `release-please-config.json` in project root
- [ ] Define packages with appropriate release types:
  - `services/go-api` → `release-type: "go"`
  - `apps/web-app` → `release-type: "node"`
  - `pkgs/pi-opennotes` → `release-type: "node"`
- [ ] Create initial `.release-please-manifest.json` with current versions
- [ ] Verify Conventional Commits usage in project

### 3. Setup Version Files for Go Packages

For Go packages, release-please needs a version file to update:

- [ ] Create `services/go-api/version.go`:
  ```go
  package main
  
  // Version is the current version of go-api
  const Version = "1.0.0"
  ```
- [ ] Update build process to use this version
- [ ] Add version flag to CLI: `opennotes --version`

### 4. Update GitHub Actions Workflows

**4.1 Update or Create `pr.yml` (PR Quality Gate)**

- [ ] Add moonrepo setup step
- [ ] Replace test commands with `moon ci --base ${{ github.base_ref }}`
- [ ] Ensure full git history is fetched (`fetch-depth: 0`)
- [ ] Add Go toolchain setup for Go packages
- [ ] Add Bun setup for TypeScript packages

**4.2 Update or Create `publish.yml` (Release Workflow)**

- [ ] Add quality gate job (runs moonrepo tests)
- [ ] Add release job that depends on quality gate
- [ ] Use `google-github-actions/release-please-action@v4`
- [ ] Configure with `release-please-config.json`
- [ ] Add publish steps for each package type:
  - Go: GitHub releases with binaries
  - Node: npm publish

**4.3 Test Workflow Structure**

```yaml
name: CI/CD

on:
  pull_request:
  push:
    branches: [main]

jobs:
  affected-tests:
    name: Run Affected Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: moonrepo/setup-toolchain@v0
      
      - name: Run tests for affected projects
        run: moon ci --base main
  
  release-please:
    name: Create Releases
    needs: affected-tests
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: google-github-actions/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

### 5. Documentation

- [ ] Document workflow in `docs/development.md`:
  - How moonrepo detects affected projects
  - How to trigger releases (Conventional Commits)
  - How to test CI locally
- [ ] Add CONTRIBUTING.md section on commit message format
- [ ] Update README with badge links to CI status

### 6. Testing

- [ ] Create test PR with changes to single package
- [ ] Verify only affected tests run
- [ ] Verify release-please creates correct PR
- [ ] Test actual release process on a test tag/version

## Expected Outcome

After completion:

1. **Automated Testing**: PRs automatically run tests for affected packages based on dependency graph
2. **Safe Releases**: Cannot release if dependent packages have test failures
3. **Independent Versions**: Each package gets its own version bump when it has changes
4. **Clean Changelogs**: Automatic changelog generation per package
5. **Reduced CI Time**: Only affected projects are tested/built
6. **Developer Experience**: Clear feedback on what's being tested and why

## Actual Outcome

_To be filled after implementation_

## Lessons Learned

_To be filled after implementation_

## GoReleaser Integration for Go Packages

### Overview

GoReleaser provides native monorepo support that can work alongside our release-please + moonrepo strategy. This is particularly relevant for our Go packages, enabling automated binary builds, cross-compilation, and distribution.

### Key GoReleaser Monorepo Features

1. **Path-based filtering**: Can build/release only changed Go modules
2. **Multiple .goreleaser.yml files**: Support for per-package configuration
3. **Integration with tag patterns**: Works with tags like `service-name/v1.2.3`
4. **Build matrix support**: Can handle multiple Go modules in one release workflow

### Integration Points with release-please

**Potential Workflow**:
1. release-please creates version bump PR with Conventional Commits
2. Merging PR triggers release-please to create tags (e.g., `go-api-v1.2.3`)
3. Tag creation triggers GoReleaser workflow
4. GoReleaser builds binaries, creates GitHub release, attaches artifacts

**Benefits**:
- release-please handles version management and changelogs
- GoReleaser handles binary compilation and distribution
- Clean separation of concerns
- Both tools play to their strengths

### Research Questions

These questions need to be answered during implementation:

1. **Tag Pattern Compatibility**: How does GoReleaser fit with release-please's tag creation?
   - Does release-please create tags in a format GoReleaser expects?
   - Do we need custom tag patterns for monorepo packages?

2. **Tool Division of Labor**: Should we use GoReleaser for Go packages and release-please for version management?
   - Or does release-please handle everything (version + release creation)?
   - What's the cleanest workflow separation?

3. **Tag Consumption**: Can GoReleaser consume the tags that release-please creates?
   - Do tag patterns need alignment (e.g., `go-api/v1.2.3` vs `go-api-v1.2.3`)?
   - Does GoReleaser need configuration to match release-please's tag format?

4. **Workflow Separation**: Do we need separate release workflows for Go vs Bun/Node packages?
   - Go packages → release-please (version) + GoReleaser (build/publish)
   - Node packages → release-please (version + npm publish)
   - Or unified workflow with conditional steps?

5. **Multiple .goreleaser.yml Files**: Should each Go package have its own GoReleaser config?
   - Pro: Package-specific build configurations
   - Con: Configuration duplication
   - Alternative: Single config with monorepo filtering

### Implementation Considerations

**Recommended Approach** (to be validated):

```yaml
# Workflow triggered by tag creation
on:
  push:
    tags:
      - 'go-api-v*'
      - 'other-go-package-v*'

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v5
      
      - uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
          workdir: ./services/go-api  # Path-based filtering
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**GoReleaser Configuration** (`.goreleaser.yml` in package directory):

```yaml
# services/go-api/.goreleaser.yml
project_name: opennotes

builds:
  - main: ./main.go
    binary: opennotes
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

release:
  github:
    owner: zenobi-us
    name: opennotes
  name_template: "go-api v{{ .Version }}"
```

### Decision Points

Before implementing, decide:

1. **Tag Format**: Align release-please and GoReleaser on tag patterns
   - Recommended: `go-api-v1.2.3` (works with both tools)

2. **Workflow Trigger**: When should GoReleaser run?
   - Option A: On tag creation (after release-please creates tag)
   - Option B: After release-please creates release (duplicate effort?)

3. **Binary Distribution**: Where should Go binaries be published?
   - GitHub Releases (GoReleaser default)
   - Package registries (Homebrew, apt, etc.)
   - Docker images

4. **Monorepo Strategy**: Single or multiple GoReleaser configs?
   - Recommendation: Per-package configs for flexibility

### Action Items for Implementation

- [ ] Review GoReleaser monorepo documentation thoroughly
- [ ] Align tag patterns between release-please and GoReleaser
- [ ] Create `.goreleaser.yml` for `services/go-api`
- [ ] Add GoReleaser workflow triggered by tag creation
- [ ] Test workflow with test tag
- [ ] Document GoReleaser + release-please integration

## References

- [moonrepo Documentation](https://moonrepo.dev/)
- [release-please Documentation](https://github.com/googleapis/release-please)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Google release-please Action](https://github.com/google-github-actions/release-please-action)
- [GoReleaser Monorepo Support](https://goreleaser.com/customization/monorepo/)

## Notes

### Package Structure Assumptions

This task assumes the following package structure:
- `services/go-api/` - Go CLI application (main OpenNotes binary)
- `apps/web-app/` - Bun/TypeScript web application (example dependent)
- `pkgs/pi-opennotes/` - Bun/TypeScript pi extension package

Adjust configurations if actual structure differs.

### Conventional Commits Requirement

For release-please to work correctly, the project MUST use Conventional Commits:
- `feat(go-api): add new flag` → minor version bump for go-api
- `fix(web-app): correct typo` → patch version bump for web-app
- `feat!: breaking change` → major version bump

### Go Version Management

Go doesn't have a standard version file like `package.json`. Options:
1. Use `version.go` with a const (recommended for simplicity)
2. Use build flags: `-ldflags "-X main.Version=1.0.0"`
3. Use go.mod version tag (less common)

The task uses option 1 for simplicity and clarity.

### Deployment Targets

Consider where releases should be published:
- **Go Binary**: GitHub Releases with compiled binaries (linux, macOS, windows)
- **Node/Bun Packages**: npm registry as scoped packages (@zenobi-us/...)
- **Docker Images**: Optional, for containerized deployments

Each publish target may need separate workflow jobs.
