/**
 * pi-opennotes Extension
 *
 * Integrates OpenNotes into the pi coding agent, enabling AI assistants to:
 * - Search and query notes using DuckDB SQL
 * - Create and manage notes in notebooks
 * - Execute reusable views
 * - Access note metadata and relationships
 *
 * @see https://github.com/zenobi-us/opennotes
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { createServices } from "./services";
import { registerTools } from "./tools";
import { getConfig } from "./config";
import { OpenNotesError, ErrorCodes } from "./utils/errors";

// =============================================================================
// Extension Entry Point
// =============================================================================

export default function piOpennotes(pi: ExtensionAPI): void {
  // Load configuration
  const config = getConfig(pi);

  // Create services
  const services = createServices(pi, config);

  // Register tools
  registerTools(pi, services, config);

  // Listen for session start to check CLI installation
  pi.on("session_start", async (_event, ctx) => {
    try {
      const installation = await services.cli.checkInstallation();

      if (!installation.installed) {
        ctx.ui.notify(
          "⚠️ OpenNotes CLI not found. Install with: go install github.com/zenobi-us/opennotes@latest",
          "warning"
        );
      } else if (installation.version) {
        // Optional: show version in status
        // ctx.ui.setStatus("opennotes", `OpenNotes ${installation.version}`);
      }
    } catch {
      // Silently ignore check failures during session start
    }
  });

  // Optional: Register custom commands
  pi.registerCommand("opennotes", {
    description: "OpenNotes status and info",
    handler: async (args, ctx) => {
      try {
        const installation = await services.cli.checkInstallation();

        if (!installation.installed) {
          ctx.ui.notify(
            "OpenNotes CLI is not installed.\n\nInstall with:\n  go install github.com/zenobi-us/opennotes@latest",
            "error"
          );
          return;
        }

        const notebooks = await services.notebook.listNotebooks();
        const lines = [
          `**OpenNotes Status**`,
          ``,
          `Version: ${installation.version ?? "unknown"}`,
          `CLI Path: ${installation.path ?? config.cliPath}`,
          ``,
          `**Notebooks:** ${notebooks.notebooks.length}`,
        ];

        if (notebooks.current) {
          lines.push(`**Current:** ${notebooks.current.name} (${notebooks.current.path})`);
        }

        ctx.ui.notify(lines.join("\n"), "info");
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        ctx.ui.notify(`Error: ${message}`, "error");
      }
    },
  });
}

// =============================================================================
// Re-exports for programmatic use
// =============================================================================

export { createServices, type Services } from "./services";
export { registerTools } from "./tools";
export { getConfig, type ExtensionConfig, type ToolConfig } from "./config";
export { OpenNotesError, ErrorCodes, type ErrorCode } from "./utils/errors";
export * from "./schemas";
