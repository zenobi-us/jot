/**
 * List service for pi-opennotes
 * Handles listing notes with sorting and filtering
 */

import type {
  IListService,
  ICliAdapter,
  IPaginationService,
  ListOptions,
  ListResult,
  NoteSummary,
  NotebookInfo,
} from "./types";
import { OpenNotesError, ErrorCodes } from "../utils/errors";
import { validatePagination } from "../utils/validation";

// =============================================================================
// List Service Implementation
// =============================================================================

export class ListService implements IListService {
  constructor(
    private readonly cli: ICliAdapter,
    private readonly pagination: IPaginationService
  ) {}

  /**
   * List notes with optional filtering and sorting
   */
  async listNotes(options: ListOptions): Promise<ListResult> {
    validatePagination(options.limit, options.offset);

    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;
    const pattern = options.pattern ?? "**/*.md";
    const sortBy = options.sortBy ?? "modified";
    const sortOrder = options.sortOrder ?? "desc";

    // Map sort fields to SQL columns
    const sortFieldMap: Record<string, string> = {
      modified: "metadata->>'modified'",
      created: "metadata->>'created'",
      title: "metadata->>'title'",
      path: "file_path",
    };

    const sortColumn = sortFieldMap[sortBy] ?? "file_path";
    const orderDirection = sortOrder.toUpperCase();

    // Build SQL query
    const sql = `
      SELECT 
        file_path as path,
        metadata->>'title' as title,
        metadata->'tags' as tags,
        metadata->>'created' as created,
        metadata->>'modified' as modified
      FROM read_markdown('${pattern}')
      ORDER BY ${sortColumn} ${orderDirection} NULLS LAST
      LIMIT ${limit}
      OFFSET ${offset}
    `.trim();

    const args = [
      "notes",
      "search",
      "--sql",
      sql,
      ...this.cli.buildNotebookArgs(options.notebook),
    ];

    const result = await this.cli.exec("opennotes", args, {
      signal: options.signal,
    });

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Failed to list notes: ${result.stderr}`,
        ErrorCodes.SEARCH_FAILED,
        { pattern, stderr: result.stderr }
      );
    }

    // Parse results
    const notes = this.cli.parseJsonOutput<NoteSummary[]>(result.stdout);

    // Get notebook info
    const notebook = await this.getNotebookInfo(options.notebook);

    // Get total count for pagination
    const total = await this.countNotes({ notebook: options.notebook, pattern });

    const { pagination } = this.pagination.paginate({
      items: notes,
      total,
      limit,
      offset,
    });

    return {
      notes,
      notebook,
      pagination,
    };
  }

  /**
   * Count total notes (for pagination)
   */
  async countNotes(
    options: Pick<ListOptions, "notebook" | "pattern">
  ): Promise<number> {
    const pattern = options.pattern ?? "**/*.md";

    const sql = `
      SELECT COUNT(*) as count
      FROM read_markdown('${pattern}')
    `.trim();

    const args = [
      "notes",
      "search",
      "--sql",
      sql,
      ...this.cli.buildNotebookArgs(options.notebook),
    ];

    const result = await this.cli.exec("opennotes", args);

    if (result.code !== 0) {
      // If count fails, return 0 as approximation
      return 0;
    }

    try {
      const data = this.cli.parseJsonOutput<Array<{ count: number }>>(result.stdout);
      return data[0]?.count ?? 0;
    } catch {
      return 0;
    }
  }

  /**
   * Get notebook info for a path
   */
  private async getNotebookInfo(notebookPath?: string): Promise<NotebookInfo> {
    // Try to get notebook info via CLI
    const args = ["notebook", "info", "--format", "json"];
    if (notebookPath) {
      args.push("--notebook", notebookPath);
    }

    const result = await this.cli.exec("opennotes", args);

    if (result.code === 0) {
      try {
        const data = this.cli.parseJsonOutput<{
          name?: string;
          path?: string;
        }>(result.stdout);
        return {
          name: data.name ?? "Unknown",
          path: data.path ?? notebookPath ?? "",
          source: notebookPath ? "explicit" : "ancestor",
        };
      } catch {
        // Fall through to default
      }
    }

    // Return default notebook info
    return {
      name: notebookPath ? notebookPath.split("/").pop() ?? "Notebook" : "Current",
      path: notebookPath ?? "",
      source: notebookPath ? "explicit" : "ancestor",
    };
  }
}
