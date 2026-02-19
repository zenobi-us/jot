/**
 * Pagination service for pi-jot
 * Handles output budgeting and pagination metadata
 */

import type {
  IPaginationService,
  PaginationConfig,
  PaginationInput,
  PaginationResult,
  PaginationMeta,
  FitToBudgetResult,
} from "./types";

// =============================================================================
// Pagination Service Implementation
// =============================================================================

export class PaginationService implements IPaginationService {
  private readonly config: PaginationConfig;

  constructor(config: Partial<PaginationConfig> = {}) {
    this.config = {
      defaultPageSize: config.defaultPageSize ?? 50,
      maxOutputBytes: config.maxOutputBytes ?? 50 * 1024, // 50KB
      maxOutputLines: config.maxOutputLines ?? 2000,
      budgetRatio: config.budgetRatio ?? 0.75,
    };
  }

  /**
   * Apply pagination to results and generate metadata
   */
  paginate<T>(input: PaginationInput<T>): PaginationResult<T> {
    const { items, total, limit, offset } = input;

    // Calculate page number (1-indexed)
    const pageSize = limit > 0 ? limit : this.config.defaultPageSize;
    const page = Math.floor(offset / pageSize) + 1;

    // Check if more results exist
    const hasMore = offset + items.length < total;

    const pagination: PaginationMeta = {
      total,
      returned: items.length,
      page,
      pageSize,
      hasMore,
    };

    // Include nextOffset only if there are more results
    if (hasMore) {
      pagination.nextOffset = offset + items.length;
    }

    return {
      items,
      pagination,
    };
  }

  /**
   * Check if content exceeds budget
   */
  exceedsBudget(content: string, budgetRatio?: number): boolean {
    const ratio = budgetRatio ?? this.config.budgetRatio;
    const maxBytes = this.config.maxOutputBytes * ratio;
    const maxLines = this.config.maxOutputLines * ratio;

    const bytes = Buffer.byteLength(content, "utf8");
    const lines = content.split("\n").length;

    return bytes > maxBytes || lines > maxLines;
  }

  /**
   * Truncate items to fit within budget
   */
  fitToBudget<T>(
    items: T[],
    serialize: (item: T) => string,
    budgetRatio?: number
  ): FitToBudgetResult<T> {
    const ratio = budgetRatio ?? this.config.budgetRatio;
    const maxBytes = this.config.maxOutputBytes * ratio;
    const maxLines = this.config.maxOutputLines * ratio;

    const originalCount = items.length;
    const result: T[] = [];
    let totalBytes = 0;
    let totalLines = 0;
    let truncated = false;

    for (const item of items) {
      const serialized = serialize(item);
      const itemBytes = Buffer.byteLength(serialized, "utf8");
      const itemLines = serialized.split("\n").length;

      // Check if adding this item would exceed budget
      if (totalBytes + itemBytes > maxBytes || totalLines + itemLines > maxLines) {
        truncated = true;
        break;
      }

      result.push(item);
      totalBytes += itemBytes;
      totalLines += itemLines;
    }

    return {
      items: result,
      truncated,
      originalCount,
    };
  }

  /**
   * Get the configured page size
   */
  getDefaultPageSize(): number {
    return this.config.defaultPageSize;
  }

  /**
   * Get budget limits
   */
  getBudgetLimits(): { maxBytes: number; maxLines: number; ratio: number } {
    return {
      maxBytes: this.config.maxOutputBytes,
      maxLines: this.config.maxOutputLines,
      ratio: this.config.budgetRatio,
    };
  }
}
