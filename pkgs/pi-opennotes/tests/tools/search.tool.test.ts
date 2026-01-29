/**
 * Integration tests for search tool
 */

import { describe, it, expect, beforeEach, mock, spyOn } from "bun:test";
import { createSearchTool } from "../../src/tools/search.tool";
import { createMockServices, FIXTURES } from "../fixtures/mocks";
import type { Services } from "../../src/services";

describe("opennotes_search tool", () => {
  let services: Services;
  let tool: ReturnType<typeof createSearchTool>;

  beforeEach(() => {
    services = createMockServices();
    tool = createSearchTool(services, { toolPrefix: "opennotes_" });
  });

  describe("tool metadata", () => {
    it("has correct name with prefix", () => {
      expect(tool.name).toBe("opennotes_search");
    });

    it("has descriptive label", () => {
      expect(tool.label).toBe("Search Notes");
    });

    it("has LLM-friendly description", () => {
      expect(tool.description).toContain("SQL");
      expect(tool.description).toContain("Text Search");
      expect(tool.description).toContain("Fuzzy Search");
    });

    it("supports custom prefix", () => {
      const customTool = createSearchTool(services, { toolPrefix: "notes_" });
      expect(customTool.name).toBe("notes_search");
    });
  });

  describe("CLI installation check", () => {
    it("returns error when CLI not installed", async () => {
      services.cli.checkInstallation = mock(async () => ({
        installed: false,
      }));

      const result = await tool.execute(
        "test-id",
        { query: "test" },
        () => {},
        {} as any,
        null
      );

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("not found");
      expect(result.content[0].text).toContain("install");
    });
  });

  describe("text search", () => {
    it("executes text search with query", async () => {
      const textSearchSpy = spyOn(services.search, "textSearch");

      await tool.execute(
        "test-id",
        { query: "meeting" },
        () => {},
        {} as any,
        null
      );

      expect(textSearchSpy).toHaveBeenCalledWith(
        "meeting",
        expect.objectContaining({})
      );
    });

    it("passes limit and offset to service", async () => {
      const textSearchSpy = spyOn(services.search, "textSearch");

      await tool.execute(
        "test-id",
        { query: "test", limit: 25, offset: 50 },
        () => {},
        {} as any,
        null
      );

      expect(textSearchSpy).toHaveBeenCalledWith(
        "test",
        expect.objectContaining({ limit: 25, offset: 50 })
      );
    });
  });

  describe("fuzzy search", () => {
    it("executes fuzzy search when fuzzy=true", async () => {
      const fuzzySpy = spyOn(services.search, "fuzzySearch");

      await tool.execute(
        "test-id",
        { query: "meetng", fuzzy: true },
        () => {},
        {} as any,
        null
      );

      expect(fuzzySpy).toHaveBeenCalledWith(
        "meetng",
        expect.objectContaining({})
      );
    });
  });

  describe("SQL search", () => {
    it("executes SQL search when sql provided", async () => {
      const sqlSpy = spyOn(services.search, "sqlSearch");
      const sql = "SELECT * FROM read_markdown('**/*.md')";

      await tool.execute(
        "test-id",
        { sql },
        () => {},
        {} as any,
        null
      );

      expect(sqlSpy).toHaveBeenCalledWith(
        sql,
        expect.objectContaining({})
      );
    });

    it("prefers SQL over query when both provided", async () => {
      const sqlSpy = spyOn(services.search, "sqlSearch");
      const textSpy = spyOn(services.search, "textSearch");

      await tool.execute(
        "test-id",
        { query: "test", sql: "SELECT 1" },
        () => {},
        {} as any,
        null
      );

      expect(sqlSpy).toHaveBeenCalled();
      // textSearch should NOT be called
      expect(textSpy.mock.calls?.length ?? 0).toBe(0);
    });
  });

  describe("boolean search", () => {
    it("executes boolean search when filters provided", async () => {
      const boolSpy = spyOn(services.search, "booleanSearch");

      await tool.execute(
        "test-id",
        { filters: { and: ["data.tag=epic"] } },
        () => {},
        {} as any,
        null
      );

      expect(boolSpy).toHaveBeenCalledWith(
        { and: ["data.tag=epic"] },
        expect.objectContaining({})
      );
    });
  });

  describe("response format", () => {
    it("returns formatted results", async () => {
      services.search.textSearch = mock(async () => ({
        results: FIXTURES.notes,
        query: { type: "text" as const, executed: "..." },
        pagination: { total: 3, returned: 3, page: 1, pageSize: 50, hasMore: false },
      }));

      const result = await tool.execute(
        "test-id",
        { query: "test" },
        () => {},
        {} as any,
        null
      );

      expect(result.isError).toBeFalsy();
      const text = result.content[0].text;
      expect(text).toContain("Search Results");
      expect(text).toContain("Project Alpha");
    });
  });

  describe("error handling", () => {
    it("wraps service errors", async () => {
      services.search.textSearch = mock(async () => {
        throw new Error("Database connection failed");
      });

      const result = await tool.execute(
        "test-id",
        { query: "test" },
        () => {},
        {} as any,
        null
      );

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("Error");
    });
  });
});
