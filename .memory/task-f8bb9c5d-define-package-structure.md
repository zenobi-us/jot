---
id: f8bb9c5d
title: Define Package Structure
created_at: 2026-01-28T23:30:00+10:30
updated_at: 2026-01-29T09:25:00+10:30
status: done
epic_id: 1f41631e
phase_id: 43842f12
assigned_to: null
---

# Define Package Structure

## Objective

Design the final directory structure for the `pkgs/pi-opennotes/` package, including service layer architecture, configuration files, and module organization.

## Completed Steps

- [x] Design directory structure with service layer
- [x] Create package.json template
- [x] Create tsconfig.json template
- [x] Define module organization (services/, tools/, types/)
- [x] Plan test file organization
- [x] Document build and publish scripts
- [x] Define peer dependencies correctly
- [x] Plan README structure

---

## Final Package Structure

```
pkgs/pi-opennotes/
├── package.json              # npm package + pi manifest
├── tsconfig.json             # TypeScript config
├── bun.lockb                 # Bun lockfile
├── README.md                 # Usage documentation
├── CHANGELOG.md              # Version history
├── LICENSE                   # MIT License
│
├── src/
│   ├── index.ts              # Extension entry point (thin)
│   ├── config.ts             # Extension configuration
│   │
│   ├── services/             # Business logic layer
│   │   ├── index.ts          # Service exports
│   │   ├── types.ts          # Service interfaces
│   │   ├── cli-adapter.ts    # CLI execution adapter
│   │   ├── search.service.ts # Search operations
│   │   ├── list.service.ts   # List operations
│   │   ├── note.service.ts   # Note CRUD operations
│   │   ├── notebook.service.ts # Notebook operations
│   │   ├── views.service.ts  # Views operations
│   │   └── pagination.service.ts # Pagination logic
│   │
│   ├── tools/                # Thin tool wrappers
│   │   ├── index.ts          # Tool registration
│   │   ├── search.tool.ts    # opennotes_search
│   │   ├── list.tool.ts      # opennotes_list
│   │   ├── get.tool.ts       # opennotes_get
│   │   ├── create.tool.ts    # opennotes_create
│   │   ├── notebooks.tool.ts # opennotes_notebooks
│   │   └── views.tool.ts     # opennotes_views
│   │
│   ├── schemas/              # TypeBox schemas
│   │   ├── index.ts          # Schema exports
│   │   ├── common.ts         # Shared types (pagination, etc.)
│   │   ├── search.ts         # Search params/response
│   │   ├── list.ts           # List params/response
│   │   ├── note.ts           # Note types
│   │   ├── notebook.ts       # Notebook types
│   │   └── views.ts          # Views types
│   │
│   └── utils/                # Shared utilities
│       ├── index.ts          # Utility exports
│       ├── errors.ts         # Error handling + hints
│       ├── output.ts         # Output formatting
│       └── truncate.ts       # Truncation helpers
│
└── tests/
    ├── setup.ts              # Test setup (mocks)
    │
    ├── services/             # Service unit tests
    │   ├── cli-adapter.test.ts
    │   ├── search.service.test.ts
    │   ├── list.service.test.ts
    │   ├── note.service.test.ts
    │   ├── notebook.service.test.ts
    │   ├── views.service.test.ts
    │   └── pagination.service.test.ts
    │
    ├── tools/                # Tool integration tests
    │   ├── search.tool.test.ts
    │   ├── list.tool.test.ts
    │   ├── get.tool.test.ts
    │   ├── create.tool.test.ts
    │   ├── notebooks.tool.test.ts
    │   └── views.tool.test.ts
    │
    └── integration/          # End-to-end tests
        ├── cli-integration.test.ts
        └── extension.test.ts
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         pi-opennotes Extension                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      Tool Layer (Thin)                           │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │   │
│  │  │ search   │ │ list     │ │ get      │ │ create   │ │ views  │ │   │
│  │  │ .tool.ts │ │ .tool.ts │ │ .tool.ts │ │ .tool.ts │ │.tool.ts│ │   │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └───┬────┘ │   │
│  └───────┼────────────┼────────────┼────────────┼───────────┼──────┘   │
│          │            │            │            │           │          │
│          ▼            ▼            ▼            ▼           ▼          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Service Layer (Fat)                           │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │   │
│  │  │ SearchService│  │ ListService  │  │ NoteService          │   │   │
│  │  │ - text       │  │ - all notes  │  │ - get                │   │   │
│  │  │ - fuzzy      │  │ - filtered   │  │ - create             │   │   │
│  │  │ - sql        │  │ - sorted     │  └──────────┬───────────┘   │   │
│  │  │ - boolean    │  └──────┬───────┘             │               │   │
│  │  └──────┬───────┘         │        ┌────────────┴───────────┐   │   │
│  │         │                 │        │                        │   │   │
│  │  ┌──────┴─────────────────┴────────┴──────┐  ┌────────────┐ │   │   │
│  │  │ NotebookService                         │  │ViewsService│ │   │   │
│  │  │ - list notebooks                        │  │ - list     │ │   │   │
│  │  │ - get current                           │  │ - execute  │ │   │   │
│  │  └──────────────────┬─────────────────────┘  └─────┬──────┘ │   │   │
│  │                     │                              │        │   │   │
│  │  ┌──────────────────┴──────────────────────────────┴──────┐ │   │   │
│  │  │              PaginationService                          │ │   │   │
│  │  │              - paginate results                         │ │   │   │
│  │  │              - add metadata                             │ │   │   │
│  │  └─────────────────────────┬──────────────────────────────┘ │   │   │
│  └────────────────────────────┼────────────────────────────────┘   │
│                               │                                    │
│                               ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     CLI Adapter Layer                            │   │
│  │  ┌──────────────────────────────────────────────────────────┐   │   │
│  │  │ CliAdapter                                                │   │   │
│  │  │ - exec(command, args, options): Promise<CliResult>        │   │   │
│  │  │ - checkInstallation(): Promise<boolean>                   │   │   │
│  │  │ - parseJsonOutput<T>(stdout): T                           │   │   │
│  │  │ - parseTextOutput(stdout): string[]                       │   │   │
│  │  └──────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                               │                                        │
└───────────────────────────────┼────────────────────────────────────────┘
                                │
                                ▼
                       ┌───────────────┐
                       │  opennotes    │
                       │  CLI binary   │
                       └───────────────┘
```

---

## Package Configuration Files

### package.json

```json
{
  "name": "@zenobi-us/pi-opennotes",
  "version": "0.1.0",
  "description": "Pi extension for OpenNotes - search and manage markdown notes with AI",
  "author": "zenobi-us",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/zenobi-us/opennotes.git",
    "directory": "pkgs/pi-opennotes"
  },
  "keywords": [
    "pi-package",
    "opennotes",
    "markdown",
    "notes",
    "ai",
    "duckdb"
  ],
  "type": "module",
  "main": "src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./services": "./src/services/index.ts",
    "./schemas": "./src/schemas/index.ts"
  },
  "files": [
    "src/",
    "README.md",
    "CHANGELOG.md",
    "LICENSE"
  ],
  "scripts": {
    "test": "bun test",
    "test:watch": "bun test --watch",
    "test:coverage": "bun test --coverage",
    "typecheck": "bunx tsc --noEmit",
    "lint": "bunx eslint src/",
    "lint:fix": "bunx eslint src/ --fix",
    "format": "bunx prettier --write src/",
    "prepublishOnly": "bun run typecheck && bun run lint && bun run test"
  },
  "peerDependencies": {
    "@mariozechner/pi-coding-agent": ">=0.50.0",
    "@sinclair/typebox": ">=0.32.0"
  },
  "devDependencies": {
    "@types/bun": "latest",
    "typescript": "^5.0.0",
    "eslint": "^8.0.0",
    "prettier": "^3.0.0"
  },
  "pi": {
    "extensions": ["./src/index.ts"],
    "config": {
      "toolPrefix": {
        "type": "string",
        "default": "opennotes_",
        "description": "Prefix for tool names (e.g., 'opennotes_search')"
      },
      "defaultPageSize": {
        "type": "number",
        "default": 50,
        "description": "Default number of results per page"
      },
      "cliPath": {
        "type": "string",
        "default": "opennotes",
        "description": "Path to opennotes CLI binary"
      },
      "cliTimeout": {
        "type": "number",
        "default": 30000,
        "description": "CLI command timeout in milliseconds"
      }
    }
  }
}
```

### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "types": ["bun-types"],
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

---

## Module Responsibilities

### Entry Point (`src/index.ts`)

**THIN**: Only wires components together.

```typescript
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { registerTools } from "./tools";
import { createServices } from "./services";
import { getConfig } from "./config";

export default function piOpennotes(pi: ExtensionAPI) {
  const config = getConfig(pi);
  const services = createServices(pi, config);
  
  registerTools(pi, services, config);
}
```

### Services (`src/services/`)

**FAT**: All business logic lives here.

Each service:
- Single responsibility
- Depends on `CliAdapter` for CLI execution
- Returns typed responses
- Handles its own error cases

### Tools (`src/tools/`)

**THIN**: Only orchestrate services.

Each tool:
- Validates parameters (via TypeBox)
- Calls appropriate service method
- Formats response for LLM
- Under 50 lines ideally

### Schemas (`src/schemas/`)

All TypeBox schemas for:
- Tool parameters
- Response types
- Shared types (pagination, errors)

### Utils (`src/utils/`)

Stateless utilities:
- Error creation with hints
- Output truncation
- Format helpers

---

## Test Organization

### Unit Tests (`tests/services/`)

Test services in isolation with mocked `CliAdapter`:

```typescript
// tests/services/search.service.test.ts
import { describe, it, expect, mock } from "bun:test";
import { SearchService } from "../../src/services/search.service";
import type { CliAdapter } from "../../src/services/cli-adapter";

describe("SearchService", () => {
  const mockAdapter: CliAdapter = {
    exec: mock(() => Promise.resolve({ code: 0, stdout: "[]", stderr: "" })),
    checkInstallation: mock(() => Promise.resolve(true)),
    parseJsonOutput: mock((s) => JSON.parse(s)),
  };

  it("executes text search", async () => {
    const service = new SearchService(mockAdapter);
    const result = await service.textSearch("meeting", {});
    
    expect(mockAdapter.exec).toHaveBeenCalledWith(
      "opennotes",
      ["notes", "search", "--sql", expect.stringContaining("meeting")],
      expect.any(Object)
    );
  });
});
```

### Integration Tests (`tests/tools/`)

Test tools with services (mocked CLI):

```typescript
// tests/tools/search.tool.test.ts
import { describe, it, expect } from "bun:test";
import { createSearchTool } from "../../src/tools/search.tool";
import { createMockServices } from "../setup";

describe("opennotes_search tool", () => {
  const { services } = createMockServices();
  const tool = createSearchTool(services, { toolPrefix: "opennotes_" });

  it("has correct name and description", () => {
    expect(tool.name).toBe("opennotes_search");
    expect(tool.description).toContain("Search notes");
  });

  it("executes search with query parameter", async () => {
    const result = await tool.execute("test-id", { query: "meeting" }, () => {}, {}, null);
    expect(result.content[0].text).toContain("results");
  });
});
```

### E2E Tests (`tests/integration/`)

Test with real CLI (requires opennotes installed):

```typescript
// tests/integration/cli-integration.test.ts
import { describe, it, expect, beforeAll } from "bun:test";
import { CliAdapter } from "../../src/services/cli-adapter";

describe("CLI Integration", () => {
  let adapter: CliAdapter;

  beforeAll(async () => {
    adapter = new CliAdapter({ cliPath: "opennotes" });
    const installed = await adapter.checkInstallation();
    if (!installed) {
      throw new Error("opennotes CLI not installed - skipping integration tests");
    }
  });

  it("lists notebooks", async () => {
    const result = await adapter.exec("opennotes", ["notebook", "list"]);
    expect(result.code).toBe(0);
  });
});
```

---

## Build & Publish Scripts

### Development

```bash
# Install dependencies
bun install

# Run tests
bun test

# Type check
bun run typecheck

# Lint
bun run lint
```

### Publishing

```bash
# Ensure quality
bun run prepublishOnly

# Publish to npm
npm publish --access public

# Or for beta
npm publish --tag beta --access public
```

---

## README Structure

1. **Overview** - What is pi-opennotes
2. **Installation** - npm install + prerequisites
3. **Quick Start** - Enable extension, basic usage
4. **Tools Reference** - Each tool with examples
5. **Configuration** - Available config options
6. **Troubleshooting** - Common issues + solutions
7. **Development** - Contributing guide

---

## Expected Outcome

✅ Package structure designed with service layer architecture.

## Actual Outcome

Complete package structure defined with:
- Service-based architecture (fat services, thin tools)
- Clear module responsibilities
- Test organization strategy
- Package configuration (package.json, tsconfig.json)
- Architecture diagram showing component relationships

## Lessons Learned

1. Separate schemas from services to enable reuse
2. CliAdapter should be injectable for testing
3. Config should support customization (tool prefix, timeouts)
4. Tests mirror source structure for easy navigation
