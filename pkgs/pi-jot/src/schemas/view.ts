/**
 * View-related TypeBox schemas
 */

import { Type, type Static } from "@sinclair/typebox";
import { StringEnum } from "@mariozechner/pi-ai";

// =============================================================================
// View Origin
// =============================================================================

export const ViewOrigin = StringEnum(
  ["built-in", "notebook", "global"] as const,
  {
    description: "Where the view is defined",
  }
);

export type ViewOriginType = Static<typeof ViewOrigin>;

// =============================================================================
// View Parameter Definition
// =============================================================================

export const ViewParameter = Type.Object(
  {
    name: Type.String({ description: "Parameter name" }),
    type: Type.String({ description: "Parameter type" }),
    required: Type.Boolean({ description: "Whether parameter is required" }),
    default: Type.Optional(Type.String({ description: "Default value" })),
    description: Type.Optional(Type.String({ description: "Parameter description" })),
  },
  { description: "Definition of a view parameter" }
);

export type ViewParameterType = Static<typeof ViewParameter>;

// =============================================================================
// View Definition
// =============================================================================

export const ViewDefinition = Type.Object(
  {
    name: Type.String({ description: "View name" }),
    origin: ViewOrigin,
    description: Type.Optional(Type.String({ description: "View description" })),
    parameters: Type.Optional(
      Type.Array(ViewParameter, { description: "View parameters" })
    ),
  },
  { description: "Definition of an available view" }
);

export type ViewDefinitionType = Static<typeof ViewDefinition>;
