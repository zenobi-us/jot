/**
 * Note service for pi-opennotes
 * Handles getting and creating individual notes
 */

import type {
  INoteService,
  ICliAdapter,
  GetOptions,
  CreateOptions,
  GetResult,
  CreateResult,
  NoteContent,
  NotebookInfo,
} from "./types";
import { OpenNotesError, ErrorCodes } from "../utils/errors";
import { validateNotePath, validateNoteTitle, escapeSqlString } from "../utils/validation";

// =============================================================================
// Note Service Implementation
// =============================================================================

export class NoteService implements INoteService {
  constructor(private readonly cli: ICliAdapter) {}

  /**
   * Get a specific note by path
   */
  async getNote(path: string, options: GetOptions): Promise<GetResult> {
    validateNotePath(path);

    const includeContent = options.includeContent ?? true;

    // Use SQL query to get note data
    const sql = includeContent
      ? `
        SELECT 
          file_path as path,
          metadata->>'title' as title,
          content,
          metadata as frontmatter,
          length(content) - length(replace(content, ' ', '')) + 1 as wordCount
        FROM read_markdown('${escapeSqlString(path)}')
        LIMIT 1
      `.trim()
      : `
        SELECT 
          file_path as path,
          metadata->>'title' as title,
          metadata as frontmatter
        FROM read_markdown('${escapeSqlString(path)}')
        LIMIT 1
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
        `Failed to get note: ${result.stderr}`,
        ErrorCodes.NOTE_NOT_FOUND,
        { path, stderr: result.stderr }
      );
    }

    const notes = this.cli.parseJsonOutput<
      Array<{
        path: string;
        title?: string;
        content?: string;
        frontmatter?: Record<string, unknown>;
        wordCount?: number;
      }>
    >(result.stdout);

    if (notes.length === 0) {
      throw new OpenNotesError(
        `Note not found: ${path}`,
        ErrorCodes.NOTE_NOT_FOUND,
        { path }
      );
    }

    const noteData = notes[0];
    const note: NoteContent = {
      path: noteData.path,
      title: noteData.title,
      content: noteData.content ?? "",
      frontmatter: noteData.frontmatter,
      wordCount: noteData.wordCount,
    };

    const notebook = await this.getNotebookInfo(options.notebook);

    return { note, notebook };
  }

  /**
   * Create a new note
   */
  async createNote(title: string, options: CreateOptions): Promise<CreateResult> {
    validateNoteTitle(title);

    // Build args for note add command
    const args = ["notes", "add", "--title", title];

    // Add optional path
    if (options.path) {
      args.push("--path", options.path);
    }

    // Add template if specified
    if (options.template) {
      args.push("--template", options.template);
    }

    // Add notebook if specified
    args.push(...this.cli.buildNotebookArgs(options.notebook));

    // Execute note creation
    const result = await this.cli.exec("opennotes", args, {
      signal: options.signal,
    });

    if (result.code !== 0) {
      // Check for specific errors
      if (result.stderr.includes("template") && result.stderr.includes("not found")) {
        throw new OpenNotesError(
          `Template not found: ${options.template}`,
          ErrorCodes.TEMPLATE_NOT_FOUND,
          { template: options.template }
        );
      }
      if (result.stderr.includes("already exists")) {
        throw new OpenNotesError(
          `Note already exists: ${title}`,
          ErrorCodes.NOTE_CREATE_FAILED,
          { title, stderr: result.stderr }
        );
      }
      throw new OpenNotesError(
        `Failed to create note: ${result.stderr}`,
        ErrorCodes.NOTE_CREATE_FAILED,
        { title, stderr: result.stderr }
      );
    }

    // Parse the output to get the created note path
    // Output format is typically: "Created note: <path>"
    const pathMatch = result.stdout.match(/[Cc]reated.*?:\s*(\S+\.md)/);
    const createdPath = pathMatch?.[1] ?? `${title.toLowerCase().replace(/\s+/g, "-")}.md`;

    // Add content if specified (by editing the file)
    if (options.content) {
      await this.appendContent(createdPath, options.content, options.notebook);
    }

    // Add data fields if specified
    if (options.data && Object.keys(options.data).length > 0) {
      // For now, we'd need to edit the file to add frontmatter
      // This would require reading the file, modifying, and writing back
      // For MVP, skip this - the CLI should handle it with template variables
    }

    const notebook = await this.getNotebookInfo(options.notebook);

    return {
      created: {
        path: createdPath,
        absolutePath: `${notebook.path}/${createdPath}`,
        title,
      },
      notebook,
    };
  }

  /**
   * Check if a note exists
   */
  async noteExists(
    path: string,
    options: Pick<GetOptions, "notebook">
  ): Promise<boolean> {
    try {
      validateNotePath(path);

      const sql = `
        SELECT 1
        FROM read_markdown('${escapeSqlString(path)}')
        LIMIT 1
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
        return false;
      }

      const data = this.cli.parseJsonOutput<unknown[]>(result.stdout);
      return data.length > 0;
    } catch {
      return false;
    }
  }

  /**
   * Append content to a note
   */
  private async appendContent(
    path: string,
    content: string,
    notebook?: string
  ): Promise<void> {
    // Use echo with SQL UPDATE (if supported) or file append
    // For now, this is a placeholder - full implementation would
    // require file system access or a CLI command for appending
    // The CLI might have a --content flag for notes add
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
