/**
 * CLI Adapter for jot commands
 * Central abstraction for all CLI interactions
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import type { ICliAdapter, CliResult, CliOptions, InstallationInfo } from "./types";
import { JotError, ErrorCodes } from "../utils/errors";

// =============================================================================
// CLI Adapter Configuration
// =============================================================================

export interface CliAdapterConfig {
  cliPath: string;
  defaultTimeout: number;
}

// =============================================================================
// CLI Adapter Implementation
// =============================================================================

export class CliAdapter implements ICliAdapter {
  private installationCache: InstallationInfo | null = null;

  constructor(
    private readonly pi: ExtensionAPI,
    private readonly config: CliAdapterConfig
  ) {}

  /**
   * Execute CLI command
   */
  async exec(
    command: string,
    args: string[],
    options?: CliOptions
  ): Promise<CliResult> {
    const fullArgs = [...args];

    // Add notebook flag if specified
    if (options?.notebook) {
      fullArgs.push("--notebook", options.notebook);
    }

    const timeout = options?.timeout ?? this.config.defaultTimeout;

    try {
      // Use pi.exec which handles process spawning
      const result = await this.pi.exec(command, fullArgs, {
        timeout,
        signal: options?.signal ?? undefined,
        env: options?.env,
      });

      return {
        code: result.code,
        stdout: result.stdout,
        stderr: result.stderr,
        timedOut: false,
      };
    } catch (error) {
      // Handle abort/timeout
      if (error instanceof Error) {
        if (error.name === "AbortError" || error.message.includes("abort")) {
          return {
            code: -1,
            stdout: "",
            stderr: "Command aborted",
            timedOut: false,
          };
        }
        if (error.message.includes("timeout") || error.message.includes("TIMEOUT")) {
          return {
            code: -1,
            stdout: "",
            stderr: `Command timed out after ${timeout}ms`,
            timedOut: true,
          };
        }
      }
      throw error;
    }
  }

  /**
   * Check if CLI is installed and accessible
   */
  async checkInstallation(): Promise<InstallationInfo> {
    // Return cached result if available
    if (this.installationCache) {
      return this.installationCache;
    }

    try {
      const result = await this.exec(this.config.cliPath, ["version"], {
        timeout: 5000,
      });

      if (result.code === 0) {
        // Parse version from output like "jot version 0.10.0"
        const versionMatch = result.stdout.match(/(?:jot\s+)?version\s+(\S+)/i);
        this.installationCache = {
          installed: true,
          version: versionMatch?.[1],
          path: this.config.cliPath,
        };
        return this.installationCache;
      }

      this.installationCache = { installed: false };
      return this.installationCache;
    } catch {
      this.installationCache = { installed: false };
      return this.installationCache;
    }
  }

  /**
   * Parse JSON output with error handling
   */
  parseJsonOutput<T>(stdout: string): T {
    const trimmed = stdout.trim();
    
    // Handle empty output
    if (!trimmed) {
      return [] as unknown as T;
    }

    try {
      return JSON.parse(trimmed) as T;
    } catch (error) {
      throw new JotError(
        `Failed to parse CLI output as JSON: ${error instanceof Error ? error.message : String(error)}`,
        ErrorCodes.PARSE_ERROR,
        { stdout: stdout.slice(0, 500) }
      );
    }
  }

  /**
   * Build notebook flag args
   */
  buildNotebookArgs(notebook?: string): string[] {
    return notebook ? ["--notebook", notebook] : [];
  }

  /**
   * Clear installation cache (for testing)
   */
  clearCache(): void {
    this.installationCache = null;
  }
}
