/**
 * Common TypeBox schemas shared across tools
 */

import { Type, type Static } from "@sinclair/typebox";
import { StringEnum } from "@mariozechner/pi-ai";

// =============================================================================
// Pagination
// =============================================================================

export const PaginationMeta = Type.Object(
  {
    total: Type.Number({ description: "Total number of results" }),
    returned: Type.Number({ description: "Number returned in this response" }),
    page: Type.Number({ description: "Current page (1-indexed)" }),
    pageSize: Type.Number({ description: "Results per page" }),
    hasMore: Type.Boolean({ description: "Whether more results exist" }),
    nextOffset: Type.Optional(
      Type.Number({ description: "Offset for next page" })
    ),
  },
  { description: "Pagination metadata for large result sets" }
);

export type PaginationMetaType = Static<typeof PaginationMeta>;

// =============================================================================
// Sort Options
// =============================================================================

export const SortField = StringEnum(
  ["modified", "created", "title", "path"] as const,
  {
    description: "Field to sort by",
    default: "modified",
  }
);

export type SortFieldType = Static<typeof SortField>;

export const SortOrder = StringEnum(["asc", "desc"] as const, {
  description: "Sort order (asc or desc)",
  default: "desc",
});

export type SortOrderType = Static<typeof SortOrder>;

// =============================================================================
// Notebook Parameter
// =============================================================================

export const NotebookParam = Type.Optional(
  Type.String({
    description: "Path to notebook. Omit to use current context.",
  })
);

// =============================================================================
// Pagination Parameters
// =============================================================================

export const LimitParam = Type.Optional(
  Type.Number({
    description: "Maximum results to return (default: 50, max: 1000)",
    default: 50,
    minimum: 1,
    maximum: 1000,
  })
);

export const OffsetParam = Type.Optional(
  Type.Number({
    description: "Offset for pagination (default: 0)",
    default: 0,
    minimum: 0,
  })
);
