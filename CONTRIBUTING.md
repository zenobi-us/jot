# Contributing to opennotes

Thank you for your interest in contributing! This guide will help you get started with development.

## Getting Started

### Prerequisites

- [Bun](https://bun.sh) - JavaScript runtime
- [Mise](https://mise.jdx.dev) - Task runner

### Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/zenobi-us/opennotes.git
   cd opennotes
   ```

2. **Install dependencies:**

   ```bash
   mise run setup
   ```

3. **Verify setup:**
   ```bash
   mise run build
   mise run test
   ```

## Development Workflow

### Build & Test

All commands should be run from the project root using `mise run`:

- `mise run build` - Compile to native binary at `dist/opennotes`
- `mise run test` - Run all tests
- `mise run test -- ConfigService.test.ts` - Run a single test file
- `mise run test -- --watch` - Run tests in watch mode
- `mise run lint` - Check code style
- `mise run lint:fix` - Auto-fix linting issues
- `mise run format` - Format code with Prettier

**Important:** Do NOT use `bun` directly for builds or tests. Always use `mise run`.

## Code Style

We follow strict TypeScript and code style guidelines. See [AGENTS.md](AGENTS.md) for detailed information on:

- Module system and imports
- Formatting with Prettier (single quotes, 100 char lines, 2 space tabs)
- TypeScript naming conventions (PascalCase for classes, camelCase for functions)
- Error handling patterns
- Linting rules

**Quick reference:**

- Use ES6 `import`/`export`
- Add explicit `.ts` file extensions for internal imports
- Always annotate parameter and return types
- Exit early to avoid nested conditionals
- Use LoggerService instead of `console.log`

Run `mise run format` and `mise run lint:fix` before committing.

## Project Architecture

opennotes is a CLI tool with a service-oriented architecture:

- **Services** - Core business logic (ConfigService, NotebookService, NoteService, DbService)
- **Commands** - CLI command handlers using Clerc framework
- **Middleware** - Notebook resolution and dependency injection
- **Storage** - DuckDB for SQL queries across markdown files

See [AGENTS.md](AGENTS.md) for detailed architecture overview.

## Making Changes

1. **Create a branch** from `main`
2. **Make your changes** and write tests
3. **Run tests & lint:** `mise run test && mise run lint:fix && mise run format`
4. **Commit** with [conventional commits](https://www.conventionalcommits.org/):
   - `feat: add search filtering`
   - `fix: notebook discovery bug`
   - `docs: update README`
   - `refactor: simplify service initialization`

5. **Push your branch** and open a pull request

## Pull Request Process

- Use a descriptive title following conventional commits
- Describe what changed and why
- Reference any related issues
- Ensure all tests pass
- Get approval from maintainers

## Reporting Issues

When filing an issue, include:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Bun version, opennotes version)

## Questions?

Feel free to open a discussion or issue on GitHub. We're here to help!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
