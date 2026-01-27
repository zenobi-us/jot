# OpenNotes - Active Tasks

**Status**: 🎯 **READY FOR NEXT EPIC** - Maintenance & Cleanup  
**Current Work**: Documentation updates and bug fixes  
**Last Updated**: 2026-01-26T17:40:00+10:30  
**Status**: ✅ **EPIC CLEANUP COMPLETE** - Memory structure clean

---

## 🔴 High Priority - Implementation Work

### 🚀 Missing View System Features - Phase 1 Complete ✅

1. **[task-3d477ab8]** Implement Missing View System Features (GROUP BY, DISTINCT, OFFSET, HAVING, Aggregations)
   - **Status**: ✅ **PHASE 1 COMPLETE** (2026-01-27)
   - **Priority**: **HIGH** - Group By is critical for dashboards
   - **Phases**: 3 phases (2 + 4 + 2 hours)
   - **Phase 1 (✅ 45 min)**: GROUP BY, DISTINCT, OFFSET - COMPLETE
     - ✅ GROUP BY implementation with field validation
     - ✅ DISTINCT support added
     - ✅ OFFSET for pagination implemented
     - ✅ 8 new tests, all 671+ tests passing
     - ✅ Zero regressions, semantic commit 898d97ec
   - **Phase 2 (✅ COMPLETE)**: HAVING, Aggregations - Full analytics
     - ✅ HAVING clause implementation with condition validation
     - ✅ Aggregate functions (COUNT, SUM, AVG, MAX, MIN) implemented
     - ✅ 13 new tests (exceeds requirement of 9), all 684+ tests passing
     - ✅ Zero regressions, SQL clause ordering validated
     - ✅ Commits: 1657848 (feature), e68465b (docs)
     - Duration: ~1 hour actual vs 4 hours estimate (75% faster!)
   - **Phase 3 (⏳ Ready)**: Templates, env vars - Optional enhancements (2 hrs)
     - Time arithmetic: {{today-N}}, {{today+N}}, etc.
     - Environment variables: {{env:VAR}} syntax
     - Period shortcuts: {{next_week}}, {{next_month}}, etc.
     - 9 new test cases planned
   - **Investigation**: Complete (4 documents available)
   - **Next**: Phase 3 if approved, or proceed to production deployment

## 🟡 High Priority Maintenance

### 📝 Documentation Updates

1. **[task-3f8e2a91]** Update Views Documentation with Correct DuckDB Schema
   - **Status**: 🆕 TODO
   - **Priority**: Medium
   - **Context**: After fixing views DuckDB schema bug, documentation needs updates
   - **Files**: docs/views-guide.md, docs/views-examples.md, docs/views-api.md
   - **Action**: Replace `data.*` references with `metadata->>'*'` syntax

### 🐛 Bug Fixes & Improvements

1. **[task-f4e5d6g7]** Fix Notebook Resolution Order Priority
   - **Status**: 🆕 TODO
   - **Priority**: High
   - **Context**: Resolution order violates principle of least surprise (ignores env var)
   - **Action**: Update `requireNotebook` and `Infer` logic to prioritize EnvVar > Flag > CWD > Context > Ancestor

---

## Potential Next Actions

### Short-term

1. **Continue Storage Abstraction Epic**
   - Research files exist for VFS integration
   - See: epic-a9b3f2c1, task-c81f27bd

2. **Release Preparation**
   - Prepare for version 0.1.0 release with new features

---

## Recently Completed

- ✅ **Basic Getting Started Guide** (2026-01-26): Complete dual-path onboarding
- ✅ **Notebook Creation Fix** (2026-01-26): Use `.` as root for existing directories
- ✅ **Advanced Operations Epic** (2026-01-25): Search, Views, Creation enhancements
