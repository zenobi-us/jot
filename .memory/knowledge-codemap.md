---
id: a1b2c3d4
title: OpenNotes Codebase Structure Map
created_at: 2026-01-18T19:31:53+10:30
updated_at: 2026-02-01T21:25:00+10:30
status: active
area: codebase-structure
tags: [architecture, codebase, state-machine, bleve, search]
learned_from: [test-improvement-epic, codebase-exploration, architecture-review, epic-f661c068]
---

# OpenNotes Codebase Structure Map

## Overview

OpenNotes is a CLI tool for managing markdown-based notes organized in notebooks. **Currently transitioning from DuckDB to pure Go search** (Bleve + Parser). Templates used for display.

## Details

### ASCII State Machine Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           OPENNOTES CLI TOOL                           │
│                         (Go-based Architecture)                        │
└─────────────────────────────────────────────────────────────────────────┘

                                   [main.go]
                                       │
                                       ▼
                              ┌─────────────────┐
                              │   cmd/root.go   │
                              │ (Service Init)  │
                              │                 │
                              │ ┌─ ConfigSvc   │
                              │ ├─ DbSvc       │
                              │ ├─ NotebookSvc │
                              │ ├─ NoteSvc     │
                              │ ├─ DisplaySvc  │
                              │ └─ LoggerSvc   │
                              └─────────────────┘
                                       │
                         ┌─────────────┼─────────────┐
                         ▼             ▼             ▼
                ┌─────────────┐ ┌──────────────┐ ┌──────────────┐
                │    init     │ │   notebook   │ │    notes     │
                │  commands   │ │   commands   │ │   commands   │
                │             │ │              │ │              │
                │ • init      │ │ • list       │ │ • add        │
                │             │ │ • info       │ │ • list       │
                │             │ │ • switch     │ │ • search     │
                │             │ │              │ │ • show       │
                └─────────────┘ └──────────────┘ └──────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                           SERVICE LAYER                                │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Config Service  │    │ Notebook Svc    │    │  Note Service   │
│                 │    │                 │    │ (LEGACY/DuckDB) │
│ • LoadConfig    │◀───│ • Discover      │◀───│ • SearchNotes   │
│ • SaveConfig    │    │ • LoadConfig    │    │ • GetNote       │
│ • GetNotebooks  │    │ • Validate      │    │ • ExtractMeta   │
│                 │    │                 │    │ • DisplayName   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Database Svc   │    │  Display Svc    │    │  Logger Svc     │
│ (LEGACY/DuckDB) │    │                 │    │                 │
│ • GetReadDB     │    │ • TuiRender     │    │ • Info/Error    │
│ • GetWriteDB    │    │ • RenderSQL     │    │ • Debug/Warn    │
│ • CloseAll      │    │ • Templates     │    │ • WithField     │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                   NEW SEARCH SYSTEM (Phase 4 Complete)                 │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   internal/search/                             │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐    │
│  │   Index      │  │   Query      │  │   FindOpts        │    │
│  │ (interface)  │  │   (AST)      │  │ (functional opts) │    │
│  └──────────────┘  └──────────────┘  └───────────────────┘    │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐    │
│  │  Document    │  │   Results    │  │   Storage         │    │
│  │  (metadata)  │  │ (search res) │  │   (afero)         │    │
│  └──────────────┘  └──────────────┘  └───────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌────────────────────┐  ┌────────────────────┐  ┌────────────────┐
│  internal/search/  │  │  internal/search/  │  │ spf13/afero    │
│  parser/           │  │  bleve/            │  │ (filesystem)   │
│                    │  │                    │  │                │
│ • Parser           │  │ • Index impl       │  │ • MemMapFs     │
│ • Grammar (EBNF)   │  │ • Mapping (BM25)   │  │ • OsFs         │
│ • Convert (AST)    │  │ • Query translate  │  │ • Storage if   │
│ • 10 tests         │  │ • 36 tests         │  │                │
└────────────────────┘  └────────────────────┘  └────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                    NEW SEARCH STATE MACHINE                            │
└─────────────────────────────────────────────────────────────────────────┘

Index Lifecycle:
[Unopened] → [NewIndex()] → [Ready]
     │              │           │
     ▼              ▼           ▼
[Error] ←─── [Init Failed] [Reindex()] → [Indexing] → [Ready]
     │                          │              │
     │                          ▼              ▼
     └────────────────── [Walk Failed] ←── [Add Failed]

Search Flow:
[Query String] → [Parser.Parse()] → [Query AST] → [TranslateQuery()] → [Bleve Query]
       │              │                  │               │                    │
       ▼              ▼                  ▼               ▼                    ▼
[Gmail DSL] → [Participle Grammar] → [Convert] → [FindOpts merge] → [Index.Search()]
                                                                            │
                                                                            ▼
                                                                      [Extract Results]
                                                                            │
                                                                            ▼
                                                                      [search.Results]

Document Flow:
[Note File] → [Read frontmatter] → [Document] → [BleveDocument] → [Index.Add()]
     │              │                   │              │                │
     ▼              ▼                   ▼              ▼                ▼
[Markdown] → [Parse YAML] → [Field extraction] → [Apply weights] → [Bleve index]

BM25 Weighting:
Path: 1000 → Title: 500 → Tags: 300 → Lead: 50 → Body: 1

┌─────────────────────────────────────────────────────────────────────────┐
│                    LEGACY DATA FLOW (TO BE REMOVED)                    │
└─────────────────────────────────────────────────────────────────────────┘

[Start] → [Find Notebook] → [Load Config] → [Initialize Services] → [Execute Command]
   │            │               │                    │                │
   │            ▼               ▼                    ▼                ▼
   │      [Ancestor Search] [JSON Parse]      [DuckDB Connect]  [Parse Args]
   │            │               │                    │                │
   │            ▼               ▼                    ▼                ▼
   │      [Config Override] [Validate]         [Markdown Ext]   [Command Logic]
   │            │               │                    │                │
   │            ▼               ▼                    ▼                ▼
   └──────> [Notebook Ready] [Services Ready] [Database Ready] [Execute & Render]
                                   │
                                   ▼
                              [Template Render] → [Glamour Output] → [Success/Error]

┌─────────────────────────────────────────────────────────────────────────┐
│                         LIFECYCLE STATES                               │
└─────────────────────────────────────────────────────────────────────────┘

Notebook Lifecycle:
[Uninitialized] → [init command] → [.opennotes.json created] → [Ready]
       │                               │                         │
       ▼                               ▼                         ▼
[Error: No Config] ←──────── [JSON Error] ←──────── [Config Operations]

Note Lifecycle:
[Template Selected] → [Create File] → [Edit Content] → [Save] → [Indexed by DuckDB]
       │                    │             │            │              │
       ▼                    ▼             ▼            ▼              ▼
[Template Render] → [File Write] → [User Editor] → [Disk Sync] → [Search Ready]

Search Lifecycle:
[Query Input] → [SQL Validation] → [DuckDB Execute] → [Format Results] → [Display]
      │               │                  │                │               │
      ▼               ▼                  ▼                ▼               ▼
[Parse Args] → [Security Check] → [Read-Only Conn] → [Table Format] → [Terminal]
```

### Component Relationships

```
CLI Commands (cmd/)
    ├── Thin orchestration layer
    ├── Parse flags → Call services → Render output
    └── Max 50-125 lines per command

Internal Services (internal/services/)
    ├── ConfigService: Global settings & notebook registry
    ├── DbService: DuckDB connections (read/write isolation)
    ├── NotebookService: Discovery, validation, lifecycle
    ├── NoteService: SQL queries, metadata extraction
    ├── DisplayService: Template rendering, table formatting
    └── LoggerService: Structured logging (zap)

Core Utilities (internal/core/)
    ├── Validation: Input sanitization, path checking
    └── Utils: String manipulation, slugification

Test Structure (tests/)
    ├── Unit tests: *_test.go alongside source
    ├── Integration: Service interaction tests
    ├── E2E: Full command execution tests
    └── Performance: Stress tests, benchmarks
```

### Key Patterns

1. **Singleton Services**: Initialized once in cmd/root.go
2. **Read-Only Database**: Separate connections for safety
3. **Template-Driven Output**: go:embed templates with glamour
4. **Defense in Depth**: Validation at multiple layers
5. **Error Propagation**: Explicit error handling throughout
6. **Service-Oriented**: Fat services, thin commands

### Quality Metrics

- **Files**: 79 Go files, 307KB total
- **Tests**: 202+ test functions, 84%+ coverage
- **Performance**: Sub-100ms for typical operations
- **Architecture**: Clean separation of concerns
- **Status**: Production-ready, enterprise-validated

### Planned: Pi-OpenNotes Extension

```
┌────────────────────────────────────────────────────────────────────────┐
│                     PI-OPENNOTES EXTENSION (Planned)                   │
│                         pkgs/pi-opennotes/                             │
└────────────────────────────────────────────────────────────────────────┘

                              [Pi Agent]
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│              pi-opennotes Extension (Bun/TS)               │
│                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐    │
│  │ search_notes │ │ list_notes   │ │ get_note         │    │
│  └──────┬───────┘ └──────┬───────┘ └────────┬─────────┘    │
│         │                │                   │              │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐    │
│  │ create_note  │ │ notebooks    │ │ views            │    │
│  └──────┬───────┘ └──────┬───────┘ └────────┬─────────┘    │
│         │                │                   │              │
│         └────────────────┼───────────────────┘              │
│                          ▼                                  │
│                 ┌─────────────────┐                         │
│                 │   CLI Adapter   │ ← Executes via shell    │
│                 └────────┬────────┘                         │
└──────────────────────────┼──────────────────────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │  opennotes CLI  │ ← Go binary
                  │  (dist/opennotes)│
                  └────────┬────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │   DuckDB +      │
                  │   Markdown Ext  │
                  └─────────────────┘
```