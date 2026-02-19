/**
 * Integration tests for views tool
 */

import { describe, it, expect, beforeEach, mock, spyOn } from "bun:test";
import { createViewsTool } from "../../src/tools/views.tool";
import { createMockServices, FIXTURES } from "../fixtures/mocks";
import { JotError, ErrorCodes } from "../../src/utils/errors";
import type { Services } from "../../src/services";

describe("jot_views tool", () => {
  let services: Services;
  let tool: ReturnType<typeof createViewsTool>;

  beforeEach(() => {
    services = createMockServices();
    tool = createViewsTool(services, { toolPrefix: "jot_" });
  });

  describe("tool metadata", () => {
    it("has correct name", () => {
      expect(tool.name).toBe("jot_views");
    });

    it("describes dual-mode behavior", () => {
      expect(tool.description).toContain("List");
      expect(tool.description).toContain("Execute");
    });
  });

  describe("list mode (no view parameter)", () => {
    it("lists views when called without view param", async () => {
      const listSpy = spyOn(services.views, "listViews");

      await tool.execute(
        "test-id",
        {},
        () => {},
        {} as any,
        null
      );

      expect(listSpy).toHaveBeenCalled();
    });

    it("returns view definitions", async () => {
      services.views.listViews = mock(async () => ({
        views: FIXTURES.views,
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      }));

      const result = await tool.execute(
        "test-id",
        {},
        () => {},
        {} as any,
        null
      );

      expect(result.isError).toBeFalsy();
      const text = result.content[0].text;
      expect(text).toContain("today");
      expect(text).toContain("recent");
      expect(text).toContain("kanban");
    });

    it("groups views by origin", async () => {
      services.views.listViews = mock(async () => ({
        views: [
          { name: "today", origin: "built-in" as const },
          { name: "custom", origin: "notebook" as const },
        ],
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      }));

      const result = await tool.execute(
        "test-id",
        {},
        () => {},
        {} as any,
        null
      );

      const text = result.content[0].text;
      expect(text).toContain("Built-in");
      expect(text).toContain("Notebook");
    });
  });

  describe("execute mode (with view parameter)", () => {
    it("executes view when view param provided", async () => {
      const execSpy = spyOn(services.views, "executeView");

      await tool.execute(
        "test-id",
        { view: "kanban" },
        () => {},
        {} as any,
        null
      );

      expect(execSpy).toHaveBeenCalledWith(
        "kanban",
        expect.objectContaining({})
      );
    });

    it("passes params to view execution", async () => {
      const execSpy = spyOn(services.views, "executeView");

      await tool.execute(
        "test-id",
        {
          view: "kanban",
          params: { status: "todo,done" },
        },
        () => {},
        {} as any,
        null
      );

      expect(execSpy).toHaveBeenCalledWith(
        "kanban",
        expect.objectContaining({
          params: { status: "todo,done" },
        })
      );
    });

    it("passes limit and offset", async () => {
      const execSpy = spyOn(services.views, "executeView");

      await tool.execute(
        "test-id",
        { view: "recent", limit: 10, offset: 5 },
        () => {},
        {} as any,
        null
      );

      expect(execSpy).toHaveBeenCalledWith(
        "recent",
        expect.objectContaining({
          limit: 10,
          offset: 5,
        })
      );
    });

    it("returns VIEW_NOT_FOUND for missing view", async () => {
      services.views.executeView = mock(async () => {
        throw new JotError(
          "View not found: nonexistent",
          ErrorCodes.VIEW_NOT_FOUND
        );
      });

      const result = await tool.execute(
        "test-id",
        { view: "nonexistent" },
        () => {},
        {} as any,
        null
      );

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("not found");
    });

    it("returns formatted view results", async () => {
      services.views.executeView = mock(async () => ({
        view: { name: "recent", description: "Recent notes" },
        results: [
          { path: "note1.md", modified: "2026-01-28" },
          { path: "note2.md", modified: "2026-01-27" },
        ],
        pagination: { total: 2, returned: 2, page: 1, pageSize: 50, hasMore: false },
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      }));

      const result = await tool.execute(
        "test-id",
        { view: "recent" },
        () => {},
        {} as any,
        null
      );

      expect(result.isError).toBeFalsy();
      const text = result.content[0].text;
      expect(text).toContain("View: recent");
      expect(text).toContain("note1.md");
    });
  });

  describe("notebook parameter", () => {
    it("passes notebook to listViews", async () => {
      const listSpy = spyOn(services.views, "listViews");

      await tool.execute(
        "test-id",
        { notebook: "/my/notebook" },
        () => {},
        {} as any,
        null
      );

      expect(listSpy).toHaveBeenCalledWith("/my/notebook");
    });

    it("passes notebook to executeView", async () => {
      const execSpy = spyOn(services.views, "executeView");

      await tool.execute(
        "test-id",
        { view: "today", notebook: "/my/notebook" },
        () => {},
        {} as any,
        null
      );

      expect(execSpy).toHaveBeenCalledWith(
        "today",
        expect.objectContaining({ notebook: "/my/notebook" })
      );
    });
  });
});
