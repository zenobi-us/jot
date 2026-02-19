/**
 * Views tool for pi-jot
 * Wraps ViewsService to provide LLM-callable tool
 * Dual-mode: list views (no view param) or execute view (with view param)
 */

import type { Tool } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ToolConfig } from "../config";
import { ViewsParams, type ViewsParamsType } from "../schemas/tools";
import { JotError, ErrorCodes, wrapError } from "../utils/errors";
import { formatViewsList, formatViewResults } from "../utils/output";

// =============================================================================
// Tool Description
// =============================================================================

const DESCRIPTION = `List available views or execute a named view.

**List Mode** (no 'view' parameter):
Returns all available views with descriptions and parameters.

**Execute Mode** (with 'view' parameter):
Executes the named view and returns results.

Built-in views: today, recent, kanban, untagged, orphans, broken-links

Example: { view: 'kanban', params: { status: 'todo,in-progress,done' } }`;

// =============================================================================
// Tool Factory
// =============================================================================

export function createViewsTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}views`,
    label: "Views",
    description: DESCRIPTION,
    parameters: ViewsParams,

    async execute(toolCallId, params: ViewsParamsType, onUpdate, ctx, signal) {
      try {
        // Check CLI installation
        const installation = await services.cli.checkInstallation();
        if (!installation.installed) {
          throw new JotError(
            "Jot CLI not found",
            ErrorCodes.CLI_NOT_FOUND,
            { searchedPaths: process.env.PATH?.split(":") }
          );
        }

        // Determine mode: list vs execute
        if (params.view) {
          // Execute mode
          const result = await services.views.executeView(params.view, {
            notebook: params.notebook,
            params: params.params,
            limit: params.limit,
            offset: params.offset,
            signal,
          });

          // Format output for LLM
          const output = formatViewResults(
            result.view.name,
            result.results,
            result.pagination
          );

          return {
            content: [{ type: "text", text: output }],
          };
        } else {
          // List mode
          const result = await services.views.listViews(params.notebook);

          // Format output for LLM
          const output = formatViewsList(result.views, result.notebook);

          return {
            content: [{ type: "text", text: output }],
          };
        }
      } catch (error) {
        const wrapped = wrapError(error, ErrorCodes.VIEW_NOT_FOUND);
        return wrapped.toToolResult();
      }
    },
  };
}
