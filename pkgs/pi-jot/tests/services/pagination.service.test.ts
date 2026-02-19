/**
 * Unit tests for PaginationService
 */

import { describe, it, expect } from "bun:test";
import { PaginationService } from "../../src/services/pagination.service";

describe("PaginationService", () => {
  const service = new PaginationService({
    defaultPageSize: 50,
    maxOutputBytes: 1000, // Small for testing
    maxOutputLines: 100,
    budgetRatio: 0.75,
  });

  describe("paginate", () => {
    it("returns all items when under limit", () => {
      const result = service.paginate({
        items: [1, 2, 3],
        total: 3,
        limit: 50,
        offset: 0,
      });

      expect(result.items).toEqual([1, 2, 3]);
      expect(result.pagination).toEqual({
        total: 3,
        returned: 3,
        page: 1,
        pageSize: 50,
        hasMore: false,
      });
    });

    it("calculates correct page number for first page", () => {
      const result = service.paginate({
        items: [1, 2, 3],
        total: 100,
        limit: 50,
        offset: 0,
      });

      expect(result.pagination.page).toBe(1);
    });

    it("calculates correct page number for second page", () => {
      const result = service.paginate({
        items: [51, 52, 53],
        total: 100,
        limit: 50,
        offset: 50,
      });

      expect(result.pagination.page).toBe(2);
      expect(result.pagination.hasMore).toBe(true);
    });

    it("includes nextOffset when more results exist", () => {
      const result = service.paginate({
        items: Array(50).fill(0),
        total: 127,
        limit: 50,
        offset: 0,
      });

      expect(result.pagination.hasMore).toBe(true);
      expect(result.pagination.nextOffset).toBe(50);
    });

    it("does not include nextOffset when no more results", () => {
      const result = service.paginate({
        items: [1, 2, 3],
        total: 3,
        limit: 50,
        offset: 0,
      });

      expect(result.pagination.hasMore).toBe(false);
      expect(result.pagination.nextOffset).toBeUndefined();
    });

    it("handles empty items", () => {
      const result = service.paginate({
        items: [],
        total: 0,
        limit: 50,
        offset: 0,
      });

      expect(result.items).toEqual([]);
      expect(result.pagination.total).toBe(0);
      expect(result.pagination.returned).toBe(0);
    });
  });

  describe("exceedsBudget", () => {
    it("returns false for small content", () => {
      const content = "Small content";
      expect(service.exceedsBudget(content)).toBe(false);
    });

    it("returns true when exceeds byte budget", () => {
      const content = "x".repeat(1000); // 1000 bytes, budget is 750 (75% of 1000)
      expect(service.exceedsBudget(content)).toBe(true);
    });

    it("respects custom budget ratio", () => {
      const content = "x".repeat(500); // 500 bytes
      expect(service.exceedsBudget(content, 0.4)).toBe(true); // 40% of 1000 = 400
      expect(service.exceedsBudget(content, 0.6)).toBe(false); // 60% of 1000 = 600
    });
  });

  describe("fitToBudget", () => {
    it("returns all items when under budget", () => {
      const items = ["a", "b", "c"];

      const result = service.fitToBudget(items, (item) => item, 0.75);

      expect(result.items).toEqual(items);
      expect(result.truncated).toBe(false);
      expect(result.originalCount).toBe(3);
    });

    it("truncates items to fit byte budget", () => {
      // Each item is ~50 bytes when serialized as JSON
      const items = Array(100).fill("This is a test string for budget");

      const result = service.fitToBudget(
        items,
        (item) => JSON.stringify(item),
        0.75
      );

      expect(result.items.length).toBeLessThan(100);
      expect(result.truncated).toBe(true);
      expect(result.originalCount).toBe(100);
    });

    it("returns empty array for items that exceed budget individually", () => {
      const items = ["x".repeat(2000)]; // Single item exceeds budget

      const result = service.fitToBudget(items, (item) => item, 0.75);

      expect(result.items).toEqual([]);
      expect(result.truncated).toBe(true);
    });
  });

  describe("getDefaultPageSize", () => {
    it("returns configured page size", () => {
      expect(service.getDefaultPageSize()).toBe(50);
    });
  });

  describe("getBudgetLimits", () => {
    it("returns configured limits", () => {
      const limits = service.getBudgetLimits();
      expect(limits.maxBytes).toBe(1000);
      expect(limits.maxLines).toBe(100);
      expect(limits.ratio).toBe(0.75);
    });
  });
});
