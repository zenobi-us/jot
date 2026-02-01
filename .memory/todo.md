# OpenNotes TODO

## Current Status

**Active Epics**: 2
1. **Pi-OpenNotes Extension** - Phase 3 Complete, Ready for Phase 4 (Distribution)
2. **Remove DuckDB Epic** - 🆕 Research Phase Starting

---

## 🆕 Remove DuckDB - Alternative Search

### Epic f661c068 - Research Phase

**Current Tasks**: 
1. [research-dbb5cdc8-zk-search-analysis.md](research-dbb5cdc8-zk-search-analysis.md) - ZK search analysis
2. [research-45af3ec0-golang-vector-rag-search.md](research-45af3ec0-golang-vector-rag-search.md) - 🆕 Go vector RAG exploration

**Research Checklist - ZK Search Analysis**:
- [ ] Clone zk-org/zk repository to /tmp/zk-analysis
- [ ] Run CodeMapper analysis (`cm stats`, `cm query`, `cm trace`)
- [ ] Use LSP tools to understand key types and interfaces
- [ ] Map query parsing code path
- [ ] Map indexing code path  
- [ ] Map search execution code path
- [ ] Create ASCII state machine diagrams for:
  - [ ] Query parsing flow
  - [ ] Indexing flow
  - [ ] Search execution flow
- [ ] Document integration opportunities with afero
- [ ] Write recommendations for OpenNotes implementation
- [ ] Update epic with refined phase definitions

**Goal**: Understand how zk implements search without DuckDB, identify code paths, create state machine diagrams to guide our implementation.

**Research Checklist - Go Vector RAG Search**:
- [ ] Survey Go vector database/search libraries (Chroma-go, Milvus, Weaviate, pure-Go)
- [ ] Identify embedding generation options (local ONNX, API-based)
- [ ] Document RAG architecture patterns in Go
- [ ] Prototype minimal RAG example with sample markdown notes
- [ ] Measure performance (indexing time, query latency, memory)
- [ ] Compare vector search vs traditional text search
- [ ] Evaluate hybrid search strategy (combining both approaches)
- [ ] Check afero filesystem compatibility
- [ ] Recommend: include in epic, defer to future, or skip
- [ ] Update epic-f661c068 with decision and findings

**Goal**: Explore semantic/vector search capabilities in Go as complementary or alternative to traditional text search. Inspired by qmd (Node.js tool we can't use), find Go equivalents for RAG patterns.

---

## ⏳ Awaiting Review

### Pi-OpenNotes Extension - Phase 1 Complete

**Request**: Please review Phase 1 design documents before proceeding to implementation.

**Documents to Review**:
1. [task-a0236e7c-document-opennotes-cli.md](task-a0236e7c-document-opennotes-cli.md) - CLI interface reference
2. [task-4b6f9ebd-design-tool-api.md](task-4b6f9ebd-design-tool-api.md) - Tool API specifications
3. [task-f8bb9c5d-define-package-structure.md](task-f8bb9c5d-define-package-structure.md) - Package structure
4. [task-e1x1x1x1-design-service-architecture.md](task-e1x1x1x1-design-service-architecture.md) - Service architecture
5. [task-e2x2x2x2-design-error-handling.md](task-e2x2x2x2-design-error-handling.md) - Error handling
6. [task-e3x3x3x3-design-test-strategy.md](task-e3x3x3x3-design-test-strategy.md) - Test strategy

**Key Decisions for Review**:
- [ ] Tool naming prefix: `opennotes_` (configurable)
- [ ] Pagination: 75% budget + metadata
- [ ] Service architecture: fat services, thin tools
- [ ] Error hints: full installation guide for CLI_NOT_FOUND
- [ ] Test pyramid: 70/25/5 (unit/integration/e2e)

---

## 🔜 Phase 2: Implementation (Pending Approval)

After review, these tasks will be created:

### Core Infrastructure
- [ ] Create `pkgs/pi-opennotes/` directory structure
- [ ] Initialize package.json with pi manifest
- [ ] Implement `CliAdapter` service
- [ ] Implement `PaginationService`
- [ ] Implement error utilities

### Services
- [ ] Implement `SearchService`
- [ ] Implement `ListService`
- [ ] Implement `NoteService`
- [ ] Implement `NotebookService`
- [ ] Implement `ViewsService`

### Tools
- [ ] Implement `opennotes_search` tool
- [ ] Implement `opennotes_list` tool
- [ ] Implement `opennotes_get` tool
- [ ] Implement `opennotes_create` tool
- [ ] Implement `opennotes_notebooks` tool
- [ ] Implement `opennotes_views` tool

### Tests
- [ ] Unit tests for all services (62 tests)
- [ ] Integration tests for all tools (27 tests)
- [ ] E2E tests with real CLI (9 tests)

### Documentation
- [ ] Write README.md
- [ ] Write CHANGELOG.md
- [ ] Create examples

---

## 📋 Future Phases

### Phase 3: Testing & Distribution
- [ ] Final test coverage validation
- [ ] npm publish setup
- [ ] Beta release
- [ ] Documentation review
- [ ] GA release

---

## 🔧 Infrastructure Tasks

### CI/CD Improvements
- [ ] [task-9c4a2f8d-github-actions-moonrepo-releases.md](task-9c4a2f8d-github-actions-moonrepo-releases.md) - GitHub Actions with moonrepo + release-please
  - Integrate moonrepo affected command in CI
  - Setup release-please manifest mode
  - Independent package releases with dependency graph safety

---

## Notes

- **Blocked**: Phase 2 cannot start until Phase 1 is reviewed
- **Recommendation**: Review service interfaces first (task-e1x1x1x1)
- **Time Estimate**: Phase 2 implementation ~4-6 hours
