---
id: e2x2x2x2
title: Design Error Handling Strategy
created_at: 2026-01-29T09:40:00+10:30
updated_at: 2026-01-29T09:40:00+10:30
status: done
epic_id: 1f41631e
phase_id: 43842f12
assigned_to: null
---

# Design Error Handling Strategy

## Objective

Design comprehensive error handling for pi-opennotes, including error classification, installation hints, and user-friendly messages that help resolve issues.

## Completed Steps

- [x] Define error classification system
- [x] Design OpenNotesError class
- [x] Create error code enumeration
- [x] Design installation hints strategy
- [x] Define error response format
- [x] Document recovery actions
- [x] Create error handling middleware
- [x] Design validation layer

---

## Design Principles

1. **Fail with guidance**: Every error includes actionable next steps
2. **Installation awareness**: CLI-missing errors include full installation instructions
3. **Context preservation**: Errors include relevant context for debugging
4. **User-first messages**: Avoid technical jargon in user-facing messages
5. **Structured format**: All errors follow consistent JSON schema

---

## Error Classification

### Error Categories

```
┌─────────────────────────────────────────────────────────────┐
│                    Error Categories                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │  INSTALLATION   │    │   NOTEBOOK      │                │
│  │  CLI not found  │    │   Not found     │                │
│  │  Wrong version  │    │   Invalid path  │                │
│  │  Permission     │    │   Config error  │                │
│  └─────────────────┘    └─────────────────┘                │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │    QUERY        │    │    NOTE         │                │
│  │  Invalid SQL    │    │  Not found      │                │
│  │  Timeout        │    │  Invalid path   │                │
│  │  Security       │    │  Template error │                │
│  └─────────────────┘    └─────────────────┘                │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │    VIEW         │    │   SYSTEM        │                │
│  │  Not found      │    │  Network        │                │
│  │  Invalid params │    │  Disk full      │                │
│  │  Execute failed │    │  Permissions    │                │
│  └─────────────────┘    └─────────────────┘                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Error Codes

```typescript
// src/utils/errors.ts

export const ErrorCodes = {
  // Installation errors (1xx)
  CLI_NOT_FOUND: "OPENNOTES_CLI_NOT_FOUND",
  CLI_VERSION_MISMATCH: "OPENNOTES_CLI_VERSION_MISMATCH",
  CLI_PERMISSION_DENIED: "OPENNOTES_CLI_PERMISSION_DENIED",

  // Notebook errors (2xx)
  NOTEBOOK_NOT_FOUND: "OPENNOTES_NOTEBOOK_NOT_FOUND",
  NOTEBOOK_INVALID_PATH: "OPENNOTES_NOTEBOOK_INVALID_PATH",
  NOTEBOOK_CONFIG_ERROR: "OPENNOTES_NOTEBOOK_CONFIG_ERROR",
  NOTEBOOK_NOT_REGISTERED: "OPENNOTES_NOTEBOOK_NOT_REGISTERED",

  // Query errors (3xx)
  INVALID_SQL: "OPENNOTES_INVALID_SQL",
  QUERY_TIMEOUT: "OPENNOTES_QUERY_TIMEOUT",
  QUERY_SECURITY: "OPENNOTES_QUERY_SECURITY",
  SEARCH_FAILED: "OPENNOTES_SEARCH_FAILED",

  // Note errors (4xx)
  NOTE_NOT_FOUND: "OPENNOTES_NOTE_NOT_FOUND",
  NOTE_INVALID_PATH: "OPENNOTES_NOTE_INVALID_PATH",
  NOTE_CREATE_FAILED: "OPENNOTES_NOTE_CREATE_FAILED",
  TEMPLATE_NOT_FOUND: "OPENNOTES_TEMPLATE_NOT_FOUND",

  // View errors (5xx)
  VIEW_NOT_FOUND: "OPENNOTES_VIEW_NOT_FOUND",
  VIEW_INVALID_PARAMS: "OPENNOTES_VIEW_INVALID_PARAMS",
  VIEW_EXECUTE_FAILED: "OPENNOTES_VIEW_EXECUTE_FAILED",
  VIEW_LIST_FAILED: "OPENNOTES_VIEW_LIST_FAILED",

  // System errors (9xx)
  NETWORK_ERROR: "OPENNOTES_NETWORK_ERROR",
  PARSE_ERROR: "OPENNOTES_PARSE_ERROR",
  UNKNOWN_ERROR: "OPENNOTES_UNKNOWN_ERROR",
  ABORTED: "OPENNOTES_ABORTED",
} as const;

export type ErrorCode = (typeof ErrorCodes)[keyof typeof ErrorCodes];
```

---

## OpenNotesError Class

```typescript
// src/utils/errors.ts

export interface ErrorDetails {
  [key: string]: unknown;
}

export class OpenNotesError extends Error {
  public readonly code: ErrorCode;
  public readonly details?: ErrorDetails;
  public readonly hint?: string;
  public readonly recoverable: boolean;

  constructor(
    message: string,
    code: ErrorCode,
    details?: ErrorDetails,
    hint?: string
  ) {
    super(message);
    this.name = "OpenNotesError";
    this.code = code;
    this.details = details;
    this.hint = hint ?? getDefaultHint(code);
    this.recoverable = isRecoverableError(code);

    // Maintain proper stack trace
    Error.captureStackTrace?.(this, OpenNotesError);
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
   * Format for LLM consumption
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

  private formatForLLM(): string {
    let result = `**Error**: ${this.message}\n`;
    
    if (this.hint) {
      result += `\n**How to fix**:\n${this.hint}\n`;
    }
    
    if (this.recoverable) {
      result += `\n*This error can be resolved by the user.*`;
    }
    
    return result;
  }
}

export interface ErrorResponse {
  error: true;
  message: string;
  code: ErrorCode;
  hint?: string;
  details?: ErrorDetails;
  recoverable: boolean;
}

function isRecoverableError(code: ErrorCode): boolean {
  const recoverableCodes: ErrorCode[] = [
    ErrorCodes.CLI_NOT_FOUND,
    ErrorCodes.NOTEBOOK_NOT_FOUND,
    ErrorCodes.NOTE_NOT_FOUND,
    ErrorCodes.VIEW_NOT_FOUND,
    ErrorCodes.TEMPLATE_NOT_FOUND,
    ErrorCodes.NOTEBOOK_NOT_REGISTERED,
  ];
  return recoverableCodes.includes(code);
}
```

---

## Installation Hints

### CLI Not Found - Full Installation Guide

```typescript
// src/utils/hints.ts

export const INSTALLATION_HINTS = {
  [ErrorCodes.CLI_NOT_FOUND]: `
**OpenNotes CLI is not installed or not in PATH.**

Install OpenNotes using one of these methods:

**Option 1: Go Install (Recommended)**
\`\`\`bash
go install github.com/zenobi-us/opennotes@latest
\`\`\`

**Option 2: Download Binary**
Download from: https://github.com/zenobi-us/opennotes/releases

**Option 3: Build from Source**
\`\`\`bash
git clone https://github.com/zenobi-us/opennotes.git
cd opennotes
go build -o opennotes .
sudo mv opennotes /usr/local/bin/
\`\`\`

**Verify Installation:**
\`\`\`bash
opennotes version
\`\`\`

**If already installed, ensure it's in PATH:**
\`\`\`bash
# Check current PATH
echo $PATH

# Add to PATH (bash/zsh)
export PATH="$PATH:$HOME/go/bin"
\`\`\`
`.trim(),

  [ErrorCodes.CLI_VERSION_MISMATCH]: `
**OpenNotes CLI version is incompatible.**

This extension requires OpenNotes v0.10.0 or later.

**Upgrade OpenNotes:**
\`\`\`bash
go install github.com/zenobi-us/opennotes@latest
\`\`\`

**Check current version:**
\`\`\`bash
opennotes version
\`\`\`
`.trim(),
};
```

### Context-Specific Hints

```typescript
// src/utils/hints.ts

export function getDefaultHint(code: ErrorCode): string {
  const hints: Record<ErrorCode, string> = {
    [ErrorCodes.CLI_NOT_FOUND]: INSTALLATION_HINTS[ErrorCodes.CLI_NOT_FOUND],
    [ErrorCodes.CLI_VERSION_MISMATCH]: INSTALLATION_HINTS[ErrorCodes.CLI_VERSION_MISMATCH],
    [ErrorCodes.CLI_PERMISSION_DENIED]: 
      "Check file permissions on the opennotes binary. Try: chmod +x $(which opennotes)",
    
    [ErrorCodes.NOTEBOOK_NOT_FOUND]: 
      "No notebook found in current directory or ancestors.\n" +
      "Either:\n" +
      "1. Navigate to a directory containing .opennotes.json\n" +
      "2. Specify notebook path: { notebook: '/path/to/notebook' }\n" +
      "3. Create a notebook: opennotes notebook create 'My Notes'",
    
    [ErrorCodes.NOTEBOOK_INVALID_PATH]:
      "The specified notebook path does not exist or is not a valid notebook.\n" +
      "A valid notebook contains a .opennotes.json config file.",
    
    [ErrorCodes.NOTEBOOK_CONFIG_ERROR]:
      "The notebook's .opennotes.json file is invalid.\n" +
      "Check for JSON syntax errors or missing required fields.",
    
    [ErrorCodes.INVALID_SQL]:
      "Only SELECT and WITH queries are allowed (read-only).\n" +
      "Example: SELECT * FROM read_markdown('**/*.md') LIMIT 10\n" +
      "Docs: https://github.com/zenobi-us/opennotes/blob/main/docs/sql-guide.md",
    
    [ErrorCodes.QUERY_TIMEOUT]:
      "Query exceeded 30-second timeout.\n" +
      "Simplify your query or add LIMIT to reduce results.",
    
    [ErrorCodes.QUERY_SECURITY]:
      "Path traversal (../) is not allowed in queries.\n" +
      "Use paths relative to the notebook root.",
    
    [ErrorCodes.NOTE_NOT_FOUND]:
      "The specified note does not exist.\n" +
      "Use opennotes_list to see available notes.\n" +
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
      "Check available templates in notebook's .opennotes.json file.\n" +
      "Templates are defined in the 'templates' section.",
    
    [ErrorCodes.VIEW_NOT_FOUND]:
      "The specified view does not exist.\n" +
      "Use opennotes_views (without arguments) to list available views.\n" +
      "Built-in views: today, recent, kanban, untagged, orphans, broken-links",
    
    [ErrorCodes.VIEW_INVALID_PARAMS]:
      "Invalid parameters for this view.\n" +
      "Use opennotes_views to see view parameter requirements.",
    
    [ErrorCodes.VIEW_EXECUTE_FAILED]:
      "View execution failed.\n" +
      "The view's SQL query may be invalid or target non-existent data.",
    
    [ErrorCodes.PARSE_ERROR]:
      "Failed to parse CLI output.\n" +
      "This may indicate a version mismatch. Try: opennotes version",
    
    [ErrorCodes.ABORTED]:
      "Operation was cancelled.\n" +
      "This is normal if you interrupted the operation.",
    
    [ErrorCodes.UNKNOWN_ERROR]:
      "An unexpected error occurred.\n" +
      "Check the error details for more information.\n" +
      "If this persists, please report at: https://github.com/zenobi-us/opennotes/issues",
  };

  return hints[code] ?? hints[ErrorCodes.UNKNOWN_ERROR];
}
```

---

## Error Handling Patterns

### Service-Level Error Handling

```typescript
// Pattern: Services throw OpenNotesError

export class NoteService implements INoteService {
  async getNote(path: string, options: GetOptions): Promise<GetResult> {
    // Validate path first
    if (!path.endsWith(".md")) {
      throw new OpenNotesError(
        `Invalid note path: ${path}`,
        ErrorCodes.NOTE_INVALID_PATH,
        { path },
        "Note paths must end with .md"
      );
    }

    const result = await this.cli.exec(
      "opennotes",
      ["notes", "search", "--sql", this.buildGetQuery(path)],
      { notebook: options.notebook, signal: options.signal }
    );

    if (result.code !== 0) {
      if (result.stderr.includes("no rows")) {
        throw new OpenNotesError(
          `Note not found: ${path}`,
          ErrorCodes.NOTE_NOT_FOUND,
          { path, notebook: options.notebook }
        );
      }
      throw new OpenNotesError(
        `Failed to get note: ${result.stderr}`,
        ErrorCodes.UNKNOWN_ERROR,
        { path, stderr: result.stderr }
      );
    }

    // ... rest of implementation
  }
}
```

### Tool-Level Error Handling

```typescript
// Pattern: Tools catch and format errors

export function createGetTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}get`,
    // ...schema...
    
    async execute(toolCallId, params, onUpdate, ctx, signal) {
      try {
        // Check CLI installation first
        const installation = await services.cli.checkInstallation();
        if (!installation.installed) {
          throw new OpenNotesError(
            "OpenNotes CLI not found",
            ErrorCodes.CLI_NOT_FOUND,
            { searchedPaths: process.env.PATH?.split(":") }
          );
        }

        // Execute service method
        const result = await services.note.getNote(params.path, {
          notebook: params.notebook,
          includeContent: params.includeContent,
          signal,
        });

        return {
          content: [{ type: "text", text: formatNoteResult(result) }],
        };
        
      } catch (error) {
        // Convert all errors to OpenNotesError format
        if (error instanceof OpenNotesError) {
          return error.toToolResult();
        }
        
        // Wrap unexpected errors
        const wrapped = new OpenNotesError(
          error instanceof Error ? error.message : String(error),
          ErrorCodes.UNKNOWN_ERROR,
          { originalError: String(error) }
        );
        return wrapped.toToolResult();
      }
    },
  };
}
```

### Validation Layer

```typescript
// src/utils/validation.ts

import { OpenNotesError, ErrorCodes } from "./errors";

export function validateNotebookPath(path: string): void {
  if (!path) {
    throw new OpenNotesError(
      "Notebook path is required",
      ErrorCodes.NOTEBOOK_INVALID_PATH,
      { path }
    );
  }
  
  if (path.includes("..")) {
    throw new OpenNotesError(
      "Path traversal not allowed in notebook path",
      ErrorCodes.QUERY_SECURITY,
      { path }
    );
  }
}

export function validateNotePath(path: string): void {
  if (!path) {
    throw new OpenNotesError(
      "Note path is required",
      ErrorCodes.NOTE_INVALID_PATH,
      { path }
    );
  }
  
  if (!path.endsWith(".md")) {
    throw new OpenNotesError(
      `Note path must end with .md: ${path}`,
      ErrorCodes.NOTE_INVALID_PATH,
      { path }
    );
  }
  
  if (path.includes("..")) {
    throw new OpenNotesError(
      "Path traversal not allowed",
      ErrorCodes.QUERY_SECURITY,
      { path }
    );
  }
}

export function validateSql(sql: string): void {
  const trimmed = sql.trim().toLowerCase();
  
  if (!trimmed.startsWith("select") && !trimmed.startsWith("with")) {
    throw new OpenNotesError(
      "Only SELECT and WITH queries are allowed",
      ErrorCodes.INVALID_SQL,
      { sql: sql.slice(0, 100) }
    );
  }
  
  // Check for dangerous patterns
  const dangerous = ["insert", "update", "delete", "drop", "create", "alter"];
  for (const keyword of dangerous) {
    if (trimmed.includes(keyword)) {
      throw new OpenNotesError(
        `Dangerous keyword detected: ${keyword}`,
        ErrorCodes.QUERY_SECURITY,
        { sql: sql.slice(0, 100), keyword }
      );
    }
  }
}
```

---

## Error Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Error Flow                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User Request                                               │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────┐                                            │
│  │  Tool       │                                            │
│  │  Validation │ ──── ValidationError ────┐                 │
│  └──────┬──────┘                          │                 │
│         │                                 │                 │
│         ▼                                 │                 │
│  ┌─────────────┐                          │                 │
│  │  CLI Check  │ ──── InstallationError ──┤                 │
│  └──────┬──────┘                          │                 │
│         │                                 │                 │
│         ▼                                 │                 │
│  ┌─────────────┐                          │                 │
│  │  Service    │ ──── ServiceError ───────┤                 │
│  │  Method     │                          │                 │
│  └──────┬──────┘                          │                 │
│         │                                 │                 │
│         ▼                                 ▼                 │
│  ┌─────────────┐                   ┌─────────────┐         │
│  │  CLI Exec   │ ────────────────▶ │ OpenNotes   │         │
│  └──────┬──────┘                   │ Error       │         │
│         │                          └──────┬──────┘         │
│         │                                 │                 │
│         ▼                                 ▼                 │
│  ┌─────────────┐                   ┌─────────────┐         │
│  │  Success    │                   │ Error       │         │
│  │  Response   │                   │ Response    │         │
│  └─────────────┘                   │ + Hint      │         │
│                                    └─────────────┘         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Example Error Outputs

### CLI Not Found

```json
{
  "error": true,
  "message": "OpenNotes CLI not found",
  "code": "OPENNOTES_CLI_NOT_FOUND",
  "hint": "**OpenNotes CLI is not installed or not in PATH.**\n\nInstall OpenNotes using one of these methods:\n\n**Option 1: Go Install (Recommended)**\n```bash\ngo install github.com/zenobi-us/opennotes@latest\n```\n...",
  "details": {
    "searchedPaths": ["/usr/local/bin", "/usr/bin", "/home/user/go/bin"]
  },
  "recoverable": true
}
```

### Note Not Found

```json
{
  "error": true,
  "message": "Note not found: tasks/nonexistent.md",
  "code": "OPENNOTES_NOTE_NOT_FOUND",
  "hint": "The specified note does not exist.\nUse opennotes_list to see available notes.\nEnsure path is relative to notebook root (e.g., 'notes/my-note.md').",
  "details": {
    "path": "tasks/nonexistent.md",
    "notebook": "/home/user/notes"
  },
  "recoverable": true
}
```

### Invalid SQL

```json
{
  "error": true,
  "message": "Only SELECT and WITH queries are allowed",
  "code": "OPENNOTES_INVALID_SQL",
  "hint": "Only SELECT and WITH queries are allowed (read-only).\nExample: SELECT * FROM read_markdown('**/*.md') LIMIT 10\nDocs: https://github.com/zenobi-us/opennotes/blob/main/docs/sql-guide.md",
  "details": {
    "sql": "DELETE FROM notes WHERE ...",
    "keyword": "delete"
  },
  "recoverable": false
}
```

---

## Expected Outcome

✅ Error handling strategy designed with installation hints.

## Actual Outcome

Comprehensive error handling design with:
- Error classification system (6 categories, 15+ codes)
- OpenNotesError class with hint support
- Full installation guide for CLI_NOT_FOUND
- Context-specific hints for all error codes
- Service and tool error handling patterns
- Validation layer design
- Example error outputs

## Lessons Learned

1. Installation errors need full, copy-pasteable commands
2. `recoverable` flag helps LLM decide when to ask user for help
3. Validation should happen early (before CLI calls)
4. Every error code needs a default hint - no orphan codes
