# opennotes

A CLI tool for managing your markdown-based notes organized in notebooks.

## Features

- 📔 **Notebook-based organization** - Group notes into logical notebooks
- 🔍 **SQL-powered search** - Query notes using DuckDB with full-text search
- 📝 **Markdown-native** - Store notes as plain markdown with metadata
- 🏗️ **Smart discovery** - Auto-detect notebooks from directory context
- 🎨 **Template support** - Create notes from templates
- ⚡ **Fast & lightweight** - Single binary, no dependencies at runtime

## Installation

```bash
bun install -g opennotes
```

## Quick Start

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

## Contributing

Interested in contributing? See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code style guidelines, and how to submit pull requests.

## License

MIT License. See [LICENSE](LICENSE) for details.
