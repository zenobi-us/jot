/**
 * Service interfaces for pi-opennotes
 * All business logic is defined through these interfaces following SOLID principles
 */

import type { Static } from "@sinclair/typebox";

// =============================================================================
// CLI Adapter Types
// =============================================================================

export interface CliResult {
  code: number;
  stdout: string;
  stderr: string;
  timedOut?: boolean;
}

export interface CliOptions {
  notebook?: string;
  timeout?: number;
  signal?: AbortSignal | null;
  env?: Record<string, string>;
}

export interface InstallationInfo {
  installed: boolean;
  version?: string;
  path?: string;
}

export interface ICliAdapter {
  /**
   * Execute CLI command
   */
  exec(command: string, args: string[], options?: CliOptions): Promise<CliResult>;

  /**
   * Check if CLI is installed and accessible
   */
  checkInstallation(): Promise<InstallationInfo>;

  /**
   * Parse JSON output with error handling
   */
  parseJsonOutput<T>(stdout: string): T;

  /**
   * Build notebook flag args
   */
  buildNotebookArgs(notebook?: string): string[];
}

// =============================================================================
// Pagination Types
// =============================================================================

export interface PaginationMeta {
  total: number;
  returned: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
  nextOffset?: number;
}

export interface PaginationInput<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface PaginationResult<T> {
  items: T[];
  pagination: PaginationMeta;
}

export interface FitToBudgetResult<T> {
  items: T[];
  truncated: boolean;
  originalCount: number;
}

export interface PaginationConfig {
  defaultPageSize: number;
  maxOutputBytes: number;
  maxOutputLines: number;
  budgetRatio: number;
}

export interface IPaginationService {
  /**
   * Apply pagination to results
   */
  paginate<T>(input: PaginationInput<T>): PaginationResult<T>;

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
  ): FitToBudgetResult<T>;
}

// =============================================================================
// Note Types
// =============================================================================

export interface NoteSummary {
  path: string;
  title?: string;
  tags?: string[];
  created?: string;
  modified?: string;
}

export interface NoteContent {
  path: string;
  title?: string;
  content: string;
  frontmatter?: Record<string, unknown>;
  wordCount?: number;
}

export interface NotebookInfo {
  name: string;
  path: string;
  source: "registered" | "ancestor" | "explicit";
  noteCount?: number;
}

export interface ViewDefinition {
  name: string;
  origin: "built-in" | "notebook" | "global";
  description?: string;
  parameters?: Array<{
    name: string;
    type: string;
    required: boolean;
    default?: string;
    description?: string;
  }>;
}

// =============================================================================
// Search Service Types
// =============================================================================

export interface SearchOptions {
  notebook?: string;
  limit?: number;
  offset?: number;
  signal?: AbortSignal | null;
}

export interface BooleanFilters {
  and?: string[];
  or?: string[];
  not?: string[];
}

export type SearchQueryType = "text" | "fuzzy" | "sql" | "boolean";

export interface SearchResult<T = NoteSummary> {
  results: T[];
  query: {
    type: SearchQueryType;
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
  sqlSearch(
    sql: string,
    options: SearchOptions
  ): Promise<SearchResult<Record<string, unknown>>>;

  /**
   * Boolean query (AND/OR/NOT filters)
   */
  booleanSearch(
    filters: BooleanFilters,
    options: SearchOptions
  ): Promise<SearchResult>;
}

// =============================================================================
// List Service Types
// =============================================================================

export type SortField = "modified" | "created" | "title" | "path";
export type SortOrder = "asc" | "desc";

export interface ListOptions {
  notebook?: string;
  sortBy?: SortField;
  sortOrder?: SortOrder;
  pattern?: string;
  limit?: number;
  offset?: number;
  signal?: AbortSignal | null;
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

// =============================================================================
// Note Service Types
// =============================================================================

export interface GetOptions {
  notebook?: string;
  includeContent?: boolean;
  signal?: AbortSignal | null;
}

export interface CreateOptions {
  notebook?: string;
  path?: string;
  template?: string;
  content?: string;
  data?: Record<string, string | number | boolean | string[]>;
  signal?: AbortSignal | null;
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

// =============================================================================
// Notebook Service Types
// =============================================================================

export interface NotebooksResult {
  notebooks: NotebookInfo[];
  current: NotebookInfo | null;
}

export interface ValidateNotebookResult {
  valid: boolean;
  error?: string;
}

export interface INotebookService {
  /**
   * List all available notebooks
   */
  listNotebooks(): Promise<NotebooksResult>;

  /**
   * Get current notebook (from context or explicit)
   */
  getCurrentNotebook(explicitPath?: string): Promise<NotebookInfo | null>;

  /**
   * Validate a notebook path
   */
  validateNotebook(path: string): Promise<ValidateNotebookResult>;
}

// =============================================================================
// Views Service Types
// =============================================================================

export interface ViewExecuteOptions {
  notebook?: string;
  params?: Record<string, string>;
  limit?: number;
  offset?: number;
  signal?: AbortSignal | null;
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
