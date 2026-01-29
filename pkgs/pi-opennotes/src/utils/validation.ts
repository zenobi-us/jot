/**
 * Input validation utilities for pi-opennotes
 * Validates user inputs before passing to services
 */

import { OpenNotesError, ErrorCodes } from "./errors";

/**
 * Validate notebook path
 */
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

/**
 * Validate note path
 */
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

/**
 * Validate SQL query for safety
 */
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
  const dangerous = ["insert", "update", "delete", "drop", "create", "alter", "truncate"];
  for (const keyword of dangerous) {
    // Use word boundary check to avoid false positives
    const regex = new RegExp(`\\b${keyword}\\b`, "i");
    if (regex.test(sql)) {
      throw new OpenNotesError(
        `Dangerous keyword detected: ${keyword}`,
        ErrorCodes.QUERY_SECURITY,
        { sql: sql.slice(0, 100), keyword }
      );
    }
  }
}

/**
 * Validate view name
 */
export function validateViewName(name: string): void {
  if (!name) {
    throw new OpenNotesError(
      "View name is required",
      ErrorCodes.VIEW_INVALID_PARAMS,
      { name }
    );
  }

  // View names should be alphanumeric with dashes/underscores
  if (!/^[a-zA-Z0-9_-]+$/.test(name)) {
    throw new OpenNotesError(
      `Invalid view name: ${name}. Use only alphanumeric characters, dashes, and underscores.`,
      ErrorCodes.VIEW_INVALID_PARAMS,
      { name }
    );
  }
}

/**
 * Validate note title for creation
 */
export function validateNoteTitle(title: string): void {
  if (!title || title.trim().length === 0) {
    throw new OpenNotesError(
      "Note title is required",
      ErrorCodes.NOTE_CREATE_FAILED,
      { title }
    );
  }

  if (title.length > 200) {
    throw new OpenNotesError(
      "Note title is too long (max 200 characters)",
      ErrorCodes.NOTE_CREATE_FAILED,
      { title, length: title.length }
    );
  }
}

/**
 * Sanitize string for SQL LIKE queries
 */
export function escapeSqlString(value: string): string {
  return value.replace(/'/g, "''");
}

/**
 * Validate pagination parameters
 */
export function validatePagination(limit?: number, offset?: number): void {
  if (limit !== undefined) {
    if (limit < 1) {
      throw new OpenNotesError(
        "Limit must be at least 1",
        ErrorCodes.INVALID_SQL,
        { limit }
      );
    }
    if (limit > 1000) {
      throw new OpenNotesError(
        "Limit cannot exceed 1000",
        ErrorCodes.INVALID_SQL,
        { limit }
      );
    }
  }

  if (offset !== undefined && offset < 0) {
    throw new OpenNotesError(
      "Offset cannot be negative",
      ErrorCodes.INVALID_SQL,
      { offset }
    );
  }
}
