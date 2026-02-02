---
id: 7c9d2e1f
title: Semantic Search (Optional Enhancement)
created_at: 2026-02-02T21:45:00+10:30
updated_at: 2026-02-02T21:55:00+10:30
status: proposed
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
| Phase | Title | Status | File |
|-------|-------|--------|------|
| 1 | Research & User Story Discovery | 🔜 `proposed` | [phase-52a9f0b3-semantic-search-research.md](phase-52a9f0b3-semantic-search-research.md) |
| 2 | Architecture & Integration Design | ⏳ `planned` | TBD |
| 3 | Implementation & Testing | ⏳ `planned` | TBD |
| 4 | Documentation & Release Guidance | ⏳ `planned` | TBD |

## Dependencies
- chromem-go evaluation and feasibility confirmation
- Existing Bleve search remains default fallback
- Clear opt-in surface (CLI/API) for semantic mode
