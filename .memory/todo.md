# OpenNotes - Active Tasks

**Status**: 🎯 **READY FOR NEXT EPIC** - Maintenance & Cleanup  
**Current Work**: Documentation updates and bug fixes  
**Last Updated**: 2026-01-26T17:40:00+10:30  
**Status**: ✅ **EPIC CLEANUP COMPLETE** - Memory structure clean

---

## 🔴 High Priority - Implementation Work

### 🚀 Missing View System Features - Phase 1-3 Complete ✅✅✅

1. **[task-3d477ab8]** Implement Missing View System Features (GROUP BY, DISTINCT, OFFSET, HAVING, Aggregations, Templates)
   - **Status**: ✅ **COMPLETE (ALL 3 PHASES)** (2026-01-27)
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
   - **Phase 3 (✅ COMPLETE)**: Templates, env vars - Dynamic date/env support
     - ✅ Time arithmetic: {{today-N}}, {{today+N}}, {{this_week-N}}, {{this_month-N}}
     - ✅ Period shortcuts: {{next_week}}, {{next_month}}, {{last_week}}, {{last_month}}
     - ✅ Quarter/Year: {{start_of_quarter}}, {{end_of_quarter}}, {{quarter}}, {{year}}
     - ✅ Month boundaries: {{start_of_month}}, {{end_of_month}}
     - ✅ Environment variables: {{env:VAR}}, {{env:DEFAULT:VAR}}
     - ✅ 27 new tests (exceeds requirement of 9), all 711+ tests passing
     - ✅ Zero regressions, helper functions for date calculations
     - ✅ Commit: 10a7017 (feat: phase 3 implementation)
     - Duration: ~45 minutes actual vs 2 hours estimate (78% faster!)
   - **Investigation**: Complete (4 documents available)
   - **Total Duration**: ~1.5 hours actual vs 8 hours estimate (81% faster!)
   - **Next**: Ready for production deployment or next epic

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
