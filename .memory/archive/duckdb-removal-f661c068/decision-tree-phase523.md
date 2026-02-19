# Phase 5.2.3 Decision Tree

## Can we migrate SearchWithConditions() to Bleve?

```
START: Phase 5.2.3 - Migrate SearchWithConditions()
│
├─ Q: Can metadata fields be migrated?
│  └─ ✅ YES → data.* fields map to metadata.field in Bleve
│     └─ Supports: tag, status, priority, assignee, author, type, category, project, sprint
│
├─ Q: Can path fields be migrated?
│  └─ ⚠️ PARTIAL → Simple globs via PrefixQuery, complex via WildcardQuery
│     ├─ ✅ Fast: path=projects/* → PrefixQuery("projects/")
│     └─ ⚠️ Slower: path=**/tasks/*.md → WildcardQuery
│
├─ Q: Can title fields be migrated?
│  └─ ✅ YES → Direct mapping to title field
│
├─ Q: Can AND/OR/NOT logic be migrated?
│  └─ ✅ YES → Full support via BooleanQuery
│     ├─ AND → ConjunctionQuery
│     ├─ OR → DisjunctionQuery
│     └─ NOT → BooleanQuery with mustNot
│
├─ Q: Can link queries be migrated?
│  └─ ❌ NO → Requires graph index (not in Bleve)
│     ├─ links-to → Needs JOIN/subquery
│     ├─ linked-by → Needs reverse lookup
│     └─ Solution: Defer to Phase 5.3
│
├─ Q: Will this break existing functionality?
│  └─ ⚠️ PARTIAL BREAKING CHANGE
│     ├─ ✅ Metadata queries: No change
│     ├─ ✅ Path queries: No change (may be faster)
│     ├─ ✅ Title queries: No change
│     └─ ❌ Link queries: WILL BREAK
│        └─ Mitigation: Clear error + SQL workaround
│
├─ Q: Can we provide a workaround?
│  └─ ✅ YES → SQL query interface remains
│     └─ opennotes notes query "SELECT * FROM ..."
│
├─ Q: When will link queries work?
│  └─ ⏳ Phase 5.3 → Link graph index implementation
│     └─ Estimated: 14-19 hours
│
├─ Q: Is the migration worth it?
│  └─ ✅ YES
│     ├─ Removes DuckDB dependency from main query path
│     ├─ Improves performance for metadata/path queries
│     ├─ Simplifies codebase (200+ lines → 20 lines)
│     ├─ Clear migration path for link queries
│     └─ Breaking change is documented and mitigated
│
└─ DECISION: ✅ PROCEED WITH MIGRATION
   └─ Implementation Plan: .memory/plan-phase523-implementation.md
```

## Query Type Decision Tree

```
User Query: opennotes notes search query --and <field>=<value>
│
├─ Field = "data.*" (metadata)
│  └─ ✅ MIGRATE
│     └─ Implementation: FieldExpr{Field: "metadata.field", Op: OpEquals, Value: value}
│
├─ Field = "path"
│  ├─ Value has no wildcards?
│  │  └─ ✅ MIGRATE → FieldExpr with OpEquals
│  ├─ Value = "prefix/*"?
│  │  └─ ✅ MIGRATE → FieldExpr with OpPrefix (FAST)
│  └─ Value has complex wildcards (**/*, mid-pattern)?
│     └─ ⚠️ MIGRATE → WildcardExpr (SLOWER)
│
├─ Field = "title"
│  └─ ✅ MIGRATE → FieldExpr{Field: "title", Op: OpEquals, Value: value}
│
├─ Field = "links-to"
│  └─ ❌ ERROR → "Link queries not yet supported (Phase 5.3)"
│
└─ Field = "linked-by"
   └─ ❌ ERROR → "Link queries not yet supported (Phase 5.3)"
```

## Test Migration Decision Tree

```
Test: TestNoteService_SearchWithConditions_*
│
├─ Uses DuckDB?
│  └─ ✅ YES → Needs migration
│     ├─ Step 1: Remove NewDbService()
│     ├─ Step 2: Add testutil.CreateTestIndex(t, notebookDir)
│     ├─ Step 3: Update NewNoteService(cfg, nil, index, notebookDir)
│     └─ Step 4: Run test → Should pass
│
├─ Tests link queries?
│  └─ ⚠️ YES → Special handling
│     ├─ Option A: Skip test with Phase 5.3 reference
│     │  └─ t.Skip("Link queries deferred to Phase 5.3")
│     └─ Option B: Test error message
│        └─ require.Error(t, err)
│        └─ assert.Contains(t, err.Error(), "Phase 5.3")
│
└─ Test assertions still valid?
   ├─ ✅ YES → Keep assertions unchanged
   └─ ❌ NO → Update for Bleve behavior
```

## Performance Decision Tree

```
Query Performance Concern?
│
├─ Query Type = Metadata field?
│  └─ ✅ FASTER in Bleve
│     └─ Reason: Inverted index vs SQL WHERE clause
│
├─ Query Type = Simple path prefix?
│  └─ ✅ FASTER in Bleve
│     └─ Reason: PrefixQuery vs LIKE with %
│
├─ Query Type = Complex path wildcard?
│  └─ ⚠️ SIMILAR or SLOWER
│     └─ Reason: WildcardQuery scans index
│     └─ Mitigation: Recommend prefix patterns
│
├─ Query Type = Link query?
│  └─ ❌ N/A → Not supported yet
│     └─ Phase 5.3: Will be FASTER with graph index
│
└─ Overall?
   └─ ✅ MAINTAINED or IMPROVED
```

## Risk Acceptance Decision Tree

```
Should we accept the risks and proceed?
│
├─ Risk: Link queries will break
│  ├─ Severity: HIGH (users may depend on this)
│  ├─ Mitigation: Clear error + workaround
│  ├─ Timeline: Phase 5.3 (planned)
│  └─ Accept? ✅ YES
│     └─ Reason: Workaround available, clear migration path
│
├─ Risk: Path globs may be slower
│  ├─ Severity: MEDIUM (performance concern)
│  ├─ Mitigation: Optimize prefix, document wildcards
│  ├─ Impact: Only complex patterns affected
│  └─ Accept? ✅ YES
│     └─ Reason: Most queries use prefixes, docs guide users
│
├─ Risk: Test migration effort
│  ├─ Severity: LOW (just time)
│  ├─ Mitigation: Incremental updates, testutil helper
│  ├─ Impact: 40 tests to update
│  └─ Accept? ✅ YES
│     └─ Reason: Straightforward pattern replacement
│
└─ Overall Risk Profile
   └─ ✅ ACCEPTABLE
      └─ All risks mitigated with clear plan
```

## Implementation Order Decision

```
Which implementation order is best?
│
├─ Option A: Top-down (SearchWithConditions first)
│  ├─ Pros: See end goal quickly
│  ├─ Cons: Can't test until BuildQuery() exists
│  └─ Verdict: ❌ NOT RECOMMENDED
│
├─ Option B: Bottom-up (BuildQuery first)
│  ├─ Pros: Test each piece independently
│  ├─ Pros: Incremental validation
│  ├─ Cons: Longer to see end result
│  └─ Verdict: ✅ RECOMMENDED
│     └─ Order: BuildQuery → Tests → SearchWithConditions → Integration
│
└─ Chosen: Bottom-up (Plan Phase 1-5)
```

## Breaking Change Communication Decision

```
How should we communicate the link query breaking change?
│
├─ In Error Message?
│  └─ ✅ YES
│     ├─ What broke: "Link queries not yet supported"
│     ├─ Why: "Requires dedicated link graph index (Phase 5.3)"
│     ├─ Workaround: "Use SQL query interface"
│     └─ Tracking: "github.com/zenobi-us/opennotes/issues/XXX"
│
├─ In CHANGELOG.md?
│  └─ ✅ YES
│     ├─ Section: "## Breaking Changes"
│     ├─ Commands affected
│     ├─ Workaround
│     └─ Timeline
│
├─ In Documentation?
│  └─ ✅ YES
│     ├─ Update: docs/commands/notes-search.md
│     ├─ Add: "Link Queries (Coming in Phase 5.3)" section
│     └─ Examples: Workaround SQL queries
│
├─ In GitHub Issue?
│  └─ ✅ YES
│     └─ Create: Phase 5.3 - Link Graph Index implementation
│
└─ Verdict: ✅ OVER-COMMUNICATE
   └─ Users see the message everywhere, can't miss it
```

## Final Decision

```
Should we proceed with Phase 5.2.3?

Checklist:
├─ ✅ Core functionality migratable (11/13 fields)
├─ ✅ Breaking changes documented and mitigated
├─ ✅ Clear workaround available
├─ ✅ Future plan exists (Phase 5.3)
├─ ✅ Performance maintained or improved
├─ ✅ Tests can be migrated
├─ ✅ Implementation plan is detailed
└─ ✅ Time estimate is reasonable (8-11 hours)

DECISION: ✅ PROCEED

Next Action: Begin Phase 1 - Implement BuildQuery()
Reference: .memory/plan-phase523-implementation.md
```

## Quick Reference

**Documents**:
- Assessment: `.memory/assessment-phase523-migration.md` (24KB)
- Plan: `.memory/plan-phase523-implementation.md` (33KB)
- Summary: `.memory/summary-phase523-migration.md` (8KB)
- Decision Tree: `.memory/decision-tree-phase523.md` (this file)

**Status**: ✅ Ready for implementation

**Estimated Time**: 8-11 hours

**Risk Level**: ⚠️ Medium (link queries deferred)

**Recommendation**: Proceed with migration
