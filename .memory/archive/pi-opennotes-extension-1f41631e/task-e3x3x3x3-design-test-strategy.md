---
id: e3x3x3x3
title: Design Test Strategy
created_at: 2026-01-29T09:50:00+10:30
updated_at: 2026-01-29T09:50:00+10:30
status: done
epic_id: 1f41631e
phase_id: 43842f12
assigned_to: null
---

# Design Test Strategy

## Objective

Design comprehensive unit and integration test approach for pi-opennotes services, ensuring high coverage while maintaining test maintainability.

## Completed Steps

- [x] Define test pyramid strategy
- [x] Design unit test patterns for services
- [x] Design integration test patterns for tools
- [x] Design E2E test patterns with real CLI
- [x] Create mock patterns for CliAdapter
- [x] Define test fixtures and helpers
- [x] Document coverage targets
- [x] Create test matrix

---

## Test Pyramid Strategy

```
                        ┌───────┐
                        │  E2E  │  5-10 tests
                        │ Tests │  (real CLI, real notebooks)
                       ─┴───────┴─
                      ┌───────────┐
                      │Integration│  20-30 tests
                      │  Tests    │  (tools + mocked services)
                     ─┴───────────┴─
                    ┌───────────────┐
                    │  Unit Tests   │  50-80 tests
                    │  (Services)   │  (mocked CLI adapter)
                   ─┴───────────────┴─

Ratio Target: 70% Unit / 25% Integration / 5% E2E
```

### Test Distribution by Component

| Component | Unit Tests | Integration Tests | E2E Tests |
|-----------|-----------|-------------------|-----------|
| CliAdapter | 10 | 2 | 3 |
| SearchService | 12 | - | - |
| ListService | 8 | - | - |
| NoteService | 10 | - | - |
| NotebookService | 6 | - | - |
| ViewsService | 8 | - | - |
| PaginationService | 8 | - | - |
| search.tool | - | 5 | 1 |
| list.tool | - | 4 | 1 |
| get.tool | - | 4 | 1 |
| create.tool | - | 4 | 1 |
| notebooks.tool | - | 3 | 1 |
| views.tool | - | 5 | 1 |
| **Total** | **62** | **27** | **9** |

---

## Unit Test Patterns

### Service Unit Test Structure

```typescript
// tests/services/search.service.test.ts

import { describe, it, expect, beforeEach, mock, spyOn } from "bun:test";
import { SearchService } from "../../src/services/search.service";
import { PaginationService } from "../../src/services/pagination.service";
import { createMockCliAdapter } from "../fixtures/mocks";
import type { ICliAdapter } from "../../src/services/types";

describe("SearchService", () => {
  let service: SearchService;
  let mockCli: ICliAdapter;
  let pagination: PaginationService;

  beforeEach(() => {
    mockCli = createMockCliAdapter();
    pagination = new PaginationService({ defaultPageSize: 50, budgetRatio: 0.75 });
    service = new SearchService(mockCli, pagination);
  });

  describe("textSearch", () => {
    it("converts text query to SQL for JSON output", async () => {
      mockCli.exec = mock(() => Promise.resolve({
        code: 0,
        stdout: JSON.stringify([{ file_path: "test.md", content: "meeting notes" }]),
        stderr: "",
      }));

      await service.textSearch("meeting", { limit: 10 });

      expect(mockCli.exec).toHaveBeenCalledWith(
        "opennotes",
        ["notes", "search", "--sql", expect.stringContaining("meeting")],
        expect.any(Object)
      );
    });

    it("escapes single quotes in query", async () => {
      mockCli.exec = mock(() => Promise.resolve({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.textSearch("John's meeting", {});

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1][3]).toContain("John''s");  // Escaped
    });

    it("respects limit and offset parameters", async () => {
      mockCli.exec = mock(() => Promise.resolve({
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
      mockCli.exec = mock(() => Promise.resolve({
        code: 1,
        stdout: "",
        stderr: "database error",
      }));

      await expect(service.textSearch("test", {}))
        .rejects.toThrow("SQL query failed");
    });
  });

  describe("sqlSearch", () => {
    it("executes raw SQL query", async () => {
      const sql = "SELECT * FROM read_markdown('**/*.md') LIMIT 5";
      mockCli.exec = mock(() => Promise.resolve({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.sqlSearch(sql, {});

      expect(mockCli.exec).toHaveBeenCalledWith(
        "opennotes",
        ["notes", "search", "--sql", sql],
        expect.any(Object)
      );
    });

    it("includes notebook flag when provided", async () => {
      mockCli.exec = mock(() => Promise.resolve({
        code: 0,
        stdout: "[]",
        stderr: "",
      }));

      await service.sqlSearch("SELECT 1", { notebook: "/path/to/nb" });

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toContain("--notebook");
      expect(call[1]).toContain("/path/to/nb");
    });
  });

  describe("booleanSearch", () => {
    it("builds correct AND conditions", async () => {
      mockCli.exec = mock(() => Promise.resolve({
        code: 0,
        stdout: "",
        stderr: "",
      }));

      await service.booleanSearch({
        and: ["data.tag=meeting", "data.status=active"],
      }, {});

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toContain("--and");
      expect(call[1]).toContain("data.tag=meeting");
      expect(call[1]).toContain("data.status=active");
    });

    it("combines AND, OR, and NOT conditions", async () => {
      mockCli.exec = mock(() => Promise.resolve({
        code: 0,
        stdout: "",
        stderr: "",
      }));

      await service.booleanSearch({
        and: ["data.tag=epic"],
        or: ["data.priority=high", "data.priority=critical"],
        not: ["data.status=archived"],
      }, {});

      const call = (mockCli.exec as any).mock.calls[0];
      expect(call[1]).toEqual(expect.arrayContaining([
        "--and", "data.tag=epic",
        "--or", "data.priority=high",
        "--or", "data.priority=critical",
        "--not", "data.status=archived",
      ]));
    });
  });
});
```

### PaginationService Unit Tests

```typescript
// tests/services/pagination.service.test.ts

import { describe, it, expect } from "bun:test";
import { PaginationService } from "../../src/services/pagination.service";

describe("PaginationService", () => {
  const service = new PaginationService({
    defaultPageSize: 50,
    maxOutputBytes: 1000,  // Small for testing
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

    it("calculates correct page number", () => {
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

      expect(result.pagination.nextOffset).toBe(50);
    });
  });

  describe("fitToBudget", () => {
    it("truncates items to fit byte budget", () => {
      const items = Array(100).fill("This is a long string that takes up space");
      
      const result = service.fitToBudget(
        items,
        (item) => JSON.stringify(item),
        0.75
      );

      expect(result.items.length).toBeLessThan(100);
      expect(result.truncated).toBe(true);
      expect(result.originalCount).toBe(100);
    });

    it("returns all items when under budget", () => {
      const items = ["a", "b", "c"];
      
      const result = service.fitToBudget(
        items,
        (item) => item,
        0.75
      );

      expect(result.items).toEqual(items);
      expect(result.truncated).toBe(false);
    });
  });
});
```

---

## Integration Test Patterns

### Tool Integration Tests

```typescript
// tests/tools/search.tool.test.ts

import { describe, it, expect, beforeEach } from "bun:test";
import { createSearchTool } from "../../src/tools/search.tool";
import { createMockServices } from "../fixtures/mocks";
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
      expect(tool.label).toContain("Search");
    });

    it("has LLM-friendly description", () => {
      expect(tool.description).toContain("SQL");
      expect(tool.description).toContain("text");
    });
  });

  describe("parameter validation", () => {
    it("accepts query parameter", async () => {
      const result = await tool.execute("id", { query: "meeting" }, () => {}, {}, null);
      expect(result.isError).toBeFalsy();
    });

    it("accepts sql parameter", async () => {
      const result = await tool.execute("id", { sql: "SELECT 1" }, () => {}, {}, null);
      expect(result.isError).toBeFalsy();
    });

    it("accepts filters parameter", async () => {
      const result = await tool.execute("id", {
        filters: { and: ["data.tag=epic"] },
      }, () => {}, {}, null);
      expect(result.isError).toBeFalsy();
    });
  });

  describe("CLI integration", () => {
    it("returns CLI_NOT_FOUND error when CLI missing", async () => {
      services.cli.checkInstallation = async () => ({ installed: false });

      const result = await tool.execute("id", { query: "test" }, () => {}, {}, null);

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("not installed");
    });

    it("passes notebook parameter to service", async () => {
      const spy = spyOn(services.search, "textSearch");

      await tool.execute("id", {
        query: "test",
        notebook: "/my/notebook",
      }, () => {}, {}, null);

      expect(spy).toHaveBeenCalledWith("test", expect.objectContaining({
        notebook: "/my/notebook",
      }));
    });
  });

  describe("response format", () => {
    it("returns results with pagination metadata", async () => {
      services.search.textSearch = async () => ({
        results: [{ path: "test.md" }],
        query: { type: "text", executed: "..." },
        pagination: { total: 1, returned: 1, page: 1, pageSize: 50, hasMore: false },
      });

      const result = await tool.execute("id", { query: "test" }, () => {}, {}, null);

      const text = result.content[0].text;
      expect(text).toContain("results");
      expect(text).toContain("pagination");
    });
  });
});
```

### Views Tool Dual-Mode Tests

```typescript
// tests/tools/views.tool.test.ts

import { describe, it, expect, beforeEach, spyOn } from "bun:test";
import { createViewsTool } from "../../src/tools/views.tool";
import { createMockServices } from "../fixtures/mocks";

describe("opennotes_views tool", () => {
  let services: ReturnType<typeof createMockServices>;
  let tool: ReturnType<typeof createViewsTool>;

  beforeEach(() => {
    services = createMockServices();
    tool = createViewsTool(services, { toolPrefix: "opennotes_" });
  });

  describe("list mode (no view parameter)", () => {
    it("lists views when called without view param", async () => {
      const listSpy = spyOn(services.views, "listViews");

      await tool.execute("id", {}, () => {}, {}, null);

      expect(listSpy).toHaveBeenCalled();
    });

    it("returns view definitions", async () => {
      services.views.listViews = async () => ({
        views: [
          { name: "today", origin: "built-in", description: "Today's notes" },
          { name: "recent", origin: "built-in", description: "Recent notes" },
        ],
        notebook: { name: "Test", path: "/test", source: "explicit" },
      });

      const result = await tool.execute("id", {}, () => {}, {}, null);

      expect(result.content[0].text).toContain("today");
      expect(result.content[0].text).toContain("recent");
    });
  });

  describe("execute mode (with view parameter)", () => {
    it("executes view when view param provided", async () => {
      const execSpy = spyOn(services.views, "executeView");

      await tool.execute("id", { view: "kanban" }, () => {}, {}, null);

      expect(execSpy).toHaveBeenCalledWith("kanban", expect.any(Object));
    });

    it("passes params to view execution", async () => {
      const execSpy = spyOn(services.views, "executeView");

      await tool.execute("id", {
        view: "kanban",
        params: { status: "todo,done" },
      }, () => {}, {}, null);

      expect(execSpy).toHaveBeenCalledWith("kanban", expect.objectContaining({
        params: { status: "todo,done" },
      }));
    });

    it("returns VIEW_NOT_FOUND for missing view", async () => {
      services.views.executeView = async () => {
        throw new OpenNotesError("View not found", ErrorCodes.VIEW_NOT_FOUND);
      };

      const result = await tool.execute("id", { view: "nonexistent" }, () => {}, {}, null);

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("not found");
    });
  });
});
```

---

## E2E Test Patterns

```typescript
// tests/integration/cli-integration.test.ts

import { describe, it, expect, beforeAll, afterAll } from "bun:test";
import { CliAdapter } from "../../src/services/cli-adapter";
import { SearchService } from "../../src/services/search.service";
import { PaginationService } from "../../src/services/pagination.service";
import { exec } from "child_process";
import { promisify } from "util";
import * as fs from "fs/promises";
import * as path from "path";

const execAsync = promisify(exec);

describe("E2E: CLI Integration", () => {
  let testNotebookPath: string;
  let adapter: CliAdapter;
  let searchService: SearchService;
  let cliInstalled: boolean;

  beforeAll(async () => {
    // Check if CLI is installed
    try {
      await execAsync("opennotes version");
      cliInstalled = true;
    } catch {
      cliInstalled = false;
      console.warn("⚠️  opennotes CLI not installed - skipping E2E tests");
      return;
    }

    // Create test notebook
    testNotebookPath = path.join("/tmp", `opennotes-e2e-${Date.now()}`);
    await fs.mkdir(testNotebookPath, { recursive: true });
    
    // Initialize notebook
    await execAsync(`cd ${testNotebookPath} && opennotes notebook create "E2E Test"`);
    
    // Create test notes
    await fs.writeFile(
      path.join(testNotebookPath, "meeting.md"),
      "---\ntitle: Meeting Notes\ntags: [meeting, work]\n---\n\n# Meeting Notes\n\nDiscuss project timeline."
    );
    await fs.writeFile(
      path.join(testNotebookPath, "todo.md"),
      "---\ntitle: TODO List\ntags: [todo]\n---\n\n# TODO\n\n- [ ] Complete tests"
    );

    // Setup services
    const mockPi = {
      exec: async (cmd: string, args: string[], opts?: any) => {
        const result = await execAsync(`${cmd} ${args.join(" ")}`, {
          cwd: testNotebookPath,
        });
        return { code: 0, stdout: result.stdout, stderr: result.stderr };
      },
    } as any;

    adapter = new CliAdapter(mockPi, {
      cliPath: "opennotes",
      defaultTimeout: 30000,
    });
    
    const pagination = new PaginationService({ defaultPageSize: 50, budgetRatio: 0.75 });
    searchService = new SearchService(adapter, pagination);
  });

  afterAll(async () => {
    if (testNotebookPath) {
      await fs.rm(testNotebookPath, { recursive: true, force: true });
    }
  });

  it.skipIf(!cliInstalled)("searches notes with SQL", async () => {
    const result = await searchService.sqlSearch(
      "SELECT file_path FROM read_markdown('**/*.md')",
      { notebook: testNotebookPath }
    );

    expect(result.results.length).toBeGreaterThan(0);
    expect(result.results.some((r: any) => r.file_path?.includes("meeting"))).toBe(true);
  });

  it.skipIf(!cliInstalled)("searches notes with text query", async () => {
    const result = await searchService.textSearch("meeting", {
      notebook: testNotebookPath,
    });

    expect(result.results.length).toBeGreaterThan(0);
  });

  it.skipIf(!cliInstalled)("handles empty results gracefully", async () => {
    const result = await searchService.textSearch("nonexistent12345", {
      notebook: testNotebookPath,
    });

    expect(result.results).toEqual([]);
    expect(result.pagination.total).toBe(0);
  });
});
```

---

## Mock Fixtures

```typescript
// tests/fixtures/mocks.ts

import type { ICliAdapter, CliResult, CliOptions } from "../../src/services/types";
import type { Services } from "../../src/services";
import { mock } from "bun:test";

export function createMockCliAdapter(): ICliAdapter {
  return {
    exec: mock(async (cmd: string, args: string[], opts?: CliOptions): Promise<CliResult> => ({
      code: 0,
      stdout: "[]",
      stderr: "",
    })),
    
    checkInstallation: mock(async () => ({
      installed: true,
      version: "0.10.0",
      path: "/usr/local/bin/opennotes",
    })),
    
    parseJsonOutput: <T>(stdout: string): T => JSON.parse(stdout),
    
    buildNotebookArgs: (notebook?: string) => notebook ? ["--notebook", notebook] : [],
  };
}

export function createMockServices(): Services {
  const cli = createMockCliAdapter();
  const pagination = {
    paginate: mock(({ items, total, limit, offset }) => ({
      items,
      pagination: {
        total,
        returned: items.length,
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
        hasMore: offset + items.length < total,
      },
    })),
    fitToBudget: mock((items, serialize, ratio) => ({
      items,
      truncated: false,
      originalCount: items.length,
    })),
    exceedsBudget: mock(() => false),
  };

  return {
    cli,
    pagination,
    search: {
      textSearch: mock(async () => ({ results: [], query: { type: "text", executed: "" }, pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false } })),
      fuzzySearch: mock(async () => ({ results: [], query: { type: "fuzzy", executed: "" }, pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false } })),
      sqlSearch: mock(async () => ({ results: [], query: { type: "sql", executed: "" }, pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false } })),
      booleanSearch: mock(async () => ({ results: [], query: { type: "boolean", executed: "" }, pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false } })),
    },
    list: {
      listNotes: mock(async () => ({ notes: [], notebook: { name: "", path: "", source: "explicit" }, pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false } })),
      countNotes: mock(async () => 0),
    },
    note: {
      getNote: mock(async () => ({ note: { path: "", content: "" }, notebook: { name: "", path: "", source: "explicit" } })),
      createNote: mock(async () => ({ created: { path: "", absolutePath: "", title: "" }, notebook: { name: "", path: "", source: "explicit" } })),
      noteExists: mock(async () => false),
    },
    notebook: {
      listNotebooks: mock(async () => ({ notebooks: [], current: null })),
      getCurrentNotebook: mock(async () => null),
      validateNotebook: mock(async () => ({ valid: true })),
    },
    views: {
      listViews: mock(async () => ({ views: [], notebook: { name: "", path: "", source: "explicit" } })),
      executeView: mock(async () => ({ view: { name: "", description: "" }, results: [], pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false }, notebook: { name: "", path: "", source: "explicit" } })),
      getView: mock(async () => null),
    },
  } as unknown as Services;
}
```

---

## Test Coverage Targets

| Component | Target | Critical Paths |
|-----------|--------|----------------|
| CliAdapter | 90% | exec, checkInstallation, parseJsonOutput |
| SearchService | 85% | textSearch, sqlSearch, booleanSearch |
| ListService | 85% | listNotes, countNotes |
| NoteService | 85% | getNote, createNote |
| NotebookService | 80% | listNotebooks, getCurrentNotebook |
| ViewsService | 85% | listViews, executeView |
| PaginationService | 95% | paginate, fitToBudget |
| Tools (all) | 80% | execute, error handling |
| **Overall** | **85%** | - |

---

## Test Execution Commands

```bash
# Run all tests
bun test

# Run with coverage
bun test --coverage

# Run specific test file
bun test tests/services/search.service.test.ts

# Run tests matching pattern
bun test --filter "SearchService"

# Run E2E tests only (requires opennotes CLI)
bun test tests/integration/

# Watch mode for development
bun test --watch
```

---

## CI Integration

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: oven-sh/setup-bun@v1
        with:
          bun-version: latest

      - name: Install dependencies
        run: bun install
        working-directory: pkgs/pi-opennotes

      - name: Run unit tests
        run: bun test --coverage
        working-directory: pkgs/pi-opennotes

      - name: Install opennotes for E2E
        run: go install github.com/zenobi-us/opennotes@latest

      - name: Run E2E tests
        run: bun test tests/integration/
        working-directory: pkgs/pi-opennotes
```

---

## Expected Outcome

✅ Test strategy designed with unit, integration, and E2E patterns.

## Actual Outcome

Comprehensive test strategy with:
- Test pyramid (70/25/5 ratio)
- Unit test patterns for all services
- Integration test patterns for tools
- E2E test patterns with real CLI
- Mock fixtures for isolation
- Coverage targets per component
- CI integration configuration

## Lessons Learned

1. Mock CliAdapter for unit tests, real CLI for E2E only
2. Views tool needs explicit dual-mode testing (list vs execute)
3. E2E tests should be skippable when CLI not installed
4. Coverage targets should be highest for shared services (PaginationService)
