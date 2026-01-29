/**
 * Search tool for pi-opennotes
 * Wraps SearchService to provide LLM-callable tool
 */

import type { Tool } from "@mariozechner/pi-coding-agent";
import type { Services } from "../services";
import type { ToolConfig } from "../config";
import { SearchParams, type SearchParamsType } from "../schemas/tools";
import { OpenNotesError, ErrorCodes, wrapError } from "../utils/errors";
import { formatSearchResults } from "../utils/output";

// =============================================================================
// Tool Description
// =============================================================================

const DESCRIPTION = `Search notes in an OpenNotes notebook using multiple methods:

1. **Text Search**: Set 'query' for substring matching in titles and content
2. **Fuzzy Search**: Set 'query' + 'fuzzy: true' for typo-tolerant, ranked results
3. **SQL Query**: Set 'sql' for full DuckDB SQL power
4. **Boolean Filters**: Use 'filters' for structured AND/OR/NOT queries

Common filter fields: data.tag, data.status, data.priority, path, title

Example SQL: SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%' LIMIT 10

Returns notes matching criteria with pagination metadata.`;

// =============================================================================
// Tool Factory
// =============================================================================

export function createSearchTool(services: Services, config: ToolConfig): Tool {
  return {
    name: `${config.toolPrefix}search`,
    label: "Search Notes",
    description: DESCRIPTION,
    parameters: SearchParams,

    async execute(toolCallId, params: SearchParamsType, onUpdate, ctx, signal) {
      try {
        // Check CLI installation
        const installation = await services.cli.checkInstallation();
        if (!installation.installed) {
          throw new OpenNotesError(
            "OpenNotes CLI not found",
            ErrorCodes.CLI_NOT_FOUND,
            { searchedPaths: process.env.PATH?.split(":") }
          );
        }

        // Determine search type and execute
        const options = {
          notebook: params.notebook,
          limit: params.limit,
          offset: params.offset,
          signal,
        };

        let result;

        if (params.sql) {
          // SQL query mode
          result = await services.search.sqlSearch(params.sql, options);
        } else if (params.filters) {
          // Boolean filter mode
          result = await services.search.booleanSearch(params.filters, options);
        } else if (params.query) {
          // Text or fuzzy search mode
          if (params.fuzzy) {
            result = await services.search.fuzzySearch(params.query, options);
          } else {
            result = await services.search.textSearch(params.query, options);
          }
        } else {
          // No search criteria provided - list recent notes
          result = await services.search.textSearch("", options);
        }

        // Format output for LLM
        const output = formatSearchResults(
          result.results as any[],
          result.pagination,
          result.query
        );

        return {
          content: [{ type: "text", text: output }],
        };
      } catch (error) {
        const wrapped = wrapError(error, ErrorCodes.SEARCH_FAILED);
        return wrapped.toToolResult();
      }
    },
  };
}
