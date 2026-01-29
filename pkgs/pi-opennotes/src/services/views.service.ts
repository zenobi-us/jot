/**
 * Views service for pi-opennotes
 * Handles listing and executing views
 */

import type {
  IViewsService,
  ICliAdapter,
  IPaginationService,
  ViewExecuteOptions,
  ViewsListResult,
  ViewExecuteResult,
  ViewDefinition,
  NotebookInfo,
} from "./types";
import { OpenNotesError, ErrorCodes } from "../utils/errors";
import { validateViewName, validatePagination } from "../utils/validation";

// =============================================================================
// Views Service Implementation
// =============================================================================

export class ViewsService implements IViewsService {
  constructor(
    private readonly cli: ICliAdapter,
    private readonly pagination: IPaginationService
  ) {}

  /**
   * List all available views
   */
  async listViews(notebook?: string): Promise<ViewsListResult> {
    const args = [
      "notes",
      "view",
      "--list",
      "--format",
      "json",
      ...this.cli.buildNotebookArgs(notebook),
    ];

    const result = await this.cli.exec("opennotes", args);

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Failed to list views: ${result.stderr}`,
        ErrorCodes.VIEW_LIST_FAILED,
        { stderr: result.stderr }
      );
    }

    // Parse the output
    let views: ViewDefinition[] = [];
    try {
      const data = this.cli.parseJsonOutput<{
        views?: Array<{
          name: string;
          origin?: string;
          description?: string;
          parameters?: Array<{
            name: string;
            type: string;
            required?: boolean;
            default?: string;
            description?: string;
          }>;
        }>;
      }>(result.stdout);

      views = (data.views ?? []).map((v) => ({
        name: v.name,
        origin: (v.origin as ViewDefinition["origin"]) ?? "built-in",
        description: v.description,
        parameters: v.parameters?.map((p) => ({
          name: p.name,
          type: p.type,
          required: p.required ?? false,
          default: p.default,
          description: p.description,
        })),
      }));
    } catch {
      // If parsing fails, use built-in views as fallback
      views = this.getBuiltInViews();
    }

    const notebookInfo = await this.getNotebookInfo(notebook);

    return {
      views,
      notebook: notebookInfo,
    };
  }

  /**
   * Execute a named view
   */
  async executeView(
    name: string,
    options: ViewExecuteOptions
  ): Promise<ViewExecuteResult> {
    validateViewName(name);
    validatePagination(options.limit, options.offset);

    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;

    const args = ["notes", "view", name, "--format", "json"];

    // Add parameters if specified
    if (options.params && Object.keys(options.params).length > 0) {
      const paramStr = Object.entries(options.params)
        .map(([k, v]) => `${k}=${v}`)
        .join(",");
      args.push("--param", paramStr);
    }

    args.push(...this.cli.buildNotebookArgs(options.notebook));

    const result = await this.cli.exec("opennotes", args, {
      signal: options.signal,
    });

    if (result.code !== 0) {
      // Check for specific error types
      if (result.stderr.includes("view not found") || result.stderr.includes("no such view")) {
        throw new OpenNotesError(
          `View not found: ${name}`,
          ErrorCodes.VIEW_NOT_FOUND,
          { view: name },
          "Use opennotes_views (without arguments) to list available views."
        );
      }
      if (result.stderr.includes("invalid param") || result.stderr.includes("missing required")) {
        throw new OpenNotesError(
          `Invalid parameters for view: ${name}`,
          ErrorCodes.VIEW_INVALID_PARAMS,
          { view: name, params: options.params, stderr: result.stderr }
        );
      }
      throw new OpenNotesError(
        `View execution failed: ${result.stderr}`,
        ErrorCodes.VIEW_EXECUTE_FAILED,
        { view: name, stderr: result.stderr }
      );
    }

    // Parse results
    const rows = this.cli.parseJsonOutput<Record<string, unknown>[]>(result.stdout);

    // Apply pagination
    const paginatedRows = rows.slice(offset, offset + limit);

    const { pagination } = this.pagination.paginate({
      items: paginatedRows,
      total: rows.length,
      limit,
      offset,
    });

    // Get view definition for description
    const viewDef = await this.getView(name, options.notebook);
    const notebookInfo = await this.getNotebookInfo(options.notebook);

    return {
      view: {
        name,
        description: viewDef?.description ?? "",
      },
      results: paginatedRows,
      pagination,
      notebook: notebookInfo,
    };
  }

  /**
   * Get a specific view definition
   */
  async getView(name: string, notebook?: string): Promise<ViewDefinition | null> {
    try {
      const { views } = await this.listViews(notebook);
      return views.find((v) => v.name === name) ?? null;
    } catch {
      // If listing fails, check if it's a built-in view
      const builtIn = this.getBuiltInViews();
      return builtIn.find((v) => v.name === name) ?? null;
    }
  }

  /**
   * Get built-in view definitions
   */
  private getBuiltInViews(): ViewDefinition[] {
    return [
      {
        name: "today",
        origin: "built-in",
        description: "Notes modified today",
      },
      {
        name: "recent",
        origin: "built-in",
        description: "Recently modified notes",
        parameters: [
          {
            name: "days",
            type: "number",
            required: false,
            default: "7",
            description: "Number of days to look back",
          },
        ],
      },
      {
        name: "kanban",
        origin: "built-in",
        description: "Kanban board view grouped by status",
        parameters: [
          {
            name: "status",
            type: "string",
            required: false,
            default: "todo,in-progress,done",
            description: "Comma-separated list of statuses",
          },
        ],
      },
      {
        name: "untagged",
        origin: "built-in",
        description: "Notes without any tags",
      },
      {
        name: "orphans",
        origin: "built-in",
        description: "Notes not linked to/from other notes",
      },
      {
        name: "broken-links",
        origin: "built-in",
        description: "Notes with broken internal links",
      },
    ];
  }

  /**
   * Get notebook info for a path
   */
  private async getNotebookInfo(notebookPath?: string): Promise<NotebookInfo> {
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

    return {
      name: notebookPath?.split("/").pop() ?? "Current",
      path: notebookPath ?? "",
      source: notebookPath ? "explicit" : "ancestor",
    };
  }
}
