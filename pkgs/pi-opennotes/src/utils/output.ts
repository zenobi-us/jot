/**
 * Output formatting utilities for pi-opennotes
 * Formats results for LLM consumption
 */

import type { NoteSummary, NoteContent, NotebookInfo, ViewDefinition, PaginationMeta } from "../services/types";

/**
 * Format note summary for display
 */
export function formatNoteSummary(note: NoteSummary): string {
  const parts: string[] = [`**${note.path}**`];

  if (note.title) {
    parts[0] = `**${note.title}** (${note.path})`;
  }

  if (note.tags && note.tags.length > 0) {
    parts.push(`Tags: ${note.tags.join(", ")}`);
  }

  if (note.modified) {
    parts.push(`Modified: ${note.modified}`);
  }

  return parts.join(" | ");
}

/**
 * Format full note content for display
 */
export function formatNoteContent(note: NoteContent): string {
  const lines: string[] = [];

  lines.push(`# ${note.title ?? note.path}`);
  lines.push("");

  if (note.frontmatter && Object.keys(note.frontmatter).length > 0) {
    lines.push("**Frontmatter:**");
    for (const [key, value] of Object.entries(note.frontmatter)) {
      lines.push(`- ${key}: ${JSON.stringify(value)}`);
    }
    lines.push("");
  }

  if (note.wordCount !== undefined) {
    lines.push(`*${note.wordCount} words*`);
    lines.push("");
  }

  lines.push("---");
  lines.push("");
  lines.push(note.content);

  return lines.join("\n");
}

/**
 * Format notebook info for display
 */
export function formatNotebookInfo(notebook: NotebookInfo): string {
  const parts = [`**${notebook.name}**`, `Path: ${notebook.path}`];

  if (notebook.noteCount !== undefined) {
    parts.push(`Notes: ${notebook.noteCount}`);
  }

  parts.push(`Source: ${notebook.source}`);

  return parts.join(" | ");
}

/**
 * Format view definition for display
 */
export function formatViewDefinition(view: ViewDefinition): string {
  const lines: string[] = [];

  lines.push(`**${view.name}** (${view.origin})`);

  if (view.description) {
    lines.push(`  ${view.description}`);
  }

  if (view.parameters && view.parameters.length > 0) {
    lines.push("  Parameters:");
    for (const param of view.parameters) {
      const required = param.required ? " (required)" : "";
      const def = param.default ? ` [default: ${param.default}]` : "";
      lines.push(`    - ${param.name}: ${param.type}${required}${def}`);
      if (param.description) {
        lines.push(`      ${param.description}`);
      }
    }
  }

  return lines.join("\n");
}

/**
 * Format pagination metadata for display
 */
export function formatPaginationMeta(pagination: PaginationMeta): string {
  const parts: string[] = [
    `Showing ${pagination.returned} of ${pagination.total}`,
    `Page ${pagination.page}`,
  ];

  if (pagination.hasMore && pagination.nextOffset !== undefined) {
    parts.push(`More available (next offset: ${pagination.nextOffset})`);
  }

  return parts.join(" | ");
}

/**
 * Format search results for LLM
 */
export function formatSearchResults(
  results: NoteSummary[],
  pagination: PaginationMeta,
  query: { type: string; executed: string }
): string {
  const lines: string[] = [];

  lines.push(`## Search Results (${query.type})`);
  lines.push("");
  lines.push(formatPaginationMeta(pagination));
  lines.push("");

  if (results.length === 0) {
    lines.push("*No results found*");
  } else {
    for (const note of results) {
      lines.push(`- ${formatNoteSummary(note)}`);
    }
  }

  if (pagination.hasMore) {
    lines.push("");
    lines.push(`*To fetch more results, use offset: ${pagination.nextOffset}*`);
  }

  return lines.join("\n");
}

/**
 * Format list results for LLM
 */
export function formatListResults(
  notes: NoteSummary[],
  notebook: NotebookInfo,
  pagination: PaginationMeta
): string {
  const lines: string[] = [];

  lines.push(`## Notes in ${notebook.name}`);
  lines.push("");
  lines.push(formatPaginationMeta(pagination));
  lines.push("");

  if (notes.length === 0) {
    lines.push("*No notes found*");
  } else {
    for (const note of notes) {
      lines.push(`- ${formatNoteSummary(note)}`);
    }
  }

  if (pagination.hasMore) {
    lines.push("");
    lines.push(`*To fetch more notes, use offset: ${pagination.nextOffset}*`);
  }

  return lines.join("\n");
}

/**
 * Format notebooks list for LLM
 */
export function formatNotebooksList(
  notebooks: NotebookInfo[],
  current: NotebookInfo | null
): string {
  const lines: string[] = [];

  lines.push("## Available Notebooks");
  lines.push("");

  if (current) {
    lines.push(`**Current:** ${formatNotebookInfo(current)}`);
    lines.push("");
  }

  if (notebooks.length === 0) {
    lines.push("*No notebooks registered*");
  } else {
    lines.push("**All notebooks:**");
    for (const nb of notebooks) {
      const isCurrent = current && nb.path === current.path ? " ← current" : "";
      lines.push(`- ${formatNotebookInfo(nb)}${isCurrent}`);
    }
  }

  return lines.join("\n");
}

/**
 * Format views list for LLM
 */
export function formatViewsList(views: ViewDefinition[], notebook: NotebookInfo): string {
  const lines: string[] = [];

  lines.push(`## Views in ${notebook.name}`);
  lines.push("");

  if (views.length === 0) {
    lines.push("*No views available*");
  } else {
    // Group by origin
    const builtIn = views.filter((v) => v.origin === "built-in");
    const notebookViews = views.filter((v) => v.origin === "notebook");
    const globalViews = views.filter((v) => v.origin === "global");

    if (builtIn.length > 0) {
      lines.push("### Built-in Views");
      for (const view of builtIn) {
        lines.push(formatViewDefinition(view));
      }
      lines.push("");
    }

    if (notebookViews.length > 0) {
      lines.push("### Notebook Views");
      for (const view of notebookViews) {
        lines.push(formatViewDefinition(view));
      }
      lines.push("");
    }

    if (globalViews.length > 0) {
      lines.push("### Global Views");
      for (const view of globalViews) {
        lines.push(formatViewDefinition(view));
      }
      lines.push("");
    }
  }

  return lines.join("\n");
}

/**
 * Format view execution results for LLM
 */
export function formatViewResults(
  viewName: string,
  results: Record<string, unknown>[],
  pagination: PaginationMeta
): string {
  const lines: string[] = [];

  lines.push(`## View: ${viewName}`);
  lines.push("");
  lines.push(formatPaginationMeta(pagination));
  lines.push("");

  if (results.length === 0) {
    lines.push("*No results*");
  } else {
    // Format as table if uniform structure
    lines.push("```json");
    lines.push(JSON.stringify(results, null, 2));
    lines.push("```");
  }

  if (pagination.hasMore) {
    lines.push("");
    lines.push(`*To fetch more results, use offset: ${pagination.nextOffset}*`);
  }

  return lines.join("\n");
}

/**
 * Format note creation result for LLM
 */
export function formatCreateResult(
  created: { path: string; absolutePath: string; title: string },
  notebook: NotebookInfo
): string {
  const lines: string[] = [];

  lines.push("## Note Created");
  lines.push("");
  lines.push(`**Title:** ${created.title}`);
  lines.push(`**Path:** ${created.path}`);
  lines.push(`**Full path:** ${created.absolutePath}`);
  lines.push(`**Notebook:** ${notebook.name}`);

  return lines.join("\n");
}
