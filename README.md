# OpenNotes

![OpenNotes Banner](./banner.png)

A CLI tool for managing your markdown-based notes organized in notebooks with **powerful SQL querying and automation capabilities**.

## Why OpenNotes? (Key Differentiators)

Unlike basic note tools, OpenNotes provides:

- 🔍 **SQL-Powered Search** - Query notes using DuckDB's full SQL capabilities and markdown functions
- 📋 **Intelligent Markdown Parsing** - Extract structure, statistics, and metadata from markdown content
- 🤖 **Automation Ready** - JSON output designed for piping to jq, scripts, and external tools
- 📔 **Multi-Notebook Organization** - Manage multiple notebook contexts with smart auto-discovery
- 🎯 **Developer-First** - CLI-native, git-friendly, markdown-native, zero external runtime dependencies
- ⚡ **Fast & Lightweight** - Single compiled binary, in-process database, no setup required

## Installation

```bash
go install github.com/zenobi-us/opennotes@latest
```

Requires Go 1.24+. The binary will be placed in `$GOPATH/bin/`.

## Power User: 5-Minute Quick Start

### Import Your Existing Notes

OpenNotes works best with your existing markdown files. No migration needed—just point it at your notes directory:

```bash
# Initialize with your existing markdown folder
opennotes notebook create "My Notes" --path ~/my-notes

# List all notes (instantly discovers your markdown files)
opennotes notes list
```

### Unlock SQL Querying Power

Execute sophisticated queries against your entire notebook:

```bash
# Find all notes mentioning "deadline" (across entire notebook)
opennotes notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) WHERE content ILIKE '%deadline%' LIMIT 5"

# Get word count statistics sorted by complexity
opennotes notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY words DESC LIMIT 10"

# Find checked-off tasks
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%[x]%' ORDER BY file_path"
```

### Automation Ready: JSON Output

All SQL queries return clean JSON—perfect for piping to jq, scripts, and external tools:

```bash
# Export statistics to JSON
opennotes notes search --sql "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md', include_filepath:=true)" | jq 'map({path: .file_path, words: .word_count})'

# Calculate totals across your notebook
opennotes notes search --sql "SELECT (md_stats(content)).word_count FROM read_markdown('**/*.md')" | jq 'map(.word_count) | {total: add, count: length, average: add/length}'

# Integrate with external tools
opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" | jq -r '.[].file_path' | xargs wc -l
```

### Learn More

- 🚀 **[Getting Started for Power Users](docs/getting-started-power-users.md)** - Complete 15-minute onboarding with examples
- 📚 **[SQL Query Guide](docs/sql-guide.md)** - Full DuckDB markdown functions and patterns
- 🚀 **[Automation & JSON Integration](docs/json-sql-guide.md)** - Advanced piping and external tool examples
- 📋 **[Notebook Discovery](docs/notebook-discovery.md)** - Multi-notebook setup and context management

## Beginner: Basic Quick Start

If you prefer to start simple:

1. **Initialize a notebook:**

   ```bash
   opennotes init
   ```

2. **Create a note:**

   ```bash
   opennotes notes add "My First Note"
   ```

3. **List notes:**

   ```bash
   opennotes notes list
   ```

4. **Search notes:**
   ```bash
   opennotes notes search "keyword"
   ```

## Commands

### Notebook Management

- `opennotes notebook` - Display current notebook info
- `opennotes notebook list` - List all notebooks
- `opennotes notebook create <name>` - Create a new notebook

### Note Operations

- `opennotes notes list` - List all notes in current notebook
- `opennotes notes add <title>` - Create a new note
- `opennotes notes remove <path>` - Delete a note
- `opennotes notes search <query>` - Search notes

## Configuration

Global configuration is stored in:

- **Linux**: `~/.config/opennotes/config.json`
- **macOS**: `~/Library/Preferences/opennotes/config.json`
- **Windows**: `%APPDATA%\opennotes\config.json`

Each notebook has a `.opennotes.json` file with notebook-specific settings.

## Usage Examples

```bash
# Create a notebook
opennotes notebook create "Work"

# Add notes to your notebook
opennotes notes add "Team Meeting Notes"
opennotes notes add "Project Ideas"

# Search across notes
opennotes notes search "deadline"

# List all notes
opennotes notes list
```

## Advanced Usage

### Complete Power User Guide

For the most comprehensive walkthrough of OpenNotes capabilities:

- **[Getting Started for Power Users](docs/getting-started-power-users.md)** - 15-minute complete onboarding
  - Import existing markdown collections
  - Execute SQL queries with practical examples
  - Automate with JSON output and jq integration
  - Build real-world workflows

### SQL Query Reference

For complete documentation on available functions and advanced patterns:

- **[SQL Functions Reference](docs/sql-functions-reference.md)** - Complete DuckDB + markdown function reference
- **[SQL Guide](docs/sql-guide.md)** - Comprehensive query patterns and best practices
- **[JSON Output Guide](docs/json-sql-guide.md)** - Automation examples and tool integration

### Multi-Notebook Management

Manage multiple note collections with context-aware auto-discovery:

```bash
# View all registered notebooks
opennotes notebook list

# Switch context by directory (auto-discovers .opennotes.json)
cd ~/work/notes
opennotes notes list  # Automatically uses work notebook

# Use notebook flag to specify a specific notebook
opennotes notes list --notebook "Personal"
```

See [Notebook Discovery](docs/notebook-discovery.md) for advanced multi-notebook workflows.

## Contributing

Interested in contributing? See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code style guidelines, and how to submit pull requests.

## License

MIT License. See [LICENSE](LICENSE) for details.
