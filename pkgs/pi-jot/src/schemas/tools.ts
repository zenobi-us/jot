/**
 * Tool parameter schemas for all 6 tools
 */

import { Type, type Static } from "@sinclair/typebox";
import {
  NotebookParam,
  LimitParam,
  OffsetParam,
  SortField,
  SortOrder,
} from "./common";

// =============================================================================
// Search Tool Parameters
// =============================================================================

export const BooleanFilters = Type.Object(
  {
    and: Type.Optional(
      Type.Array(Type.String(), {
        description: "AND conditions: field=value pairs (all must match)",
      })
    ),
    or: Type.Optional(
      Type.Array(Type.String(), {
        description: "OR conditions: field=value pairs (any must match)",
      })
    ),
    not: Type.Optional(
      Type.Array(Type.String(), {
        description: "NOT conditions: field=value pairs (exclusions)",
      })
    ),
  },
  { description: "Boolean filter conditions" }
);

export const SearchParams = Type.Object(
  {
    query: Type.Optional(
      Type.String({
        description: "Text to search for in note titles and content",
      })
    ),
    sql: Type.Optional(
      Type.String({
        description:
          "Raw SQL query (SELECT/WITH only). Use read_markdown('**/*.md') to access notes.",
      })
    ),
    fuzzy: Type.Optional(
      Type.Boolean({
        description: "Enable fuzzy matching (typo-tolerant, ranked results)",
        default: false,
      })
    ),
    filters: Type.Optional(BooleanFilters),
    notebook: NotebookParam,
    limit: LimitParam,
    offset: OffsetParam,
  },
  {
    description:
      "Search parameters - use query for text search, sql for SQL queries, or filters for boolean search",
  }
);

export type SearchParamsType = Static<typeof SearchParams>;

// =============================================================================
// List Tool Parameters
// =============================================================================

export const ListParams = Type.Object(
  {
    notebook: NotebookParam,
    sortBy: Type.Optional(SortField),
    sortOrder: Type.Optional(SortOrder),
    limit: LimitParam,
    offset: OffsetParam,
    pattern: Type.Optional(
      Type.String({
        description: "Glob pattern to filter paths (default: **/*.md)",
        default: "**/*.md",
      })
    ),
  },
  { description: "List notes with filtering and sorting options" }
);

export type ListParamsType = Static<typeof ListParams>;

// =============================================================================
// Get Tool Parameters
// =============================================================================

export const GetParams = Type.Object(
  {
    path: Type.String({
      description: "Path to the note file (relative to notebook root)",
    }),
    notebook: NotebookParam,
    includeContent: Type.Optional(
      Type.Boolean({
        description:
          "Whether to include full markdown content (default: true)",
        default: true,
      })
    ),
  },
  { description: "Get a specific note by path" }
);

export type GetParamsType = Static<typeof GetParams>;

// =============================================================================
// Create Tool Parameters
// =============================================================================

export const CreateParams = Type.Object(
  {
    title: Type.String({
      description: "Note title (required)",
      minLength: 1,
    }),
    path: Type.Optional(
      Type.String({
        description: "Directory path within notebook (default: root)",
      })
    ),
    template: Type.Optional(
      Type.String({
        description: "Template name to use (must exist in notebook)",
      })
    ),
    content: Type.Optional(
      Type.String({
        description: "Initial markdown content (after frontmatter)",
      })
    ),
    data: Type.Optional(
      Type.Record(
        Type.String(),
        Type.Union([
          Type.String(),
          Type.Number(),
          Type.Boolean(),
          Type.Array(Type.String()),
        ]),
        {
          description: "Frontmatter fields as key-value pairs",
        }
      )
    ),
    notebook: NotebookParam,
  },
  { description: "Create a new note" }
);

export type CreateParamsType = Static<typeof CreateParams>;

// =============================================================================
// Notebooks Tool Parameters
// =============================================================================

export const NotebooksParams = Type.Object(
  {},
  { description: "List all available notebooks (no parameters required)" }
);

export type NotebooksParamsType = Static<typeof NotebooksParams>;

// =============================================================================
// Views Tool Parameters
// =============================================================================

export const ViewsParams = Type.Object(
  {
    view: Type.Optional(
      Type.String({
        description: "View name to execute. Omit to list all views.",
      })
    ),
    params: Type.Optional(
      Type.Record(Type.String(), Type.String(), {
        description: "Parameters for the view (e.g., {'status': 'todo,done'})",
      })
    ),
    notebook: NotebookParam,
    limit: LimitParam,
    offset: OffsetParam,
  },
  {
    description:
      "List available views (no view param) or execute a named view (with view param)",
  }
);

export type ViewsParamsType = Static<typeof ViewsParams>;
