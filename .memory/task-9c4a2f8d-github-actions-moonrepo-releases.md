---
id: 9c4a2f8d
title: GitHub Actions CI/CD with moonrepo, release-please, and GoReleaser
created_at: 2026-01-29T17:33:00+10:30
updated_at: 2026-01-29T17:49:00+10:30
status: todo
epic_id: null
phase_id: null
assigned_to: null
---

# GitHub Actions CI/CD with moonrepo and release-please

## Objective

Update GitHub Actions workflows to integrate moonrepo's affected command detection and release-please manifest mode for independent, safe package releases in a monorepo structure.

## Important Clarifications

### Workflow File Targets

**CRITICAL**: We are modifying `.github/workflows/publish.yml`, **NOT** `.github/workflows/release.yml`.

The `release.yml` file is used by moonrepo's own release process and should not be modified for our CI/CD integration.

### Mise Task Reorganization

As part of this work, we need to reorganize mise tasks to support multiple package ecosystems:

**Current Structure**:
```
.mise/tasks/publish
```

**New Structure**:
```
.mise/tasks/
├── node/
│   └── publish    # Node/Bun package publishing
├── go/
│   └── publish    # Go binary publishing (future)
```

**Rationale**:
- Namespaces tasks by language/runtime
- Allows different publish workflows for different package types
- Keeps tasks organized as the monorepo grows
- Enables calling specific publish tasks: `mise run node:publish`

**Mise Variadic Arguments**:

According to [mise documentation on task arguments](https://mise.jdx.dev/tasks/task-arguments.html#variadic-arguments), tasks can accept variadic arguments using the `...` suffix:

```toml
# .mise/tasks/node/publish
[tasks.publish]
run = "npm publish {{arg(name='packages', var=true)...}}"
```

This enables:
- `mise run node:publish` - Publish all packages
- `mise run node:publish @zenobi-us/opennotes` - Publish specific package
- `mise run node:publish pkg1 pkg2` - Publish multiple packages

This variadic capability integrates with moonrepo's affected detection, allowing selective publishing based on what changed.

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

### 1. Reorganize Mise Tasks

**Goal**: Create namespaced task structure for different package types

- [ ] Create directory structure:
  ```bash
  mkdir -p .mise/tasks/node
  mkdir -p .mise/tasks/go
  ```
- [ ] Move existing publish task:
  ```bash
  git mv .mise/tasks/publish .mise/tasks/node/publish
  ```
- [ ] Update `.mise/tasks/node/publish` to accept variadic arguments:
  ```toml
  # Example structure for variadic package arguments
  [tasks.publish]
  description = "Publish Node/Bun packages"
  run = """
  # Accept variadic package names as arguments
  # Usage: mise run node:publish [@zenobi-us/pkg1] [@zenobi-us/pkg2]
  npm publish {{arg(name='packages', var=true)...}}
  """
  ```
- [ ] Test task invocation:
  ```bash
  mise run node:publish --help
  mise run node:publish --dry-run
  ```
- [ ] Document new task structure in project docs

**References**:
- [Mise Task Arguments Documentation](https://mise.jdx.dev/tasks/task-arguments.html#variadic-arguments)
- Variadic args enable selective publishing: `mise run node:publish pkg1 pkg2`

### 2. Setup moonrepo Configuration

- [ ] Verify `moon.yml` exists in project root
- [ ] Define dependencies between packages in individual `moon.yml` files
  ```yaml
  # Example: apps/web-app/moon.yml
  dependsOn:
    - services/go-api
  ```
- [ ] Test locally: `moon ci --base main`

### 3. Setup release-please Configuration

- [ ] Create `release-please-config.json` in project root
- [ ] Define packages with appropriate release types:
  - `services/go-api` → `release-type: "go"`
  - `apps/web-app` → `release-type: "node"`
  - `pkgs/pi-opennotes` → `release-type: "node"`
- [ ] Create initial `.release-please-manifest.json` with current versions
- [ ] Verify Conventional Commits usage in project

### 4. Setup Version Files for Go Packages

For Go packages, release-please needs a version file to update:

- [ ] Create `services/go-api/version.go`:
  ```go
  package main
  
  // Version is the current version of go-api
  const Version = "1.0.0"
  ```
- [ ] Update build process to use this version
- [ ] Add version flag to CLI: `opennotes --version`

### 5. Update GitHub Actions Workflows

**5.1 Update or Create `pr.yml` (PR Quality Gate)**

- [ ] Add moonrepo setup step
- [ ] Replace test commands with `moon ci --base ${{ github.base_ref }}`
- [ ] Ensure full git history is fetched (`fetch-depth: 0`)
- [ ] Add Go toolchain setup for Go packages
- [ ] Add Bun setup for TypeScript packages

**5.2 Modify `publish.yml` (Release Workflow)**

**IMPORTANT**: Modify `.github/workflows/publish.yml`, NOT `release.yml`

- [ ] Add quality gate job (runs moonrepo tests)
- [ ] Add release job that depends on quality gate
- [ ] Use `google-github-actions/release-please-action@v4`
- [ ] Configure with `release-please-config.json`
- [ ] Add publish steps for each package type:
  - Node: Call `mise run node:publish` with variadic package arguments
  - Go: GitHub releases with binaries (future: `mise run go:publish`)
- [ ] Integrate with moonrepo affected detection to determine which packages to publish:
  ```yaml
  - name: Get affected packages
    id: affected
    run: moon query projects --affected --json
  
  - name: Publish affected Node packages
    if: steps.affected.outputs.packages != ''
    run: mise run node:publish ${{ steps.affected.outputs.packages }}
  ```

**5.3 Test Workflow Structure**

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

### 6. Documentation

- [ ] Document workflow in `docs/development.md`:
  - How moonrepo detects affected projects
  - How to trigger releases (Conventional Commits)
  - How to test CI locally
  - New mise task structure and usage
  - How to pass package arguments to publish tasks
- [ ] Add CONTRIBUTING.md section on commit message format
- [ ] Update README with badge links to CI status
- [ ] Document mise task reorganization:
  - `.mise/tasks/node/publish` - Node/Bun packages
  - `.mise/tasks/go/publish` - Go binaries (future)
  - Examples: `mise run node:publish @zenobi-us/pkg1`

### 7. Testing

- [ ] Create test PR with changes to single package
- [ ] Verify only affected tests run
- [ ] Verify release-please creates correct PR
- [ ] Test actual release process on a test tag/version
- [ ] Test mise task invocation:
  - `mise run node:publish --dry-run`
  - `mise run node:publish @zenobi-us/opennotes`
  - Verify variadic arguments work correctly

## Future Mise Tasks

### Go Package Publishing

When Go packages need automated publishing, create `.mise/tasks/go/publish`:

```toml
[tasks.publish]
description = "Build and publish Go binaries"
run = """
# Accept variadic package names as arguments
# Usage: mise run go:publish [service-name] [another-service]
PACKAGES={{arg(name='packages', var=true, default='all')...}}

if [ "$PACKAGES" = "all" ]; then
  # Build all Go services
  goreleaser release --clean
else
  # Build specific services
  for pkg in $PACKAGES; do
    goreleaser release --clean --single-target --id=$pkg
  done
fi
"""
```

**Benefits**:
- Consistent interface with `node:publish`
- Selective building based on affected packages
- Integration with GoReleaser for cross-compilation
- Variadic arguments for multiple services

**Integration with CI**:
```yaml
- name: Publish affected Go services
  if: contains(steps.affected.outputs.types, 'go')
  run: mise run go:publish ${{ steps.affected.outputs.go_packages }}
```

### Other Runtime Tasks

As the monorepo grows, additional task namespaces can be created:

- `.mise/tasks/rust/publish` - Rust crate publishing
- `.mise/tasks/python/publish` - Python package publishing  
- `.mise/tasks/docker/publish` - Container image publishing

Each namespace follows the same pattern:
1. Variadic argument support for selective operations
2. Integration with moonrepo affected detection
3. Consistent naming and interface across runtimes

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
- [Mise Task Arguments - Variadic Arguments](https://mise.jdx.dev/tasks/task-arguments.html#variadic-arguments)

## Notes

### Workflow File Clarification

**CRITICAL**: This task modifies `.github/workflows/publish.yml`, NOT `release.yml`.

- `publish.yml` - Custom CI/CD workflow for our package publishing
- `release.yml` - Moonrepo's own release process (do not modify)

### Package Structure Assumptions

This task assumes the following package structure:
- `services/go-api/` - Go CLI application (main OpenNotes binary)
- `apps/web-app/` - Bun/TypeScript web application (example dependent)
- `pkgs/pi-opennotes/` - Bun/TypeScript pi extension package

Adjust configurations if actual structure differs.

### Mise Task Reorganization Rationale

Moving from `.mise/tasks/publish` to `.mise/tasks/node/publish`:

**Benefits**:
1. **Namespace Clarity**: Tasks grouped by runtime/language
2. **Scalability**: Easy to add `go/publish`, `rust/publish`, etc.
3. **Selective Invocation**: `mise run node:publish` vs `mise run go:publish`
4. **Variadic Arguments**: Pass specific packages to publish
5. **Integration**: Works seamlessly with moonrepo affected detection

**Migration Path**:
```bash
# Old way
mise run publish

# New way  
mise run node:publish
mise run node:publish @zenobi-us/pkg1 @zenobi-us/pkg2
```

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
