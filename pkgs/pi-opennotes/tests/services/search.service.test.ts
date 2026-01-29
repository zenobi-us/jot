/**
 * Unit tests for SearchService
 */

import { describe, it, expect, beforeEach, mock } from "bun:test";
import { SearchService } from "../../src/services/search.service";
import { PaginationService } from "../../src/services/pagination.service";
import { OpenNotesError, ErrorCodes } from "../../src/utils/errors";
import { createMockCliAdapter } from "../fixtures/mocks";

describe("SearchService", () => {
  let service: SearchService;
  let mockCli: ReturnType<typeof createMockCliAdapter>;
  let pagination: PaginationService;

  beforeEach(() => {
    mockCli = createMockCliAdapter();
    pagination = new PaginationService({ defaultPageSize: 50, budgetRatio: 0.75 });
    service = new SearchService(mockCli, pagination);
  });

  describe("textSearch", () => {
    it("converts text query to SQL", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: JSON.stringify([{ path: "test.md" }]),
        stderr: "",
      }));

      await service.textSearch("meeting", { limit: 10 });

      expect(mockCli.exec).toHaveBeenCalled();
      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toContain("--sql");
      expect(call[1][3]).toContain("meeting");
    });

    it("escapes single quotes in query", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.textSearch("John's meeting", {});

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1][3]).toContain("John''s"); // Escaped
    });

    it("respects limit and offset parameters", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.textSearch("test", { limit: 25, offset: 50 });

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1][3]).toContain("LIMIT 25");
      expect(call[1][3]).toContain("OFFSET 50");
    });

    it("throws OpenNotesError on CLI failure", async () => {
      mockCli.exec = mock(async () => ({
        code: 1,
        stdout: "",
        stderr: "database error",
      }));

      await expect(service.textSearch("test", {})).rejects.toThrow(OpenNotesError);
    });

    it("returns results with pagination metadata", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: JSON.stringify([
          { path: "note1.md", title: "Note 1" },
          { path: "note2.md", title: "Note 2" },
        ]),
        stderr: "",
      }));

      const result = await service.textSearch("note", {});

      expect(result.results.length).toBe(2);
      expect(result.query.type).toBe("text");
      expect(result.pagination.returned).toBe(2);
    });
  });

  describe("sqlSearch", () => {
    it("executes raw SQL query", async () => {
      const sql = "SELECT * FROM read_markdown('**/*.md') LIMIT 5";
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.sqlSearch(sql, {});

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toContain("--sql");
      expect(call[1]).toContain(sql);
    });

    it("includes notebook flag when provided", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.sqlSearch("SELECT 1", { notebook: "/path/to/nb" });

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toContain("--notebook");
      expect(call[1]).toContain("/path/to/nb");
    });

    it("validates SQL (rejects non-SELECT)", async () => {
      await expect(
        service.sqlSearch("DELETE FROM notes", {})
      ).rejects.toThrow(OpenNotesError);
    });

    it("validates SQL (rejects dangerous keywords)", async () => {
      await expect(
        service.sqlSearch("SELECT * FROM notes; DROP TABLE notes", {})
      ).rejects.toThrow(OpenNotesError);
    });
  });

  describe("booleanSearch", () => {
    it("builds correct AND conditions", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.booleanSearch(
        {
          and: ["data.tag=meeting", "data.status=active"],
        },
        {}
      );

      const call = (mockCli.exec as any).mock.calls[0];
      const sql = call[1][3];
      expect(sql).toContain("metadata->>'tag'");
      expect(sql).toContain("metadata->>'status'");
      expect(sql).toContain("AND");
    });

    it("combines AND, OR, and NOT conditions", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.booleanSearch(
        {
          and: ["data.tag=epic"],
          or: ["data.priority=high", "data.priority=critical"],
          not: ["data.status=archived"],
        },
        {}
      );

      const call = (mockCli.exec as any).mock.calls[0];
      const sql = call[1][3];
      expect(sql).toContain("OR");
      expect(sql).toContain("NOT");
    });

    it("returns correct query type", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      const result = await service.booleanSearch({ and: ["data.tag=test"] }, {});

      expect(result.query.type).toBe("boolean");
    });
  });

  describe("fuzzySearch", () => {
    it("uses CLI fuzzy flag", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "• result1.md\n• result2.md",
        stderr: "",
      }));

      await service.fuzzySearch("meetng", {}); // Typo intentional

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toContain("--fuzzy");
      expect(call[1]).toContain("meetng");
    });

    it("parses text output to NoteSummary", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "• notes/meeting.md\n• tasks/task.md",
        stderr: "",
      }));

      const result = await service.fuzzySearch("meeting", {});

      expect(result.results.length).toBe(2);
      expect(result.results[0].path).toBe("notes/meeting.md");
    });

    it("returns correct query type", async () => {
      mockCli.exec = mock(async () => ({
        code: 0,
        stdout: "",
        stderr: "",
      }));

      const result = await service.fuzzySearch("test", {});

      expect(result.query.type).toBe("fuzzy");
    });
  });
});
