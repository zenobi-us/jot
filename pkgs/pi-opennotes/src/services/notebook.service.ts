/**
 * Notebook service for pi-opennotes
 * Handles listing and validating notebooks
 */

import type {
  INotebookService,
  ICliAdapter,
  NotebooksResult,
  NotebookInfo,
  ValidateNotebookResult,
} from "./types";
import { OpenNotesError, ErrorCodes } from "../utils/errors";

// =============================================================================
// Notebook Service Implementation
// =============================================================================

export class NotebookService implements INotebookService {
  constructor(private readonly cli: ICliAdapter) {}

  /**
   * List all available notebooks
   */
  async listNotebooks(): Promise<NotebooksResult> {
    // Try to get notebooks via CLI
    const result = await this.cli.exec("opennotes", ["notebook", "list", "--format", "json"]);

    if (result.code !== 0) {
      // If list fails, return empty list with no current
      return {
        notebooks: [],
        current: null,
      };
    }

    // Parse the output
    let notebooks: NotebookInfo[] = [];
    try {
      const data = this.cli.parseJsonOutput<
        Array<{
          name?: string;
          path?: string;
          source?: string;
          noteCount?: number;
        }>
      >(result.stdout);

      notebooks = data.map((nb) => ({
        name: nb.name ?? nb.path?.split("/").pop() ?? "Unknown",
        path: nb.path ?? "",
        source: (nb.source as NotebookInfo["source"]) ?? "registered",
        noteCount: nb.noteCount,
      }));
    } catch {
      // If parsing fails, return empty list
    }

    // Get current notebook
    const current = await this.getCurrentNotebook();

    return {
      notebooks,
      current,
    };
  }

  /**
   * Get current notebook (from context or explicit path)
   */
  async getCurrentNotebook(explicitPath?: string): Promise<NotebookInfo | null> {
    const args = ["notebook", "info", "--format", "json"];
    if (explicitPath) {
      args.push("--notebook", explicitPath);
    }

    const result = await this.cli.exec("opennotes", args);

    if (result.code !== 0) {
      return null;
    }

    try {
      const data = this.cli.parseJsonOutput<{
        name?: string;
        path?: string;
        noteCount?: number;
      }>(result.stdout);

      return {
        name: data.name ?? data.path?.split("/").pop() ?? "Current",
        path: data.path ?? explicitPath ?? "",
        source: explicitPath ? "explicit" : "ancestor",
        noteCount: data.noteCount,
      };
    } catch {
      return null;
    }
  }

  /**
   * Validate a notebook path
   */
  async validateNotebook(path: string): Promise<ValidateNotebookResult> {
    if (!path) {
      return {
        valid: false,
        error: "Path is required",
      };
    }

    // Check if path contains config file
    const result = await this.cli.exec("opennotes", [
      "notebook",
      "info",
      "--format",
      "json",
      "--notebook",
      path,
    ]);

    if (result.code === 0) {
      return { valid: true };
    }

    // Check specific error conditions
    if (result.stderr.includes("not found") || result.stderr.includes("does not exist")) {
      return {
        valid: false,
        error: "Notebook path does not exist",
      };
    }

    if (result.stderr.includes(".opennotes.json") || result.stderr.includes("config")) {
      return {
        valid: false,
        error: "Directory is not a valid notebook (missing .opennotes.json)",
      };
    }

    return {
      valid: false,
      error: result.stderr || "Unknown validation error",
    };
  }

  /**
   * Register a notebook (add to global config)
   */
  async registerNotebook(path: string, name?: string): Promise<void> {
    const args = ["notebook", "register", path];
    if (name) {
      args.push("--name", name);
    }

    const result = await this.cli.exec("opennotes", args);

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Failed to register notebook: ${result.stderr}`,
        ErrorCodes.NOTEBOOK_CONFIG_ERROR,
        { path, stderr: result.stderr }
      );
    }
  }
}
