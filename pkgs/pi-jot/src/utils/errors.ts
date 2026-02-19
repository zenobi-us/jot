/**
 * Error handling for pi-jot
 * Provides structured errors with installation hints and recovery guidance
 */

// =============================================================================
// Error Codes
// =============================================================================

export const ErrorCodes = {
  // Installation errors (1xx)
  CLI_NOT_FOUND: "JOT_CLI_NOT_FOUND",
  CLI_VERSION_MISMATCH: "JOT_CLI_VERSION_MISMATCH",
  CLI_PERMISSION_DENIED: "JOT_CLI_PERMISSION_DENIED",

  // Notebook errors (2xx)
  NOTEBOOK_NOT_FOUND: "JOT_NOTEBOOK_NOT_FOUND",
  NOTEBOOK_INVALID_PATH: "JOT_NOTEBOOK_INVALID_PATH",
  NOTEBOOK_CONFIG_ERROR: "JOT_NOTEBOOK_CONFIG_ERROR",
  NOTEBOOK_NOT_REGISTERED: "JOT_NOTEBOOK_NOT_REGISTERED",

  // Query errors (3xx)
  INVALID_SQL: "JOT_INVALID_SQL",
  QUERY_TIMEOUT: "JOT_QUERY_TIMEOUT",
  QUERY_SECURITY: "JOT_QUERY_SECURITY",
  SEARCH_FAILED: "JOT_SEARCH_FAILED",

  // Note errors (4xx)
  NOTE_NOT_FOUND: "JOT_NOTE_NOT_FOUND",
  NOTE_INVALID_PATH: "JOT_NOTE_INVALID_PATH",
  NOTE_CREATE_FAILED: "JOT_NOTE_CREATE_FAILED",
  TEMPLATE_NOT_FOUND: "JOT_TEMPLATE_NOT_FOUND",

  // View errors (5xx)
  VIEW_NOT_FOUND: "JOT_VIEW_NOT_FOUND",
  VIEW_INVALID_PARAMS: "JOT_VIEW_INVALID_PARAMS",
  VIEW_EXECUTE_FAILED: "JOT_VIEW_EXECUTE_FAILED",
  VIEW_LIST_FAILED: "JOT_VIEW_LIST_FAILED",

  // System errors (9xx)
  NETWORK_ERROR: "JOT_NETWORK_ERROR",
  PARSE_ERROR: "JOT_PARSE_ERROR",
  UNKNOWN_ERROR: "JOT_UNKNOWN_ERROR",
  ABORTED: "JOT_ABORTED",
} as const;

export type ErrorCode = (typeof ErrorCodes)[keyof typeof ErrorCodes];

// =============================================================================
// Installation Hints
// =============================================================================

const INSTALLATION_HINT = `
**Jot CLI is not installed or not in PATH.**

Install Jot using one of these methods:

**Option 1: Go Install (Recommended)**
\`\`\`bash
go install github.com/zenobi-us/jot@latest
\`\`\`

**Option 2: Download Binary**
Download from: https://github.com/zenobi-us/jot/releases

**Option 3: Build from Source**
\`\`\`bash
git clone https://github.com/zenobi-us/jot.git
cd jot
go build -o jot .
sudo mv jot /usr/local/bin/
\`\`\`

**Verify Installation:**
\`\`\`bash
jot version
\`\`\`

**If already installed, ensure it's in PATH:**
\`\`\`bash
# Check current PATH
echo $PATH

# Add to PATH (bash/zsh)
export PATH="$PATH:$HOME/go/bin"
\`\`\`
`.trim();

const VERSION_MISMATCH_HINT = `
**Jot CLI version is incompatible.**

This extension requires Jot v0.10.0 or later.

**Upgrade Jot:**
\`\`\`bash
go install github.com/zenobi-us/jot@latest
\`\`\`

**Check current version:**
\`\`\`bash
jot version
\`\`\`
`.trim();

// =============================================================================
// Default Hints
// =============================================================================

function getDefaultHint(code: ErrorCode): string {
  const hints: Record<ErrorCode, string> = {
    [ErrorCodes.CLI_NOT_FOUND]: INSTALLATION_HINT,
    [ErrorCodes.CLI_VERSION_MISMATCH]: VERSION_MISMATCH_HINT,
    [ErrorCodes.CLI_PERMISSION_DENIED]:
      "Check file permissions on the jot binary. Try: chmod +x $(which jot)",

    [ErrorCodes.NOTEBOOK_NOT_FOUND]:
      "No notebook found in current directory or ancestors.\n" +
      "Either:\n" +
      "1. Navigate to a directory containing .jot.json\n" +
      "2. Specify notebook path: { notebook: '/path/to/notebook' }\n" +
      "3. Create a notebook: jot notebook create 'My Notes'",

    [ErrorCodes.NOTEBOOK_INVALID_PATH]:
      "The specified notebook path does not exist or is not a valid notebook.\n" +
      "A valid notebook contains a .jot.json config file.",

    [ErrorCodes.NOTEBOOK_CONFIG_ERROR]:
      "The notebook's .jot.json file is invalid.\n" +
      "Check for JSON syntax errors or missing required fields.",

    [ErrorCodes.NOTEBOOK_NOT_REGISTERED]:
      "The notebook is not registered in your global config.\n" +
      "Register it with: jot notebook register /path/to/notebook",

    [ErrorCodes.INVALID_SQL]:
      "Only SELECT and WITH queries are allowed (read-only).\n" +
      "Example: SELECT * FROM read_markdown('**/*.md') LIMIT 10\n" +
      "Docs: https://github.com/zenobi-us/jot/blob/main/docs/sql-guide.md",

    [ErrorCodes.QUERY_TIMEOUT]:
      "Query exceeded 30-second timeout.\n" +
      "Simplify your query or add LIMIT to reduce results.",

    [ErrorCodes.QUERY_SECURITY]:
      "Path traversal (../) is not allowed in queries.\n" +
      "Use paths relative to the notebook root.",

    [ErrorCodes.SEARCH_FAILED]:
      "Search query failed. Check query syntax and try again.",

    [ErrorCodes.NOTE_NOT_FOUND]:
      "The specified note does not exist.\n" +
      "Use jot_list to see available notes.\n" +
      "Ensure path is relative to notebook root (e.g., 'notes/my-note.md').",

    [ErrorCodes.NOTE_INVALID_PATH]:
      "Invalid note path format.\n" +
      "Paths should be relative to notebook root and end with .md",

    [ErrorCodes.NOTE_CREATE_FAILED]:
      "Failed to create note.\n" +
      "Check that:\n" +
      "1. Notebook has write permissions\n" +
      "2. Target directory exists\n" +
      "3. A note with this name doesn't already exist",

    [ErrorCodes.TEMPLATE_NOT_FOUND]:
      "The specified template does not exist.\n" +
      "Check available templates in notebook's .jot.json file.\n" +
      "Templates are defined in the 'templates' section.",

    [ErrorCodes.VIEW_NOT_FOUND]:
      "The specified view does not exist.\n" +
      "Use jot_views (without arguments) to list available views.\n" +
      "Built-in views: today, recent, kanban, untagged, orphans, broken-links",

    [ErrorCodes.VIEW_INVALID_PARAMS]:
      "Invalid parameters for this view.\n" +
      "Use jot_views to see view parameter requirements.",

    [ErrorCodes.VIEW_EXECUTE_FAILED]:
      "View execution failed.\n" +
      "The view's SQL query may be invalid or target non-existent data.",

    [ErrorCodes.VIEW_LIST_FAILED]:
      "Failed to list views. Check notebook configuration.",

    [ErrorCodes.NETWORK_ERROR]:
      "Network error occurred. Check your connection.",

    [ErrorCodes.PARSE_ERROR]:
      "Failed to parse CLI output.\n" +
      "This may indicate a version mismatch. Try: jot version",

    [ErrorCodes.ABORTED]:
      "Operation was cancelled.\n" +
      "This is normal if you interrupted the operation.",

    [ErrorCodes.UNKNOWN_ERROR]:
      "An unexpected error occurred.\n" +
      "Check the error details for more information.\n" +
      "If this persists, please report at: https://github.com/zenobi-us/jot/issues",
  };

  return hints[code] ?? hints[ErrorCodes.UNKNOWN_ERROR];
}

// =============================================================================
// Recoverable Error Detection
// =============================================================================

const RECOVERABLE_CODES: ErrorCode[] = [
  ErrorCodes.CLI_NOT_FOUND,
  ErrorCodes.NOTEBOOK_NOT_FOUND,
  ErrorCodes.NOTE_NOT_FOUND,
  ErrorCodes.VIEW_NOT_FOUND,
  ErrorCodes.TEMPLATE_NOT_FOUND,
  ErrorCodes.NOTEBOOK_NOT_REGISTERED,
];

function isRecoverableError(code: ErrorCode): boolean {
  return RECOVERABLE_CODES.includes(code);
}

// =============================================================================
// Error Details Type
// =============================================================================

export interface ErrorDetails {
  [key: string]: unknown;
}

// =============================================================================
// Tool Result Type (simplified for error output)
// =============================================================================

export interface ToolResultContent {
  type: "text";
  text: string;
}

export interface ToolResult {
  content: ToolResultContent[];
  isError?: boolean;
}

// =============================================================================
// Error Response Type
// =============================================================================

export interface ErrorResponse {
  error: true;
  message: string;
  code: ErrorCode;
  hint?: string;
  details?: ErrorDetails;
  recoverable: boolean;
}

// =============================================================================
// JotError Class
// =============================================================================

export class JotError extends Error {
  public readonly code: ErrorCode;
  public readonly details?: ErrorDetails;
  public readonly hint: string;
  public readonly recoverable: boolean;

  constructor(
    message: string,
    code: ErrorCode,
    details?: ErrorDetails,
    hint?: string
  ) {
    super(message);
    this.name = "JotError";
    this.code = code;
    this.details = details;
    this.hint = hint ?? getDefaultHint(code);
    this.recoverable = isRecoverableError(code);

    // Maintain proper stack trace
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, JotError);
    }
  }

  /**
   * Convert to JSON response format
   */
  toResponse(): ErrorResponse {
    return {
      error: true,
      message: this.message,
      code: this.code,
      hint: this.hint,
      details: this.details,
      recoverable: this.recoverable,
    };
  }

  /**
   * Convert to tool result format for LLM consumption
   */
  toToolResult(): ToolResult {
    return {
      content: [
        {
          type: "text",
          text: this.formatForLLM(),
        },
      ],
      isError: true,
    };
  }

  /**
   * Format error for LLM consumption
   */
  private formatForLLM(): string {
    let result = `**Error**: ${this.message}\n`;
    result += `**Code**: ${this.code}\n`;

    if (this.hint) {
      result += `\n**How to fix**:\n${this.hint}\n`;
    }

    if (this.recoverable) {
      result += `\n*This error can be resolved by the user.*`;
    }

    return result;
  }
}

// =============================================================================
// Error Handling Utilities
// =============================================================================

/**
 * Wrap any error as JotError
 */
export function wrapError(error: unknown, defaultCode?: ErrorCode): JotError {
  if (error instanceof JotError) {
    return error;
  }

  const message = error instanceof Error ? error.message : String(error);
  const code = defaultCode ?? ErrorCodes.UNKNOWN_ERROR;

  return new JotError(message, code, {
    originalError: String(error),
  });
}

/**
 * Execute a function and convert any errors to JotError
 */
export async function withErrorHandling<T>(
  fn: () => Promise<T>,
  defaultCode?: ErrorCode
): Promise<T> {
  try {
    return await fn();
  } catch (error) {
    throw wrapError(error, defaultCode);
  }
}
