/**
 * Get tool for pi-opennotes
 * Wraps NoteService to provide LLM-callable tool for getting individual notes
 */

import type { Tool } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ToolConfig } from "../config";
import { GetParams, type GetParamsType } from "../schemas/tools";
import { OpenNotesError, ErrorCodes, wrapError } from "../utils/errors";
import { formatNoteContent, formatNotebookInfo } from "../utils/output";

// =============================================================================
// Tool Description
// =============================================================================

const DESCRIPTION = `Get the full content and metadata of a specific note by path.

The path should be relative to the notebook root (e.g., 'tasks/task-001.md').

Set 'includeContent: false' to get only metadata without the full body (faster for large notes).`;

// =============================================================================
// Tool Factory
// =============================================================================

export function createGetTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}get`,
    label: "Get Note",
    description: DESCRIPTION,
    parameters: GetParams,

    async execute(toolCallId, params: GetParamsType, onUpdate, ctx, signal) {
      try {
        // Check CLI installation
        const installation = await services.cli.checkInstallation();
        if (!installation.installed) {
          throw new OpenNotesError(
            "OpenNotes CLI not found",
            ErrorCodes.CLI_NOT_FOUND,
            { searchedPaths: process.env.PATH?.split(":") }
          );
        }

        // Get the note
        const result = await services.note.getNote(params.path, {
          notebook: params.notebook,
          includeContent: params.includeContent,
          signal,
        });

        // Format output for LLM
        const lines: string[] = [];
        lines.push(formatNoteContent(result.note));
        lines.push("");
        lines.push("---");
        lines.push(`*Notebook: ${formatNotebookInfo(result.notebook)}*`);

        return {
          content: [{ type: "text", text: lines.join("\n") }],
        };
      } catch (error) {
        const wrapped = wrapError(error, ErrorCodes.NOTE_NOT_FOUND);
        return wrapped.toToolResult();
      }
    },
  };
}
