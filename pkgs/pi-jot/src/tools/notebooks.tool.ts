/**
 * Notebooks tool for pi-jot
 * Wraps NotebookService to provide LLM-callable tool
 */

import type { Tool } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ToolConfig } from "../config";
import { NotebooksParams, type NotebooksParamsType } from "../schemas/tools";
import { JotError, ErrorCodes, wrapError } from "../utils/errors";
import { formatNotebooksList } from "../utils/output";

// =============================================================================
// Tool Description
// =============================================================================

const DESCRIPTION = `List all available Jot notebooks.

Returns notebooks from:
- Global config (registered notebooks)
- Ancestor directories (discovered notebooks)

Each notebook includes name, path, and source.`;

// =============================================================================
// Tool Factory
// =============================================================================

export function createNotebooksTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}notebooks`,
    label: "List Notebooks",
    description: DESCRIPTION,
    parameters: NotebooksParams,

    async execute(toolCallId, params: NotebooksParamsType, onUpdate, ctx, signal) {
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

        // List notebooks
        const result = await services.notebook.listNotebooks();

        // Format output for LLM
        const output = formatNotebooksList(result.notebooks, result.current);

        return {
          content: [{ type: "text", text: output }],
        };
      } catch (error) {
        const wrapped = wrapError(error, ErrorCodes.NOTEBOOK_NOT_FOUND);
        return wrapped.toToolResult();
      }
    },
  };
}
