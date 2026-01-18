# AGENTS.md

**OpenNotes** is a **Go CLI tool** for managing markdown-based notes organized in notebooks. It uses DuckDB for SQL-powered search and supports templates. The project is production-ready with comprehensive testing and clean architecture.

## Build & Test Commands

Always run commands from the project root using `mise run <command>`.

Do NOT use `go` directly for tests/build - use `mise run`.

- **Build**: `mise run build` — Compiles to native binary at `dist/opennotes`
- **Test**: `mise run test` — Run all tests (161+ tests, ~4 seconds)
- **Single Test**: `mise run test -- NoteService` — Run one test package
- **Lint**: `mise run lint` — Check code quality
- **Lint Fix**: `mise run lint:fix` — Auto-fix linting issues
- **Format**: `mise run format` — Format code with gofmt

## Code Style Guidelines

### Go Language Conventions

- **Module System**: Standard Go `import` statements with `github.com/zenobi-us/opennotes` base
- **Import Order**: Standard library → external packages → internal modules
- **Naming**:
  - **Types/Structs**: PascalCase (`ConfigService`, `NotebookService`, `Note`)
  - **Functions/Methods**: camelCase (`SearchNotes`, `DisplayName`)
  - **Constants**: SCREAMING_SNAKE_CASE only for true constants
  - **Receivers**: Single letter (e.g., `(d *Display)`, `(s *Service)`)
- **Formatting**:
  - Enforced by `gofmt` (run via `mise run format`)
  - Max line length: No strict limit (but keep reasonable ~100 chars)
  - Indentation: Tabs (Go standard)

### Type Safety & Error Handling

- **Strict typing**: Always specify return types, no implicit types
- **Error handling**: Always check errors immediately
  ```go
  result, err := someFunc()
  if err != nil {
    Log.Error("context", err)
    return err
  }
  ```
- **Nil checks**: Always check for nil before using pointers
- **Error wrapping**: Use `fmt.Errorf("action failed: %w", err)` for context
- **Logger usage**: Use `Log` namespace logger: `Log.Error("context", err)`
- **No panics**: Recover in main only, handle errors gracefully

### Testing Standards

- **Framework**: Go's built-in `testing` package
- **Test file format**: `*_test.go` files in same package
- **Test names**: `Test<Type>_<Method>_<Scenario>` (e.g., `TestNoteService_SearchNotes_FindsAllNotes`)
- **Subtests**: Use `t.Run()` for testing multiple scenarios
- **Table-driven tests**: Use slice of test cases for variations
- **Assertions**: Use `testify/assert` or manual `if` checks
- **Setup/Teardown**: Use helper functions like `createTestNotebook()`

### Command Philosophy: Thin Commands, Fat Services

Commands in `cmd/` are **thin orchestration layers only**. All business logic belongs in `internal/services/`.

**Command Responsibilities** (Limit to):
- Parse CLI flags and arguments
- Call one or more service methods
- Format and display output (via services)
- Handle command-level errors
- Return early on error (don't accumulate logic)

**NOT Command Responsibilities**:
- Business logic (queries, validation, transformations)
- Data persistence or file operations
- External API calls
- Complex control flow or conditional logic
- Type conversions or data manipulation

**Pattern Example**:

```go
// ❌ BAD: Business logic directly in command
var searchCmd = &cobra.Command{
  RunE: func(cmd *cobra.Command, args []string) error {
    query := args[0]
    results := []string{}
    // Direct business logic - belongs in service!
    for _, note := range allNotes {
      if strings.Contains(note.Content, query) {
        results = append(results, note.Title)
      }
    }
    fmt.Printf("Found %d results\n", len(results))
    return nil
  },
}

// ✅ GOOD: Delegated to service, command stays thin
var searchCmd = &cobra.Command{
  RunE: func(cmd *cobra.Command, args []string) error {
    // Step 1: Parse
    nb, err := requireNotebook(cmd)
    if err != nil {
      return err
    }
    
    // Step 2: Call service (all logic there)
    results, err := nb.Notes.SearchNotes(context.Background(), args[0])
    if err != nil {
      return fmt.Errorf("search failed: %w", err)
    }
    
    // Step 3: Display via template
    return displayNoteList(results)
  },
}
```

**Guideline**: If your command's `RunE` function exceeds 50 lines, extract logic to a service method.

**Current Command Size** (All OK):
- Smallest: 32 lines (init.go)
- Average: 76 lines
- Largest: 125 lines (notes_add.go) - but within reasonable limit

### DRY/WET/AHA Philosophy

We follow **AHA Principles** (Avoid Hasty Abstractions) over strict DRY enforcement.

**When to Extract Duplicated Code**:

| Occurrence | Action | Rationale |
|-----------|--------|-----------|
| **1st** | Accept as baseline | Learn the pattern |
| **2nd** | Document & consider | Is the pattern obvious? Can it evolve differently? |
| **3rd** | Extract to shared function | Clear pattern, worth the abstraction |
| **4+** | Mandatory refactoring | Duplication becomes maintenance burden |

**DRY (Don't Repeat Yourself)**: Extract only when:
1. Code is >80% identical between locations
2. Changes must be synchronized across multiple places
3. The abstraction is obvious and naming is clear
4. You've seen the pattern repeat at least 3 times

**WET (Write Everything Twice)**: Acceptable when:
1. Abstractions feel forced or require complex parameters
2. Shared code would obscure each caller's specific intent
3. The code may evolve differently in each location
4. Performance is critical and abstraction adds overhead

**AHA (Avoid Hasty Abstractions)**: 
- Prefer clear, simple code over premature abstraction
- Allow limited duplication in early stages
- Extract only when pattern is proven and stable

**Example: Template Display Pattern**

Current code has `displayNoteList()` and `displayNotebookList()` (~60% similar):

```go
// Both follow same pattern:
// 1. Call TuiRender with template
// 2. If error, fallback to manual fmt.Printf
// 3. Print result
```

**Why NOT extracted yet:**
- Only 2 occurrences (waiting for 3rd per AHA)
- Different data types (Note vs Notebook)
- Fallback formatting is type-specific
- Premature abstraction would require generics/interfaces
- Pattern may diverge (notes might need different fallback soon)

**When to extract**: After a 3rd similar display function is created, extract to `displayViaTemplate()`.

### Duplicate Logic Detection & Refactoring Process

Systematically scan for and refactor duplicated code. This prevents maintenance burden and keeps code DRY.

**Frequency**: Monthly or during refactoring sprints (not continuous refactoring)

**Detection Tools & Techniques**:

1. **CodeMapper (cm) - AST-based analysis**
   ```bash
   # Get project overview
   cm stats .
   
   # Find all usages of a pattern
   cm query "TuiRender" --format ai
   cm callers "displayNoteList" --format ai
   ```

2. **Manual Pattern Scan**
   ```bash
   # Find all template renders
   grep -n "TuiRender" cmd/*.go
   
   # Find all SQL displays
   grep -n "RenderSQLResults" cmd/*.go
   ```

3. **Code Review Process**
   - During PR review, flag code that "feels familiar"
   - Ask: "Have I written similar code elsewhere?"
   - Document potential duplication for monthly audit

**Patterns Currently Being Watched**:

1. **`requireNotebook()` pattern** (8+ occurrences)
   ```go
   nb, err := requireNotebook(cmd)
   if err != nil {
     return err
   }
   ```
   Status: ✅ Already extracted helper function
   Future: Consider centralizing to `cmd/root.go`

2. **Display template pattern** (3-4 occurrences)
   ```go
   output, err := services.TuiRender(template, data)
   if err != nil {
     // fallback to fmt.Printf
   }
   fmt.Print(output)
   ```
   Status: ⚠️ At extraction threshold - watch for 3rd function

3. **Flag parsing pattern** (3-4 occurrences)
   ```go
   notebook, _ := cmd.Flags().GetString("notebook")
   ```
   Status: 🔴 Extract to helper: `getNotebookFlag(cmd)`

**Extraction Workflow** (Test-Driven):

1. Write tests for the duplicated behavior
2. Create shared function with clear, descriptive name
3. Update all callers to use shared function
4. Run full test suite (`mise run test`)
5. Commit with message: `refactor: extract <pattern> to shared function`

**Monthly Audit Checklist**:

- [ ] Run `cm stats . --format ai` to get codebase overview
- [ ] Review recent commit messages for obvious duplication patterns
- [ ] Check `cmd/*.go` directory for >2 similar code blocks
- [ ] Run `grep` for patterns: TuiRender, RenderSQLResults, requireNotebook
- [ ] Create GitHub issue if 3rd occurrence found
- [ ] Prioritize extraction in next refactoring sprint
- [ ] Update this section if new patterns emerge

**Integration with External Skills**:

- **refactoring-specialist**: Use for extraction pattern guidance
- **codemapper**: Use `cm` tool for AST-based pattern detection
- **defense-in-depth**: Apply for validation at multiple layers when extracting

## Project Context

- **Type**: CLI tool for managing markdown-based notes
- **Language**: Go (1.18+)
- **Runtime Target**: Native binary (Linux, macOS, Windows)
- **Database**: DuckDB with markdown extension
- **Status**: Production-ready, fully tested

## Architecture Overview

### Service-Oriented Design

Core services are singletons initialized in `cmd/root.go`:

- **ConfigService** (`internal/services/config.go`): Global user config (~/.config/opennotes/config.json)
- **DbService** (`internal/services/db.go`): DuckDB connections with markdown extension
- **NotebookService** (`internal/services/notebook.go`): Notebook discovery & operations
- **NoteService** (`internal/services/note.go`): Note queries via DuckDB SQL
- **DisplayService** (`internal/services/display.go`): Terminal rendering with glamour
- **LoggerService** (`internal/services/logger.go`): Structured logging (zap-based)

### Command Structure

Commands are defined in `cmd/` directory and follow standard Cobra CLI pattern:

```go
var listCmd = &cobra.Command{
  Use:   "list",
  Short: "List notes in notebook",
  RunE: func(cmd *cobra.Command, args []string) error {
    // Access services via global variables
    notes, err := noteService.SearchNotes(query)
    // Render output
    output, err := TuiRender("note-list", data)
    fmt.Println(output)
    return nil
  },
}
```

### Data Flow

1. CLI parses arguments → Matches command
2. `cmd/root.go` initializes services (lazy-loaded)
3. Command handler retrieves notebook (via flag, config, or ancestor search)
4. Services execute business logic (config, database, file operations)
5. Results formatted and rendered via `TuiRender()` with templates
6. Output displayed to user with glamour markdown rendering

### Key Components

**ConfigService**: Manages registered notebooks, global settings. Supports env var overrides.

**NotebookService**: Discovers notebooks, loads `.opennotes.json` config, manages notebook lifecycle.

**NoteService**: Provides SQL query interface. Validates queries (SELECT/WITH only), handles metadata extraction.

**DbService**: Manages DuckDB connections (read-write and read-only). Pre-loads markdown extension.

**DisplayService**: Renders markdown with glamour, formats SQL results as ASCII tables.

### Templates

Templates are stored as `.gotmpl` files in `internal/services/templates/` and embedded using `go:embed`:

- `note-list.gotmpl` - Display list of notes
- `note-detail.gotmpl` - Display individual note
- `notebook-info.gotmpl` - Display notebook configuration
- `notebook-list.gotmpl` - Display all notebooks

Loaded via `TuiRender(name string, ctx any)` function.

## Key Technical Decisions

### Language: Go

- **Why**: Native binary compilation, simplicity, performance
- **Performance**: Faster startup than Node/Bun, no runtime overhead
- **Deployment**: Single binary, no external dependencies for users
- **Alternative**: Previously TypeScript/Bun (removed 2026-01-18)

### Database: DuckDB

- **Why**: SQL support for notes, in-process, supports markdown extension
- **Current**: Using neo DuckDB (C++ version)
- **Future**: Considering wasm build when markdown extension support improves

### CLI Framework: Cobra

- **Why**: Standard Go CLI library, widely used, simple to extend
- **Structure**: Root command → Subcommands (notebook, notes, init)

### Service Architecture

- **Pattern**: Singleton services initialized once
- **Access**: Global variables in `cmd/root.go` (performance over purity)
- **Thread-safety**: Safe for concurrent access (DuckDB handles locking)

### Templates: go:embed

- **Why**: Embed templates at compile time, no runtime file access needed
- **Benefits**: Binary-portable, simpler deployment, no files to distribute
- **Trade-off**: Templates must be files in `templates/` directory

## File Structure

```
.
├── cmd/                          # CLI commands
│   ├── root.go                   # Service initialization
│   ├── init.go                   # Init command
│   ├── notebook_*.go             # Notebook commands
│   └── notes_*.go                # Notes commands
├── internal/
│   ├── core/                     # Utilities (validation, strings, etc.)
│   ├── services/                 # Core business logic
│   │   ├── config.go
│   │   ├── db.go
│   │   ├── notebook.go
│   │   ├── note.go
│   │   ├── display.go
│   │   ├── logger.go
│   │   ├── templates.go
│   │   ├── templates/            # .gotmpl template files
│   │   └── *_test.go
│   └── testutil/                 # Test helpers
├── tests/
│   └── e2e/                      # End-to-end tests
├── main.go                       # Entry point
├── go.mod                        # Go module definition
└── .misrc.yaml                   # Mise task configuration
```

## Code Examples

### Logging

```go
import "github.com/zenobi-us/opennotes/internal/services"

log := services.Log("MyService")
log.Debug("debug message")
log.Info("info message")
log.Warn("warning message")
log.Error("error message", err)
```

### Service Usage

```go
// Services are initialized globally in cmd/root.go
// Access them in command handlers:
notes, err := services.NoteService.SearchNotes(query)
if err != nil {
  return err
}

output, err := services.TuiRender("note-list", map[string]any{
  "Notes": notes,
})
if err != nil {
  return err
}
fmt.Println(output)
```

### Testing

```go
func TestNoteService_SearchNotes_FindsAllNotes(t *testing.T) {
  // Setup
  nb := testutil.CreateTestNotebook(t)
  ns := services.NewNoteService(nb)
  
  // Execute
  notes, err := ns.SearchNotes("")
  
  // Assert
  require.NoError(t, err)
  require.Len(t, notes, 2)
}
```

## Commands Overview

- **`opennotes init`** - Initialize configuration
- **`opennotes notebook create`** - Create new notebook
- **`opennotes notebook register`** - Register existing notebook
- **`opennotes notebook list`** - List all notebooks
- **`opennotes notes list`** - List notes in notebook (formatted with titles/slugified names)
- **`opennotes notes search <query>`** - Search notes by content
- **`opennotes notes add <name>`** - Add new note
- **`opennotes notes remove <name>`** - Remove note
- **`opennotes notes search --sql <query>`** - Execute SQL query on notes

## Test Coverage

- **161+ tests** across all packages
- **95%+ coverage** in core logic
- **28+ end-to-end tests** for CLI commands
- Test duration: ~4 seconds

Run with: `mise run test`

## Recent Changes

- ✅ **2026-01-18**: Refactored templates to separate `.gotmpl` files
- ✅ **2026-01-18**: Removed TypeScript/Node implementation (27 files, 1,797 lines)
- ✅ **2026-01-17**: Implemented notes list format feature (frontmatter titles)
- ✅ **2026-01-17**: Completed SQL flag support (--sql flag for search)
- ✅ **2026-01-09**: Full Go rewrite complete (TypeScript → Go migration)
