# OpenNotes - Active Tasks

**Status**: 🎯 **READY FOR NEXT EPIC** - Memory Cleanup Complete  
**Current Work**: Bug fixes and maintenance  
**Last Updated**: 2026-01-28T22:15:00+10:30  
**Status**: ✅ **MEMORY CLEANUP COMPLETE** - Structure clean

---

## 🟡 Active Tasks

### 🐛 Bug Fixes & Improvements

1. **[task-f4e5d6g7]** Fix Notebook Resolution Order Priority
   - **Status**: 🆕 TODO
   - **Priority**: High
   - **Context**: Resolution order violates principle of least surprise (ignores env var)
   - **Action**: Update `requireNotebook` and `Infer` logic to prioritize EnvVar > Flag > CWD > Context > Ancestor
   - **Estimate**: 1-2 hours

---

## Potential Next Actions

### Short-term

1. **Continue Storage Abstraction Epic**
   - Research files exist for VFS integration
   - See: epic-a9b3f2c1, task-c81f27bd

2. **Release Preparation**
   - Prepare for version 0.1.0 release with new features

---

## Recently Completed (Archived 2026-01-28)

- ✅ **Missing View System Features** (task-3d477ab8): GROUP BY, DISTINCT, OFFSET, HAVING, Aggregations, Templates
- ✅ **Views Documentation Updates** (task-3f8e2a91): DuckDB schema corrections
- ✅ **Views Fault Tolerance Investigation** (task-b2d67264): Schema bug root cause analysis
- ✅ **Option 2 Refactor** (learning-c1d2e3f4): GroupResults return structure refactored

---

## Archive Summary

Completed tasks moved to: `archive/views-system-tasks-2026-01-28/`
- task-3d477ab8-missing-view-system-features.md
- task-3f8e2a91-views-documentation-updates.md  
- task-b2d67264-views-fault-tolerance-investigation.md
