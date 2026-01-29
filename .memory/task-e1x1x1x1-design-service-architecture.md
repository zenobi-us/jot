---
id: e1x1x1x1
title: Design Service Architecture
created_at: 2026-01-29T09:30:00+10:30
updated_at: 2026-01-29T09:30:00+10:30
status: done
epic_id: 1f41631e
phase_id: 43842f12
assigned_to: null
---

# Design Service Architecture

## Objective

Design the service layer architecture for pi-opennotes, defining service interfaces, dependencies, and how tools wrap services following SOLID principles.

## Completed Steps

- [x] Define service interfaces
- [x] Design CliAdapter interface
- [x] Design SearchService
- [x] Design ListService
- [x] Design NoteService
- [x] Design NotebookService
- [x] Design ViewsService
- [x] Design PaginationService
- [x] Document service composition
- [x] Create dependency injection pattern

---

## SOLID Principles Application

| Principle | Application in pi-opennotes |
|-----------|----------------------------|
| **Single Responsibility** | Each service handles one domain (search, notes, views, etc.) |
| **Open/Closed** | Services extend via composition, not modification |
| **Liskov Substitution** | All services implement interfaces, can be swapped |
| **Interface Segregation** | Small, focused interfaces per concern |
| **Dependency Inversion** | Services depend on abstractions (CliAdapter interface) |

---

## Service Interfaces

### CliAdapter Interface

The foundation for all CLI interactions:

```typescript
// src/services/types.ts

export interface CliResult {
  code: number;
  stdout: string;
  stderr: string;
  timedOut?: boolean;
}

export interface CliOptions {
  notebook?: string;
  timeout?: number;
  signal?: AbortSignal;
  env?: Record<string, string>;
}

export interface ICliAdapter {
  /**
   * Execute CLI command
   */
  exec(
    command: string,
    args: string[],
    options?: CliOptions
  ): Promise<CliResult>;

  /**
   * Check if CLI is installed and accessible
   */
  checkInstallation(): Promise<{ installed: boolean; version?: string; path?: string }>;

  /**
   * Parse JSON output with error handling
   */
  parseJsonOutput<T>(stdout: string): T;

  /**
   * Build notebook flag args
   */
  buildNotebookArgs(notebook?: string): string[];
}
```

### SearchService Interface

```typescript
// src/services/types.ts

export interface SearchOptions {
  notebook?: string;
  limit?: number;
  offset?: number;
  signal?: AbortSignal;
}

export interface BooleanFilters {
  and?: string[];
  or?: string[];
  not?: string[];
}

export interface SearchResult<T = NoteSummary> {
  results: T[];
  query: {
    type: "text" | "fuzzy" | "sql" | "boolean";
    executed: string;
  };
  pagination: PaginationMeta;
}

export interface ISearchService {
  /**
   * Text-based search (exact substring)
   */
  textSearch(query: string, options: SearchOptions): Promise<SearchResult>;

  /**
   * Fuzzy search (typo-tolerant, ranked)
   */
  fuzzySearch(query: string, options: SearchOptions): Promise<SearchResult>;

  /**
   * Raw SQL query execution
   */
  sqlSearch(sql: string, options: SearchOptions): Promise<SearchResult<Record<string, unknown>>>;

  /**
   * Boolean query (AND/OR/NOT filters)
   */
  booleanSearch(filters: BooleanFilters, options: SearchOptions): Promise<SearchResult>;
}
```

### ListService Interface

```typescript
export interface ListOptions {
  notebook?: string;
  sortBy?: "modified" | "created" | "title" | "path";
  sortOrder?: "asc" | "desc";
  pattern?: string;
  limit?: number;
  offset?: number;
  signal?: AbortSignal;
}

export interface ListResult {
  notes: NoteSummary[];
  notebook: NotebookInfo;
  pagination: PaginationMeta;
}

export interface IListService {
  /**
   * List notes with optional filtering and sorting
   */
  listNotes(options: ListOptions): Promise<ListResult>;

  /**
   * Count total notes (for pagination)
   */
  countNotes(options: Pick<ListOptions, "notebook" | "pattern">): Promise<number>;
}
```

### NoteService Interface

```typescript
export interface GetOptions {
  notebook?: string;
  includeContent?: boolean;
  signal?: AbortSignal;
}

export interface CreateOptions {
  notebook?: string;
  path?: string;
  template?: string;
  content?: string;
  data?: Record<string, string | number | boolean | string[]>;
  signal?: AbortSignal;
}

export interface GetResult {
  note: NoteContent;
  notebook: NotebookInfo;
}

export interface CreateResult {
  created: {
    path: string;
    absolutePath: string;
    title: string;
  };
  notebook: NotebookInfo;
}

export interface INoteService {
  /**
   * Get a specific note by path
   */
  getNote(path: string, options: GetOptions): Promise<GetResult>;

  /**
   * Create a new note
   */
  createNote(title: string, options: CreateOptions): Promise<CreateResult>;

  /**
   * Check if a note exists
   */
  noteExists(path: string, options: Pick<GetOptions, "notebook">): Promise<boolean>;
}
```

### NotebookService Interface

```typescript
export interface INotebookService {
  /**
   * List all available notebooks
   */
  listNotebooks(): Promise<{
    notebooks: NotebookInfo[];
    current: NotebookInfo | null;
  }>;

  /**
   * Get current notebook (from context or explicit)
   */
  getCurrentNotebook(explicitPath?: string): Promise<NotebookInfo | null>;

  /**
   * Validate a notebook path
   */
  validateNotebook(path: string): Promise<{
    valid: boolean;
    error?: string;
  }>;
}
```

### ViewsService Interface

```typescript
export interface ViewExecuteOptions {
  notebook?: string;
  params?: Record<string, string>;
  limit?: number;
  offset?: number;
  signal?: AbortSignal;
}

export interface ViewsListResult {
  views: ViewDefinition[];
  notebook: NotebookInfo;
}

export interface ViewExecuteResult {
  view: {
    name: string;
    description: string;
  };
  results: Record<string, unknown>[];
  pagination: PaginationMeta;
  notebook: NotebookInfo;
}

export interface IViewsService {
  /**
   * List all available views
   */
  listViews(notebook?: string): Promise<ViewsListResult>;

  /**
   * Execute a named view
   */
  executeView(name: string, options: ViewExecuteOptions): Promise<ViewExecuteResult>;

  /**
   * Get a specific view definition
   */
  getView(name: string, notebook?: string): Promise<ViewDefinition | null>;
}
```

### PaginationService Interface

```typescript
export interface PaginationInput<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface IPaginationService {
  /**
   * Apply pagination to results
   */
  paginate<T>(input: PaginationInput<T>): {
    items: T[];
    pagination: PaginationMeta;
  };

  /**
   * Calculate if output exceeds budget
   */
  exceedsBudget(content: string, budgetRatio?: number): boolean;

  /**
   * Truncate items to fit budget
   */
  fitToBudget<T>(
    items: T[],
    serialize: (item: T) => string,
    budgetRatio?: number
  ): {
    items: T[];
    truncated: boolean;
    originalCount: number;
  };
}
```

---

## Service Implementations

### CliAdapter Implementation

```typescript
// src/services/cli-adapter.ts

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import type { ICliAdapter, CliResult, CliOptions } from "./types";
import { OpenNotesError, ErrorCodes } from "../utils/errors";

export interface CliAdapterConfig {
  cliPath: string;
  defaultTimeout: number;
}

export class CliAdapter implements ICliAdapter {
  constructor(
    private pi: ExtensionAPI,
    private config: CliAdapterConfig
  ) {}

  async exec(command: string, args: string[], options?: CliOptions): Promise<CliResult> {
    const fullArgs = [...args];
    
    // Add notebook flag if specified
    if (options?.notebook) {
      fullArgs.push("--notebook", options.notebook);
    }

    try {
      const result = await this.pi.exec(command, fullArgs, {
        signal: options?.signal,
        timeout: options?.timeout ?? this.config.defaultTimeout,
        env: options?.env,
      });

      return {
        code: result.code,
        stdout: result.stdout,
        stderr: result.stderr,
        timedOut: false,
      };
    } catch (error) {
      if (error instanceof Error && error.name === "AbortError") {
        return {
          code: -1,
          stdout: "",
          stderr: "Command aborted",
          timedOut: true,
        };
      }
      throw error;
    }
  }

  async checkInstallation(): Promise<{ installed: boolean; version?: string; path?: string }> {
    try {
      const result = await this.exec(this.config.cliPath, ["version"], { timeout: 5000 });
      
      if (result.code === 0) {
        const versionMatch = result.stdout.match(/opennotes version (\S+)/);
        return {
          installed: true,
          version: versionMatch?.[1],
          path: this.config.cliPath,
        };
      }
      
      return { installed: false };
    } catch {
      return { installed: false };
    }
  }

  parseJsonOutput<T>(stdout: string): T {
    try {
      return JSON.parse(stdout) as T;
    } catch (error) {
      throw new OpenNotesError(
        `Failed to parse CLI output as JSON: ${error}`,
        ErrorCodes.PARSE_ERROR,
        { stdout: stdout.slice(0, 200) }
      );
    }
  }

  buildNotebookArgs(notebook?: string): string[] {
    return notebook ? ["--notebook", notebook] : [];
  }
}
```

### SearchService Implementation

```typescript
// src/services/search.service.ts

import type { ISearchService, ICliAdapter, IPaginationService, SearchOptions, SearchResult, BooleanFilters } from "./types";
import type { NoteSummary } from "../schemas/note";
import { OpenNotesError, ErrorCodes } from "../utils/errors";

export class SearchService implements ISearchService {
  constructor(
    private cli: ICliAdapter,
    private pagination: IPaginationService
  ) {}

  async textSearch(query: string, options: SearchOptions): Promise<SearchResult> {
    // Convert text search to SQL for consistent JSON output
    const escapedQuery = query.replace(/'/g, "''");
    const sql = `
      SELECT file_path, metadata, content
      FROM read_markdown('**/*.md')
      WHERE content LIKE '%${escapedQuery}%' OR metadata->>'title' LIKE '%${escapedQuery}%'
      ORDER BY file_path
      LIMIT ${options.limit ?? 50}
      OFFSET ${options.offset ?? 0}
    `;
    
    return this.executeSqlInternal(sql, "text", options);
  }

  async fuzzySearch(query: string, options: SearchOptions): Promise<SearchResult> {
    const result = await this.cli.exec(
      "opennotes",
      ["notes", "search", "--fuzzy", query, ...this.cli.buildNotebookArgs(options.notebook)],
      { signal: options.signal }
    );

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Fuzzy search failed: ${result.stderr}`,
        ErrorCodes.SEARCH_FAILED,
        { query }
      );
    }

    // Parse glamour output (not JSON) - convert to NoteSummary[]
    const notes = this.parseTextSearchOutput(result.stdout);
    
    return {
      results: notes,
      query: { type: "fuzzy", executed: `--fuzzy "${query}"` },
      pagination: this.pagination.paginate({
        items: notes,
        total: notes.length,
        limit: options.limit ?? 50,
        offset: options.offset ?? 0,
      }).pagination,
    };
  }

  async sqlSearch(sql: string, options: SearchOptions): Promise<SearchResult<Record<string, unknown>>> {
    return this.executeSqlInternal(sql, "sql", options);
  }

  async booleanSearch(filters: BooleanFilters, options: SearchOptions): Promise<SearchResult> {
    const args = ["notes", "search", "query"];
    
    for (const condition of filters.and ?? []) {
      args.push("--and", condition);
    }
    for (const condition of filters.or ?? []) {
      args.push("--or", condition);
    }
    for (const condition of filters.not ?? []) {
      args.push("--not", condition);
    }
    
    args.push(...this.cli.buildNotebookArgs(options.notebook));

    const result = await this.cli.exec("opennotes", args, { signal: options.signal });

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Boolean search failed: ${result.stderr}`,
        ErrorCodes.SEARCH_FAILED,
        { filters }
      );
    }

    const notes = this.parseTextSearchOutput(result.stdout);
    
    return {
      results: notes,
      query: { 
        type: "boolean", 
        executed: args.slice(3).join(" ") 
      },
      pagination: this.pagination.paginate({
        items: notes,
        total: notes.length,
        limit: options.limit ?? 50,
        offset: options.offset ?? 0,
      }).pagination,
    };
  }

  private async executeSqlInternal(
    sql: string,
    type: "text" | "sql",
    options: SearchOptions
  ): Promise<SearchResult<Record<string, unknown>>> {
    const result = await this.cli.exec(
      "opennotes",
      ["notes", "search", "--sql", sql, ...this.cli.buildNotebookArgs(options.notebook)],
      { signal: options.signal }
    );

    if (result.code !== 0) {
      throw new OpenNotesError(
        `SQL query failed: ${result.stderr}`,
        ErrorCodes.INVALID_SQL,
        { sql }
      );
    }

    const rows = this.cli.parseJsonOutput<Record<string, unknown>[]>(result.stdout);
    
    const { items, pagination } = this.pagination.paginate({
      items: rows,
      total: rows.length,  // Note: actual total requires separate count query
      limit: options.limit ?? 50,
      offset: options.offset ?? 0,
    });

    return {
      results: items,
      query: { type, executed: sql },
      pagination,
    };
  }

  private parseTextSearchOutput(stdout: string): NoteSummary[] {
    // Parse glamour-formatted output to extract note paths
    // This is a fallback for non-JSON output modes
    const lines = stdout.split("\n");
    const notes: NoteSummary[] = [];
    
    for (const line of lines) {
      // Look for file path patterns
      const match = line.match(/[•]\s*(\S+\.md)/);
      if (match) {
        notes.push({ path: match[1] });
      }
    }
    
    return notes;
  }
}
```

### ViewsService Implementation

```typescript
// src/services/views.service.ts

import type { IViewsService, ICliAdapter, IPaginationService, ViewExecuteOptions, ViewsListResult, ViewExecuteResult } from "./types";
import type { ViewDefinition } from "../schemas/views";
import { OpenNotesError, ErrorCodes } from "../utils/errors";

export class ViewsService implements IViewsService {
  constructor(
    private cli: ICliAdapter,
    private pagination: IPaginationService
  ) {}

  async listViews(notebook?: string): Promise<ViewsListResult> {
    const result = await this.cli.exec(
      "opennotes",
      ["notes", "view", "--list", "--format", "json", ...this.cli.buildNotebookArgs(notebook)]
    );

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Failed to list views: ${result.stderr}`,
        ErrorCodes.VIEW_LIST_FAILED
      );
    }

    const data = this.cli.parseJsonOutput<{ views: ViewDefinition[] }>(result.stdout);
    
    return {
      views: data.views,
      notebook: { name: "default", path: notebook ?? "", source: "explicit" },
    };
  }

  async executeView(name: string, options: ViewExecuteOptions): Promise<ViewExecuteResult> {
    const args = ["notes", "view", name, "--format", "json"];
    
    // Add parameters
    if (options.params && Object.keys(options.params).length > 0) {
      const paramStr = Object.entries(options.params)
        .map(([k, v]) => `${k}=${v}`)
        .join(",");
      args.push("--param", paramStr);
    }
    
    args.push(...this.cli.buildNotebookArgs(options.notebook));

    const result = await this.cli.exec("opennotes", args, { signal: options.signal });

    if (result.code !== 0) {
      if (result.stderr.includes("view not found")) {
        throw new OpenNotesError(
          `View not found: ${name}`,
          ErrorCodes.VIEW_NOT_FOUND,
          undefined,
          `Available views can be listed with opennotes_views (no parameters)`
        );
      }
      throw new OpenNotesError(
        `View execution failed: ${result.stderr}`,
        ErrorCodes.VIEW_EXECUTE_FAILED,
        { view: name }
      );
    }

    const rows = this.cli.parseJsonOutput<Record<string, unknown>[]>(result.stdout);
    
    const { items, pagination } = this.pagination.paginate({
      items: rows,
      total: rows.length,
      limit: options.limit ?? 50,
      offset: options.offset ?? 0,
    });

    // Get view definition for description
    const viewDef = await this.getView(name, options.notebook);

    return {
      view: {
        name,
        description: viewDef?.description ?? "",
      },
      results: items,
      pagination,
      notebook: { name: "default", path: options.notebook ?? "", source: "explicit" },
    };
  }

  async getView(name: string, notebook?: string): Promise<ViewDefinition | null> {
    const { views } = await this.listViews(notebook);
    return views.find(v => v.name === name) ?? null;
  }
}
```

---

## Service Composition

### ServiceContainer

```typescript
// src/services/index.ts

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { CliAdapter } from "./cli-adapter";
import { SearchService } from "./search.service";
import { ListService } from "./list.service";
import { NoteService } from "./note.service";
import { NotebookService } from "./notebook.service";
import { ViewsService } from "./views.service";
import { PaginationService } from "./pagination.service";
import type { ExtensionConfig } from "../config";

export interface Services {
  cli: CliAdapter;
  search: SearchService;
  list: ListService;
  note: NoteService;
  notebook: NotebookService;
  views: ViewsService;
  pagination: PaginationService;
}

export function createServices(pi: ExtensionAPI, config: ExtensionConfig): Services {
  // Create CLI adapter first (foundation)
  const cli = new CliAdapter(pi, {
    cliPath: config.cliPath,
    defaultTimeout: config.cliTimeout,
  });

  // Create pagination service (shared utility)
  const pagination = new PaginationService({
    defaultPageSize: config.defaultPageSize,
    maxOutputBytes: 50 * 1024,  // 50KB
    maxOutputLines: 2000,
    budgetRatio: 0.75,
  });

  // Create domain services (depend on cli + pagination)
  const search = new SearchService(cli, pagination);
  const list = new ListService(cli, pagination);
  const note = new NoteService(cli);
  const notebook = new NotebookService(cli);
  const views = new ViewsService(cli, pagination);

  return {
    cli,
    search,
    list,
    note,
    notebook,
    views,
    pagination,
  };
}

// Re-export for consumers
export * from "./types";
export { CliAdapter } from "./cli-adapter";
export { SearchService } from "./search.service";
export { ListService } from "./list.service";
export { NoteService } from "./note.service";
export { NotebookService } from "./notebook.service";
export { ViewsService } from "./views.service";
export { PaginationService } from "./pagination.service";
```

---

## Service Dependency Graph

```
┌─────────────────────────────────────────────────────────────┐
│                   Service Dependencies                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│    ┌─────────────┐                                          │
│    │ ExtensionAPI│ (from pi-coding-agent)                   │
│    └──────┬──────┘                                          │
│           │                                                 │
│           ▼                                                 │
│    ┌─────────────┐     ┌─────────────────┐                  │
│    │ CliAdapter  │     │ ExtensionConfig │                  │
│    └──────┬──────┘     └────────┬────────┘                  │
│           │                     │                           │
│           ├─────────────────────┤                           │
│           │                     │                           │
│           │        ┌────────────┴────────────┐              │
│           │        │                         │              │
│           ▼        ▼                         ▼              │
│    ┌────────────────────┐    ┌────────────────────┐         │
│    │ PaginationService  │    │ Other Config Props │         │
│    │ (budgetRatio, etc) │    │ (toolPrefix, etc)  │         │
│    └─────────┬──────────┘    └────────────────────┘         │
│              │                                              │
│    ┌─────────┼─────────────────────────────────────┐        │
│    │         │         │         │         │       │        │
│    ▼         ▼         ▼         ▼         ▼       ▼        │
│ ┌───────┐ ┌───────┐ ┌───────┐ ┌────────┐ ┌───────┐          │
│ │Search │ │ List  │ │ Note  │ │Notebook│ │ Views │          │
│ │Service│ │Service│ │Service│ │Service │ │Service│          │
│ └───────┘ └───────┘ └───────┘ └────────┘ └───────┘          │
│                                                             │
└─────────────────────────────────────────────────────────────┘

Legend:
  ──▶  Depends on
```

---

## Tool-Service Mapping

| Tool | Primary Service | Secondary Services |
|------|-----------------|-------------------|
| `opennotes_search` | SearchService | PaginationService |
| `opennotes_list` | ListService | PaginationService |
| `opennotes_get` | NoteService | NotebookService |
| `opennotes_create` | NoteService | NotebookService |
| `opennotes_notebooks` | NotebookService | - |
| `opennotes_views` | ViewsService | PaginationService |

---

## Expected Outcome

✅ Service architecture designed with SOLID principles.

## Actual Outcome

Complete service architecture with:
- All service interfaces defined
- Implementation patterns established
- Dependency injection via `createServices()`
- Clear service-tool mapping
- Composition pattern documented

## Lessons Learned

1. CliAdapter is the central abstraction - all services depend on it
2. PaginationService is shared across search, list, and views
3. Services should be stateless (config injected at construction)
4. Dependency graph is shallow (max 2 levels) for simplicity
