# AGENTS.md

## Build & Test Commands

Always run commands from the project root using `mise run <command>`.

Do NOT use `bun` directly for tests/build - use `mise run`.

- **Build**: `mise run build` — Compiles to native binary at `dist/opennotes`
- **Test**: `mise run test` — Run all tests via vitest
- **Single Test**: `mise run test -- ConfigService.test.ts` — Run one test file
- **Watch Mode**: `mise run test -- --watch` — Re-run on changes
- **Lint**: `mise run lint` — ESLint check
- **Fix Lint**: `mise run lint:fix` — Auto-fix linting issues
- **Format**: `mise run format` — Prettier formatting

## Code Style Guidelines

### Imports & Module System

- Use ES6 `import`/`export` syntax only (module: "ESNext", type: "module")
- Group imports: external libraries first, then internal modules
- Use explicit file extensions (`.ts`) for all internal imports
- Import only what you need — no barrel exports

### Formatting (Prettier)

- **Single quotes**: `'import x from 'y''`
- **Line width**: 100 characters
- **Tab width**: 2 spaces
- **Trailing commas**: ES5 (no trailing in function parameters)
- **Semicolons**: enabled

### TypeScript & Naming Conventions

- **Strict mode**: enforced (`"strict": true`)
- **Classes**: PascalCase (`ConfigService`, `NotebookService`)
- **Functions/Methods**: camelCase
- **Constants**: SCREAMING_SNAKE_CASE (true constants only)
- **Type unions for status**: `'pending' | 'running' | 'completed'` not strings
- **Explicit types**: Always annotate parameters and return types
- **NeverNesters**: Exit early, avoid deeply nested conditionals

### Error Handling

- Always check error type: `error instanceof Error ? error.message : String(error)`
- Use Logger service for errors: `Log.error('Context: %s', message)`
- Never use `console.log` (linting error)
- Catch and handle all async/await errors

### Linting Rules Enforced

- `no-console`: **Error** (use LoggerService instead)
- `prettier/prettier`: **Error** (run `mise run format`)
- `@typescript-eslint/no-explicit-any`: **Warn** (avoid `any`)
- `@typescript-eslint/no-unused-vars`: **Error**

## Testing

- Framework: **Vitest** (Bun's test runner)
- Import: `import { describe, it, expect } from 'vitest'`
- Pattern: Nested `describe()` blocks with clear test names
- Use `expect()` for assertions

## Project Context

- **Type**: CLI tool for managing markdown-based notes
- **Target**: Bun runtime, ES2021+, TypeScript strict mode
- **Framework**: Clerc CLI with plugin architecture

## Architecture Overview

### Service-Oriented Design

Core services initialized in root interceptor (`src/index.ts`):

- **ConfigService**: Global user config (~/.config/opennotes/config.json)
- **DbService**: DuckDB in-memory instance with markdown extension
- **NotebookService**: Notebook discovery & operations
- **NoteService**: Note queries via DuckDB SQL

### Command Structure

Commands follow Clerc pattern:

```typescript
export const CommandName = defineCommand(
  {
    name: 'command-name',
    description: 'Help text',
    flags: {
      /* flags */
    },
  },
  async (ctx) => {
    // Access services: ctx.store.config, ctx.store.notebooKService
  }
);
```

### Data Flow

1. CLI parsed → Command matched
2. Root interceptor initializes services into `ctx.store`
3. `requireNotebookMiddleware()` resolves notebook (flag → config → ancestor search)
4. Command handler executes with access to services
5. Results rendered via `TuiRender()` using Binja templates

### Key Components

**ConfigService**: Manages user notebooks registry and global settings via Arktype validation.

**NotebookService**: Abstracts notebook operations — discovery, loading `.opennotes.json`, template management, context matching.

**NoteService**: Provides SQL query interface via DuckDB markdown extension for searching/reading notes.

**DbService**: Singleton DuckDB connection with pre-loaded markdown extension.
