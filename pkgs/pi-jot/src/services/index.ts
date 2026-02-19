/**
 * Service exports and factory for pi-jot
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { CliAdapter, type CliAdapterConfig } from "./cli-adapter";
import { PaginationService } from "./pagination.service";
import { SearchService } from "./search.service";
import { ListService } from "./list.service";
import { NoteService } from "./note.service";
import { NotebookService } from "./notebook.service";
import { ViewsService } from "./views.service";
import type { ExtensionConfig } from "../config";

// =============================================================================
// Services Container
// =============================================================================

export interface Services {
  cli: CliAdapter;
  pagination: PaginationService;
  search: SearchService;
  list: ListService;
  note: NoteService;
  notebook: NotebookService;
  views: ViewsService;
}

// =============================================================================
// Service Factory
// =============================================================================

/**
 * Create all services with proper dependency injection
 */
export function createServices(pi: ExtensionAPI, config: ExtensionConfig): Services {
  // Create CLI adapter first (foundation)
  const cli = new CliAdapter(pi, {
    cliPath: config.cliPath,
    defaultTimeout: config.cliTimeout,
  });

  // Create pagination service (shared utility)
  const pagination = new PaginationService({
    defaultPageSize: config.defaultPageSize,
    maxOutputBytes: config.maxOutputBytes,
    maxOutputLines: config.maxOutputLines,
    budgetRatio: config.budgetRatio,
  });

  // Create domain services (depend on cli + pagination)
  const search = new SearchService(cli, pagination);
  const list = new ListService(cli, pagination);
  const note = new NoteService(cli);
  const notebook = new NotebookService(cli);
  const views = new ViewsService(cli, pagination);

  return {
    cli,
    pagination,
    search,
    list,
    note,
    notebook,
    views,
  };
}

// =============================================================================
// Re-exports
// =============================================================================

export * from "./types";
export { CliAdapter, type CliAdapterConfig } from "./cli-adapter";
export { PaginationService } from "./pagination.service";
export { SearchService } from "./search.service";
export { ListService } from "./list.service";
export { NoteService } from "./note.service";
export { NotebookService } from "./notebook.service";
export { ViewsService } from "./views.service";
