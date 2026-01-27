# OpenNotes - Active Tasks

**Status**: 🎯 **READY FOR NEXT EPIC** - Maintenance & Cleanup  
**Current Work**: Documentation updates and bug fixes  
**Last Updated**: 2026-01-26T17:40:00+10:30  
**Status**: ✅ **EPIC CLEANUP COMPLETE** - Memory structure clean

---

## 🔴 High Priority - Implementation Work

### 🚀 Missing View System Features (NEW)

1. **[task-3d477ab8]** Implement Missing View System Features (GROUP BY, DISTINCT, OFFSET, HAVING, Aggregations)
   - **Status**: 🆕 PLANNING
   - **Priority**: **HIGH** - Group By is critical for dashboards
   - **Phases**: 3 phases (2 + 4 + 2 hours)
   - **Phase 1 (2 hrs)**: GROUP BY, DISTINCT, OFFSET - Immediate impact
   - **Phase 2 (4 hrs)**: HAVING, Aggregations - Full analytics
   - **Phase 3 (2 hrs)**: Templates, env vars - Optional enhancements
   - **Investigation**: Complete (4 documents, /tmp/)
   - **Next**: Begin Phase 1 implementation

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
