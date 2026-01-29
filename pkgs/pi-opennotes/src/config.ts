/**
 * Extension configuration for pi-opennotes
 * Supports pi package config schema and environment variable overrides
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

// =============================================================================
// Configuration Interface
// =============================================================================

export interface ExtensionConfig {
  /**
   * Prefix for tool names (e.g., 'opennotes_search')
   * @default "opennotes_"
   */
  toolPrefix: string;

  /**
   * Default number of results per page
   * @default 50
   */
  defaultPageSize: number;

  /**
   * Path to opennotes CLI binary
   * @default "opennotes"
   */
  cliPath: string;

  /**
   * CLI command timeout in milliseconds
   * @default 30000
   */
  cliTimeout: number;

  /**
   * Maximum output bytes before truncation (pi default)
   * @default 51200 (50KB)
   */
  maxOutputBytes: number;

  /**
   * Maximum output lines before truncation (pi default)
   * @default 2000
   */
  maxOutputLines: number;

  /**
   * Budget ratio for content vs metadata (0.0-1.0)
   * @default 0.75
   */
  budgetRatio: number;
}

// =============================================================================
// Default Configuration
// =============================================================================

const DEFAULT_CONFIG: ExtensionConfig = {
  toolPrefix: "opennotes_",
  defaultPageSize: 50,
  cliPath: "opennotes",
  cliTimeout: 30000,
  maxOutputBytes: 50 * 1024, // 50KB
  maxOutputLines: 2000,
  budgetRatio: 0.75,
};

// =============================================================================
// Configuration Loading
// =============================================================================

/**
 * Get configuration from pi API with fallbacks to defaults and env vars
 */
export function getConfig(pi: ExtensionAPI): ExtensionConfig {
  // Start with defaults
  const config = { ...DEFAULT_CONFIG };

  // Try to get pi package config (if available)
  try {
    const piConfig = pi.getConfig?.() as Partial<ExtensionConfig> | undefined;
    if (piConfig) {
      if (typeof piConfig.toolPrefix === "string") {
        config.toolPrefix = piConfig.toolPrefix;
      }
      if (typeof piConfig.defaultPageSize === "number") {
        config.defaultPageSize = piConfig.defaultPageSize;
      }
      if (typeof piConfig.cliPath === "string") {
        config.cliPath = piConfig.cliPath;
      }
      if (typeof piConfig.cliTimeout === "number") {
        config.cliTimeout = piConfig.cliTimeout;
      }
    }
  } catch {
    // Config not available, use defaults
  }

  // Environment variable overrides (highest priority)
  const envPrefix = process.env.OPENNOTES_TOOL_PREFIX;
  if (envPrefix) {
    config.toolPrefix = envPrefix;
  }

  const envPageSize = process.env.OPENNOTES_PAGE_SIZE;
  if (envPageSize) {
    const parsed = parseInt(envPageSize, 10);
    if (!isNaN(parsed) && parsed > 0) {
      config.defaultPageSize = parsed;
    }
  }

  const envCliPath = process.env.OPENNOTES_CLI_PATH;
  if (envCliPath) {
    config.cliPath = envCliPath;
  }

  const envTimeout = process.env.OPENNOTES_CLI_TIMEOUT;
  if (envTimeout) {
    const parsed = parseInt(envTimeout, 10);
    if (!isNaN(parsed) && parsed > 0) {
      config.cliTimeout = parsed;
    }
  }

  return config;
}

// =============================================================================
// Tool Configuration
// =============================================================================

export interface ToolConfig {
  toolPrefix: string;
}

/**
 * Get tool-specific config from extension config
 */
export function getToolConfig(config: ExtensionConfig): ToolConfig {
  return {
    toolPrefix: config.toolPrefix,
  };
}
