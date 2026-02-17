---
id: 52a9f0b3
title: Phase 1 - Research & User Story Discovery
created_at: 2026-02-02T21:55:00+10:30
updated_at: 2026-02-14T22:52:00+10:30
status: completed
epic_id: 7c9d2e1f
start_criteria: Epic approved for discovery and research scope agreed
end_criteria: Research synthesized into user story candidates with acceptance criteria
---

# Phase 1 - Research & User Story Discovery

## Overview
Identify casual-user problems that semantic search solves (conceptual topic recall), evaluate feasible technical approaches, and translate findings into candidate user stories.

## Deliverables
- Research synthesis focused on casual-user conceptual search use cases
- Recommended user story set with acceptance criteria drafts
- Technical feasibility notes (index storage, embeddings, latency, privacy)

## Research Questions
1. What common “conceptual” queries do casual users attempt where keyword search fails?
2. What minimum recall/precision bar makes semantic results feel helpful (vs. confusing)?
3. How should hybrid merged results be ranked and labeled to build trust?
4. What latency is acceptable for a semantic subcommand on typical notebook sizes?
5. What are the smallest viable embedding model and index size for local usage?
6. What content should be excluded or weighted down (frontmatter, metadata, templates)?
7. How should failures be handled (missing embeddings, cold index, partial results)?
8. What documentation/examples best explain when to use semantic search?

## Tasks
- [task-1a2b3c4d-define-hybrid-merge.md](task-1a2b3c4d-define-hybrid-merge.md)
- [task-2b3c4d5e-explainability-spec.md](task-2b3c4d5e-explainability-spec.md)
- [task-3c4d5e6f-performance-benchmark-plan.md](task-3c4d5e6f-performance-benchmark-plan.md)
- [task-4d5e6f7a-dsl-filter-integration.md](task-4d5e6f7a-dsl-filter-integration.md)
- [task-5e6f7a8b-mode-controls.md](task-5e6f7a8b-mode-controls.md)

## Phase Outcome
- Completed all 5 Phase 1 discovery tasks.
- Produced and validated 5 user stories with acceptance criteria covering hybrid retrieval, explainability, performance targets, DSL filter compatibility, and retrieval mode controls.
- Captured technical recommendations in [research-3f2a9c1b-semantic-search.md](research-3f2a9c1b-semantic-search.md) for Phase 2 architecture handoff.
- Distilled key insights in [learning-b7c2d9e1-semantic-phase1-discovery.md](learning-b7c2d9e1-semantic-phase1-discovery.md).

## Dependencies
- Existing research: chromem-go evaluation notes and vector search synthesis
- Baseline metrics from Bleve search for comparison

## Next Steps
- ✅ Created and approved Phase 2 artifact: [phase-b2f4c8d1-architecture-integration-design.md](phase-b2f4c8d1-architecture-integration-design.md).
- ✅ Broke Phase 1 story outputs into Phase 2 implementation-ready task files.
- Define implementation spike scope for vector index lifecycle and local model selection.
