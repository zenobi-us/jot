/**
 * List tool for pi-jot
 * Wraps ListService to provide LLM-callable tool
 */

import type { Tool } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ToolConfig } from "../config";
import { ListParams, type ListParamsType } from "../schemas/tools";
import { JotError, ErrorCodes, wrapError } from "../utils/errors";
import { formatListResults } from "../utils/output";

// =============================================================================
// Tool Description
// =============================================================================

const DESCRIPTION = `List all notes in an Jot notebook with metadata.

Use 'sortBy' to order by: modified (default), created, title, or path.
Use 'pattern' to filter: e.g., 'tasks/*.md' for only task notes.

Returns note summaries with pagination. For full content, use jot_get.`;

// =============================================================================
// Tool Factory
// =============================================================================

export function createListTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}list`,
    label: "List Notes",
    description: DESCRIPTION,
    parameters: ListParams,

    async execute(toolCallId, params: ListParamsType, onUpdate, ctx, signal) {
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

        // Execute list
        const result = await services.list.listNotes({
          notebook: params.notebook,
          sortBy: params.sortBy,
          sortOrder: params.sortOrder,
          pattern: params.pattern,
          limit: params.limit,
          offset: params.offset,
          signal,
        });

        // Format output for LLM
        const output = formatListResults(
          result.notes,
          result.notebook,
          result.pagination
        );

        return {
          content: [{ type: "text", text: output }],
        };
      } catch (error) {
        const wrapped = wrapError(error, ErrorCodes.SEARCH_FAILED);
        return wrapped.toToolResult();
      }
    },
  };
}
