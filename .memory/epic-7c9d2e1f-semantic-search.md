---
id: 7c9d2e1f
title: Semantic Search (Optional Enhancement)
created_at: 2026-02-02T21:45:00+10:30
updated_at: 2026-02-15T01:30:00+10:30
status: in-progress
---

# Semantic Search (Optional Enhancement)

## Vision/Goal
Introduce an optional semantic search capability that augments existing Bleve full-text search with vector-based relevance, improving recall for conceptual queries and paraphrases.

## Success Criteria
- Semantic search subcommand (dedicated mode) implemented with hybrid merged results (semantic + keyword)
- Vector index lifecycle is reliable and testable
- Documentation explains when to use semantic vs. full-text search
- Benchmarks show acceptable latency for typical notebook sizes

## Phases

**Progress**: 75% complete (3 of 4 phases complete)

| Phase | Title | Status | File |
|-------|-------|--------|------|
| 1 | Research & User Story Discovery | ✅ `completed` | [phase-52a9f0b3-semantic-search-research.md](phase-52a9f0b3-semantic-search-research.md) |
| 2 | Architecture & Integration Design | ✅ `completed` | [phase-b2f4c8d1-architecture-integration-design.md](phase-b2f4c8d1-architecture-integration-design.md) |
| 3 | Implementation & Testing | ✅ `completed` | [phase-91d3f6a2-implementation-testing.md](phase-91d3f6a2-implementation-testing.md) |
| 4 | Documentation & Release Guidance | ⏳ `planned` | TBD |

## Dependencies
- chromem-go evaluation and feasibility confirmation
- Existing Bleve search remains default fallback
- Clear opt-in surface (CLI/API) for semantic mode
