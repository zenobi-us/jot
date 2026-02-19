/**
 * Note-related TypeBox schemas
 */

import { Type, type Static } from "@sinclair/typebox";
import { StringEnum } from "@mariozechner/pi-ai";

// =============================================================================
// Note Summary (for list/search results)
// =============================================================================

export const NoteSummary = Type.Object(
  {
    path: Type.String({ description: "File path relative to notebook" }),
    title: Type.Optional(
      Type.String({ description: "Note title from frontmatter" })
    ),
    tags: Type.Optional(Type.Array(Type.String(), { description: "Tags" })),
    created: Type.Optional(
      Type.String({ description: "ISO 8601 creation timestamp" })
    ),
    modified: Type.Optional(
      Type.String({ description: "ISO 8601 modification timestamp" })
    ),
  },
  { description: "Summary of a note for listings" }
);

export type NoteSummaryType = Static<typeof NoteSummary>;

// =============================================================================
// Note Content (for full note retrieval)
// =============================================================================

export const NoteContent = Type.Object(
  {
    path: Type.String({ description: "File path relative to notebook" }),
    title: Type.Optional(Type.String({ description: "Note title" })),
    content: Type.String({ description: "Full markdown content" }),
    frontmatter: Type.Optional(
      Type.Record(Type.String(), Type.Unknown(), {
        description: "Frontmatter key-value pairs",
      })
    ),
    wordCount: Type.Optional(
      Type.Number({ description: "Word count of content" })
    ),
  },
  { description: "Full note content with metadata" }
);

export type NoteContentType = Static<typeof NoteContent>;

// =============================================================================
// Notebook Info
// =============================================================================

export const NotebookSource = StringEnum(
  ["registered", "ancestor", "explicit"] as const,
  {
    description: "How the notebook was discovered",
  }
);

export const NotebookInfo = Type.Object(
  {
    name: Type.String({ description: "Notebook display name" }),
    path: Type.String({ description: "Absolute path to notebook" }),
    source: NotebookSource,
    noteCount: Type.Optional(
      Type.Number({ description: "Number of notes in notebook" })
    ),
  },
  { description: "Information about a notebook" }
);

export type NotebookInfoType = Static<typeof NotebookInfo>;
