/**
 * Mock fixtures for pi-jot tests
 */

import { mock } from "bun:test";
import type { ICliAdapter, CliResult, CliOptions, InstallationInfo } from "../../src/services/types";
import type { Services } from "../../src/services";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

// =============================================================================
// Mock CLI Adapter
// =============================================================================

export interface MockCliAdapterOptions {
  installed?: boolean;
  version?: string;
  defaultOutput?: string;
}

export function createMockCliAdapter(options: MockCliAdapterOptions = {}): ICliAdapter {
  const { installed = true, version = "0.10.0", defaultOutput = "[]" } = options;

  return {
    exec: mock(async (cmd: string, args: string[], opts?: CliOptions): Promise<CliResult> => ({
      code: 0,
      stdout: defaultOutput,
      stderr: "",
    })),

    checkInstallation: mock(async (): Promise<InstallationInfo> => ({
      installed,
      version: installed ? version : undefined,
      path: installed ? "/usr/local/bin/jot" : undefined,
    })),

    parseJsonOutput: <T>(stdout: string): T => {
      if (!stdout.trim()) return [] as unknown as T;
      return JSON.parse(stdout) as T;
    },

    buildNotebookArgs: (notebook?: string) => (notebook ? ["--notebook", notebook] : []),
  };
}

// =============================================================================
// Mock Services
// =============================================================================

export function createMockServices(): Services {
  const cli = createMockCliAdapter();
  const pagination = {
    paginate: mock(({ items, total, limit, offset }: { items: any[]; total: number; limit: number; offset: number }) => ({
      items,
      pagination: {
        total,
        returned: items.length,
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
        hasMore: offset + items.length < total,
        nextOffset: offset + items.length < total ? offset + items.length : undefined,
      },
    })),
    fitToBudget: mock((items: any[], serialize: (item: any) => string, ratio: number) => ({
      items,
      truncated: false,
      originalCount: items.length,
    })),
    exceedsBudget: mock(() => false),
    getDefaultPageSize: () => 50,
    getBudgetLimits: () => ({ maxBytes: 50 * 1024, maxLines: 2000, ratio: 0.75 }),
  };

  return {
    cli,
    pagination,
    search: {
      textSearch: mock(async () => ({
        results: [],
        query: { type: "text" as const, executed: "" },
        pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false },
      })),
      fuzzySearch: mock(async () => ({
        results: [],
        query: { type: "fuzzy" as const, executed: "" },
        pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false },
      })),
      sqlSearch: mock(async () => ({
        results: [],
        query: { type: "sql" as const, executed: "" },
        pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false },
      })),
      booleanSearch: mock(async () => ({
        results: [],
        query: { type: "boolean" as const, executed: "" },
        pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false },
      })),
    },
    list: {
      listNotes: mock(async () => ({
        notes: [],
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
        pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false },
      })),
      countNotes: mock(async () => 0),
    },
    note: {
      getNote: mock(async () => ({
        note: { path: "test.md", content: "# Test" },
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      })),
      createNote: mock(async () => ({
        created: { path: "new-note.md", absolutePath: "/test/new-note.md", title: "New Note" },
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      })),
      noteExists: mock(async () => false),
    },
    notebook: {
      listNotebooks: mock(async () => ({
        notebooks: [],
        current: null,
      })),
      getCurrentNotebook: mock(async () => null),
      validateNotebook: mock(async () => ({ valid: true })),
    },
    views: {
      listViews: mock(async () => ({
        views: [],
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      })),
      executeView: mock(async () => ({
        view: { name: "test", description: "Test view" },
        results: [],
        pagination: { total: 0, returned: 0, page: 1, pageSize: 50, hasMore: false },
        notebook: { name: "Test", path: "/test", source: "explicit" as const },
      })),
      getView: mock(async () => null),
    },
  } as unknown as Services;
}

// =============================================================================
// Mock Extension API
// =============================================================================

export function createMockExtensionAPI(): ExtensionAPI {
  const tools: any[] = [];
  const commands: Record<string, any> = {};
  const events: Record<string, any[]> = {};

  return {
    exec: mock(async (cmd: string, args: string[], opts?: any) => ({
      code: 0,
      stdout: "[]",
      stderr: "",
    })),
    
    registerTool: mock((tool: any) => {
      tools.push(tool);
    }),
    
    registerCommand: mock((name: string, handler: any) => {
      commands[name] = handler;
    }),
    
    on: mock((event: string, handler: any) => {
      if (!events[event]) events[event] = [];
      events[event].push(handler);
    }),
    
    getConfig: mock(() => ({})),
    
    // Test helpers
    _tools: tools,
    _commands: commands,
    _events: events,
  } as unknown as ExtensionAPI;
}

// =============================================================================
// Test Data Fixtures
// =============================================================================

export const FIXTURES = {
  notes: [
    {
      path: "project-alpha.md",
      title: "Project Alpha",
      tags: ["project", "active"],
      created: "2026-01-01T00:00:00Z",
      modified: "2026-01-28T14:30:00Z",
    },
    {
      path: "tasks/task-001.md",
      title: "Implement Feature X",
      tags: ["task", "alpha"],
      created: "2026-01-15T10:00:00Z",
      modified: "2026-01-27T09:15:00Z",
    },
    {
      path: "meetings/standup-2026-01-28.md",
      title: "Standup 2026-01-28",
      tags: ["meeting"],
      created: "2026-01-28T09:00:00Z",
      modified: "2026-01-28T09:30:00Z",
    },
  ],

  notebooks: [
    {
      name: "Work Notes",
      path: "/home/user/notes/work",
      source: "registered" as const,
      noteCount: 127,
    },
    {
      name: "Personal",
      path: "/home/user/notes/personal",
      source: "registered" as const,
      noteCount: 43,
    },
  ],

  views: [
    {
      name: "today",
      origin: "built-in" as const,
      description: "Notes modified today",
    },
    {
      name: "recent",
      origin: "built-in" as const,
      description: "Recently modified notes",
      parameters: [
        {
          name: "days",
          type: "number",
          required: false,
          default: "7",
          description: "Number of days to look back",
        },
      ],
    },
    {
      name: "kanban",
      origin: "built-in" as const,
      description: "Kanban board view",
      parameters: [
        {
          name: "status",
          type: "string",
          required: false,
          default: "todo,in-progress,done",
        },
      ],
    },
  ],
};
