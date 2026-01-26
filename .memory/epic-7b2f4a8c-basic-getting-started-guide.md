---
id: 7b2f4a8c
title: Create Basic Getting Started Guide for Non-Power Users
created_at: 2026-01-26T13:39:00+10:30
updated_at: 2026-01-26T13:39:00+10:30
status: planning
---

# Create Basic Getting Started Guide for Non-Power Users

## Vision/Goal

Create a **beginner-friendly getting started guide** for users who want to use OpenNotes for basic note management **without diving into SQL**. This guide complements the existing "Getting Started for Power Users" guide and provides a gentler onboarding path focused on:

- Simple note management and organization
- Basic CLI commands (no SQL required initially)
- Simple searches using natural language
- Notebook creation and navigation
- A learning path to SQL (if interested)

**Goal**: Reduce cognitive load for users who just want to "manage my notes" before learning advanced features.

---

## Success Criteria

✅ **Documentation Created**:
- [ ] New `getting-started.md` file created in `pkgs/docs/` directory
- [ ] File follows project style and conventions (see existing guides)
- [ ] All code examples are tested and working

✅ **Content Coverage**:
- [ ] Part 1: Installation and basic setup (5 min)
- [ ] Part 2: Creating your first notebook (5 min)
- [ ] Part 3: Adding and listing notes (5 min)
- [ ] Part 4: Simple searches (without SQL) (5 min)
- [ ] Part 5: Next steps and learning paths (5 min)

✅ **User Experience**:
- [ ] Clear, conversational tone (not technical jargon)
- [ ] Step-by-step instructions with examples
- [ ] Copy-paste ready commands
- [ ] Common mistakes and troubleshooting
- [ ] Clear link to power users guide when ready

✅ **Integration**:
- [ ] Updated `INDEX.md` to reference new basic guide
- [ ] Updated `pkgs/docs/INDEX.md` to reference new basic guide
- [ ] Cross-references to power users guide
- [ ] Synced with main `docs/` directory

✅ **Testing**:
- [ ] All commands tested and working
- [ ] Examples verified with real OpenNotes installation
- [ ] Typos and formatting checked

---

## Phases

1. **Phase 1: Documentation Planning** - Define structure and outline
2. **Phase 2: Content Creation** - Write all 5 parts with examples
3. **Phase 3: Testing & Examples** - Verify all commands work
4. **Phase 4: Integration & Sync** - Update INDEX files and sync with docs/
5. **Phase 5: Final Review** - Polish and prepare for release

---

## Dependencies

- OpenNotes CLI tool (already available)
- Project documentation structure established
- Existing "Getting Started for Power Users" guide
- INDEX.md files for navigation

---

## Timeline

- **Expected Duration**: ~2 hours
- **Phase 1**: 15 minutes (planning)
- **Phase 2**: 60 minutes (writing)
- **Phase 3**: 20 minutes (testing)
- **Phase 4**: 15 minutes (integration)
- **Phase 5**: 10 minutes (review)

---

## Key Notes

- This is **not** a replacement for the power users guide, but a complementary entry point
- Focus on simplicity and confidence-building for first-time users
- Leave SQL for the power users guide
- Provide clear graduation path to more advanced features
- Maintain consistency with existing documentation style

---

## Next Steps

→ Proceed to Phase 1: Documentation Planning
