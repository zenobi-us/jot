# OpenNotes TODO

## Current Status

**Phase 1: Research & Design** - ✅ COMPLETE  
**Next**: Human review, then Phase 2 (Implementation)

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

## Notes

- **Blocked**: Phase 2 cannot start until Phase 1 is reviewed
- **Recommendation**: Review service interfaces first (task-e1x1x1x1)
- **Time Estimate**: Phase 2 implementation ~4-6 hours
