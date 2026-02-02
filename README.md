# OpenNotes

![OpenNotes Banner](./banner.png)

A simple, fast, and powerful CLI tool for managing your markdown notes. OpenNotes helps you organize, search, and manage your thoughts without leaving the terminal.

## Features

- Notebook Storage: choose where you want the notes stored, here in your repo, there in your home directory, or anywhere else on your filesystem.
- Create, list, search, and view notes from the command line.
- Associate default metadata based on groups or folders.
- Preset queries: Kanban, Daily Notes, Tags, and more
- Full-text search with fuzzy matching and boolean queries

## Installation

```bash
go install github.com/zenobi-us/opennotes@latest
```

*Requires Go 1.24+*

## Quick Start

Get started in seconds:

1.  **Initialize a notebook** in your current directory:
    ```bash
    opennotes init
    ```

2.  **Create your first note**:
    ```bash
    opennotes notes add "My First Idea"
    ```

3.  **List your notes**:
    ```bash
    opennotes notes list
    ```

4.  **Search your notes**:
    ```bash
    opennotes notes search "Idea"
    ```

## Common Commands

### Creating Notes

Create a new note with a title. You can also specify a path if you want to organize notes into folders.

```bash
# Create a note in the root
opennotes notes add "Meeting Notes"

# Create a note in a subfolder
opennotes notes add "Project Specs" projects/
```

### Listing Notes

See all your notes in the current notebook.

```bash
opennotes notes list
```

### Searching Notes

Find notes instantly by keyword.

```bash
opennotes notes search "important"
```

### Views 

Opennotes supports various preset views to help you organize and visualize your notes.

Output is JSON by default.

```bash
# view all available views
opennotes notes view

# display notes in kanban view
opennotes notes view kanban
```

## Configuration

OpenNotes works out of the box, but you can customize it.

Global configuration is stored in:
- **Linux**: `~/.config/opennotes/config.json`
- **macOS**: `~/Library/Preferences/opennotes/config.json`
- **Windows**: `%APPDATA%\opennotes\config.json`

## Advanced Usage

OpenNotes provides powerful search capabilities with full-text search, fuzzy matching, and boolean query operators for complex filtering.

For advanced features like boolean queries, JSON output for automation, and multi-notebook management, please see our **[Advanced Documentation](docs/getting-started-power-users.md)**.
