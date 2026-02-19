/**
 * Create tool for pi-jot
 * Wraps NoteService to provide LLM-callable tool for creating notes
 */

import type { Tool } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ToolConfig } from "../config";
import { CreateParams, type CreateParamsType } from "../schemas/tools";
import { JotError, ErrorCodes, wrapError } from "../utils/errors";
import { formatCreateResult } from "../utils/output";

// =============================================================================
// Tool Description
// =============================================================================

const DESCRIPTION = `Create a new note in an Jot notebook.

Required: 'title' - becomes the note's title in frontmatter

Optional:
- 'path': Directory within notebook (e.g., 'tasks/' creates in tasks folder)
- 'template': Use a predefined template (e.g., 'meeting', 'task')
- 'content': Initial markdown body
- 'data': Additional frontmatter fields (e.g., {"tag": "meeting", "priority": "high"})

Returns the created note path.`;

// =============================================================================
// Tool Factory
// =============================================================================

export function createCreateTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}create`,
    label: "Create Note",
    description: DESCRIPTION,
    parameters: CreateParams,

    async execute(toolCallId, params: CreateParamsType, onUpdate, ctx, signal) {
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

        // Create the note
        const result = await services.note.createNote(params.title, {
          notebook: params.notebook,
          path: params.path,
          template: params.template,
          content: params.content,
          data: params.data,
          signal,
        });

        // Format output for LLM
        const output = formatCreateResult(result.created, result.notebook);

        return {
          content: [{ type: "text", text: output }],
        };
      } catch (error) {
        const wrapped = wrapError(error, ErrorCodes.NOTE_CREATE_FAILED);
        return wrapped.toToolResult();
      }
    },
  };
}
