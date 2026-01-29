/**
 * Search service for pi-opennotes
 * Handles text, fuzzy, SQL, and boolean search operations
 */

import type {
  ISearchService,
  ICliAdapter,
  IPaginationService,
  SearchOptions,
  SearchResult,
  BooleanFilters,
  NoteSummary,
} from "./types";
import { OpenNotesError, ErrorCodes } from "../utils/errors";
import { escapeSqlString, validateSql, validatePagination } from "../utils/validation";

// =============================================================================
// Search Service Implementation
// =============================================================================

export class SearchService implements ISearchService {
  constructor(
    private readonly cli: ICliAdapter,
    private readonly pagination: IPaginationService
  ) {}

  /**
   * Text-based search (exact substring)
   */
  async textSearch(query: string, options: SearchOptions): Promise<SearchResult> {
    validatePagination(options.limit, options.offset);

    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;

    // Convert text search to SQL for consistent JSON output
    const escapedQuery = escapeSqlString(query);
    const sql = `
      SELECT 
        file_path as path,
        metadata->>'title' as title,
        metadata->'tags' as tags,
        metadata->>'created' as created,
        metadata->>'modified' as modified
      FROM read_markdown('**/*.md')
      WHERE content LIKE '%${escapedQuery}%' 
         OR metadata->>'title' LIKE '%${escapedQuery}%'
      ORDER BY file_path
      LIMIT ${limit}
      OFFSET ${offset}
    `.trim();

    return this.executeSqlInternal(sql, "text", options);
  }

  /**
   * Fuzzy search (typo-tolerant, ranked)
   */
  async fuzzySearch(query: string, options: SearchOptions): Promise<SearchResult> {
    validatePagination(options.limit, options.offset);

    // Use CLI's native fuzzy search
    const args = [
      "notes",
      "search",
      "--fuzzy",
      query,
      ...this.cli.buildNotebookArgs(options.notebook),
    ];

    const result = await this.cli.exec("opennotes", args, {
      signal: options.signal,
    });

    if (result.code !== 0) {
      throw new OpenNotesError(
        `Fuzzy search failed: ${result.stderr}`,
        ErrorCodes.SEARCH_FAILED,
        { query }
      );
    }

    // Parse glamour output (not JSON) - convert to NoteSummary[]
    const notes = this.parseTextSearchOutput(result.stdout);
    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;

    // Apply pagination
    const paginatedNotes = notes.slice(offset, offset + limit);

    const { pagination } = this.pagination.paginate({
      items: paginatedNotes,
      total: notes.length,
      limit,
      offset,
    });

    return {
      results: paginatedNotes,
      query: { type: "fuzzy", executed: `--fuzzy "${query}"` },
      pagination,
    };
  }

  /**
   * Raw SQL query execution
   */
  async sqlSearch(
    sql: string,
    options: SearchOptions
  ): Promise<SearchResult<Record<string, unknown>>> {
    validateSql(sql);
    validatePagination(options.limit, options.offset);

    return this.executeSqlInternal(sql, "sql", options);
  }

  /**
   * Boolean query (AND/OR/NOT filters)
   */
  async booleanSearch(
    filters: BooleanFilters,
    options: SearchOptions
  ): Promise<SearchResult> {
    validatePagination(options.limit, options.offset);

    // Build SQL WHERE clause from filters
    const conditions: string[] = [];
    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;

    // Process AND conditions
    if (filters.and && filters.and.length > 0) {
      for (const condition of filters.and) {
        const sqlCondition = this.parseFilterCondition(condition);
        if (sqlCondition) {
          conditions.push(sqlCondition);
        }
      }
    }

    // Process OR conditions
    const orConditions: string[] = [];
    if (filters.or && filters.or.length > 0) {
      for (const condition of filters.or) {
        const sqlCondition = this.parseFilterCondition(condition);
        if (sqlCondition) {
          orConditions.push(sqlCondition);
        }
      }
    }

    // Process NOT conditions
    if (filters.not && filters.not.length > 0) {
      for (const condition of filters.not) {
        const sqlCondition = this.parseFilterCondition(condition);
        if (sqlCondition) {
          conditions.push(`NOT (${sqlCondition})`);
        }
      }
    }

    // Combine conditions
    let whereClause = "";
    if (conditions.length > 0 || orConditions.length > 0) {
      const andPart = conditions.join(" AND ");
      const orPart =
        orConditions.length > 0 ? `(${orConditions.join(" OR ")})` : "";

      if (andPart && orPart) {
        whereClause = `WHERE ${andPart} AND ${orPart}`;
      } else if (andPart) {
        whereClause = `WHERE ${andPart}`;
      } else if (orPart) {
        whereClause = `WHERE ${orPart}`;
      }
    }

    const sql = `
      SELECT 
        file_path as path,
        metadata->>'title' as title,
        metadata->'tags' as tags,
        metadata->>'created' as created,
        metadata->>'modified' as modified
      FROM read_markdown('**/*.md')
      ${whereClause}
      ORDER BY file_path
      LIMIT ${limit}
      OFFSET ${offset}
    `.trim();

    const result = await this.executeSqlInternal<NoteSummary>(sql, "boolean", options);

    // Update the executed query to show the filter format
    result.query.executed = JSON.stringify(filters);

    return result;
  }

  /**
   * Parse a filter condition like "data.tag=meeting" to SQL
   */
  private parseFilterCondition(condition: string): string | null {
    const match = condition.match(/^([^=<>!]+)(=|!=|<|>|<=|>=|LIKE)(.*)$/i);
    if (!match) {
      return null;
    }

    const [, field, operator, value] = match;
    const trimmedField = field.trim();
    const trimmedValue = value.trim();

    // Handle data.* fields (frontmatter)
    if (trimmedField.startsWith("data.")) {
      const jsonField = trimmedField.replace("data.", "");
      const escapedValue = escapeSqlString(trimmedValue);

      if (operator.toUpperCase() === "LIKE") {
        return `metadata->>'${jsonField}' LIKE '${escapedValue}'`;
      }
      return `metadata->>'${jsonField}' ${operator} '${escapedValue}'`;
    }

    // Handle standard fields
    const fieldMap: Record<string, string> = {
      path: "file_path",
      title: "metadata->>'title'",
      content: "content",
    };

    const sqlField = fieldMap[trimmedField] ?? trimmedField;
    const escapedValue = escapeSqlString(trimmedValue);

    if (operator.toUpperCase() === "LIKE") {
      return `${sqlField} LIKE '${escapedValue}'`;
    }
    return `${sqlField} ${operator} '${escapedValue}'`;
  }

  /**
   * Execute SQL query and return structured results
   */
  private async executeSqlInternal<T = Record<string, unknown>>(
    sql: string,
    type: "text" | "fuzzy" | "sql" | "boolean",
    options: SearchOptions
  ): Promise<SearchResult<T>> {
    const args = [
      "notes",
      "search",
      "--sql",
      sql,
      ...this.cli.buildNotebookArgs(options.notebook),
    ];

    const result = await this.cli.exec("opennotes", args, {
      signal: options.signal,
    });

    if (result.code !== 0) {
      // Check for specific error types
      if (result.stderr.includes("no such file") || result.stderr.includes("not found")) {
        throw new OpenNotesError(
          `Notebook not found`,
          ErrorCodes.NOTEBOOK_NOT_FOUND,
          { notebook: options.notebook, stderr: result.stderr }
        );
      }
      throw new OpenNotesError(
        `SQL query failed: ${result.stderr}`,
        ErrorCodes.INVALID_SQL,
        { sql, stderr: result.stderr }
      );
    }

    // Parse JSON results
    const rows = this.cli.parseJsonOutput<T[]>(result.stdout);
    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;

    // Note: For actual total, we'd need a COUNT query
    // For now, use the returned count as an approximation
    const { items, pagination } = this.pagination.paginate({
      items: rows,
      total: rows.length + offset, // Approximation when at limit
      limit,
      offset,
    });

    return {
      results: items,
      query: { type, executed: sql },
      pagination,
    };
  }

  /**
   * Parse glamour-formatted text output to extract notes
   */
  private parseTextSearchOutput(stdout: string): NoteSummary[] {
    const lines = stdout.split("\n");
    const notes: NoteSummary[] = [];

    for (const line of lines) {
      // Look for file path patterns (bullet points or plain paths)
      const match = line.match(/[•\-*]\s*(\S+\.md)|^\s*(\S+\.md)/);
      if (match) {
        const path = match[1] ?? match[2];
        notes.push({ path });
      }
    }

    return notes;
  }
}
