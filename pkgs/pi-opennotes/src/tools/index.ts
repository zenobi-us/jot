/**
 * Tool registration for pi-opennotes
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ExtensionConfig, ToolConfig } from "../config";

import { createSearchTool } from "./search.tool";
import { createListTool } from "./list.tool";
import { createGetTool } from "./get.tool";
import { createCreateTool } from "./create.tool";
import { createNotebooksTool } from "./notebooks.tool";
import { createViewsTool } from "./views.tool";

// =============================================================================
// Tool Registration
// =============================================================================

/**
 * Register all tools with pi
 */
export function registerTools(
  pi: ExtensionAPI,
  services: Services,
  config: ExtensionConfig
): void {
  const toolConfig: ToolConfig = {
    toolPrefix: config.toolPrefix,
  };

  // Register all 6 tools
  pi.registerTool(createSearchTool(services, toolConfig));
  pi.registerTool(createListTool(services, toolConfig));
  pi.registerTool(createGetTool(services, toolConfig));
  pi.registerTool(createCreateTool(services, toolConfig));
  pi.registerTool(createNotebooksTool(services, toolConfig));
  pi.registerTool(createViewsTool(services, toolConfig));
}

// =============================================================================
// Re-exports
// =============================================================================

export { createSearchTool } from "./search.tool";
export { createListTool } from "./list.tool";
export { createGetTool } from "./get.tool";
export { createCreateTool } from "./create.tool";
export { createNotebooksTool } from "./notebooks.tool";
export { createViewsTool } from "./views.tool";
