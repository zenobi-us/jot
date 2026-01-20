# Notebook Discovery

OpenNotes automatically discovers and loads notebooks based on the user's current working directory using a sophisticated 3-tier priority system. This document outlines the complete algorithm and provides a visual flowchart of the discovery process.

## Overview

The notebook discovery follows a **3-tier priority system**:

1. **Declared Path** (highest priority) - From `--notebook` flag or `OPENNOTES_NOTEBOOK` env var
2. **Registered Notebooks** (medium priority) - Check each registered notebook for context match
3. **Ancestor Search** (fallback) - Walk up directory tree looking for `.opennotes.json`

## Discovery Flowchart

![Notebook Discovery Flowchart](notebook-discovery.svg)

> **Note**: This diagram is generated from [notebook-discovery.d2](notebook-discovery.d2) using [D2](https://d2lang.com/).  
> To regenerate: `d2 docs/notebook-discovery.d2 docs/notebook-discovery.svg`

## Detailed Process

### 1. Declared Path (Tier 1 - Highest Priority)

The system first checks if a notebook path has been explicitly declared via:
- CLI flag: `opennotes --notebook /path/to/notebook`
- Environment variable: `OPENNOTES_NOTEBOOK=/path/to/notebook`

If a declared path exists:
1. Check if `.opennotes.json` exists in that path
2. If yes: Load and open the notebook → **SUCCESS**
3. If no: Continue to Tier 2

### 2. Registered Notebooks (Tier 2 - Context Matching)

The system checks notebooks registered in the global configuration:

1. Load global config from `~/.config/opennotes/config.json`
2. For each registered notebook path:
   - Check if `.opennotes.json` exists
   - If exists: Load notebook configuration
   - Check if current directory matches any context path using **Context Matching Algorithm**
   - If match found: Load and open the notebook → **SUCCESS**
3. If no registered notebooks match: Continue to Tier 3

#### Context Matching Algorithm

```go
// For each context path in notebook config
for _, context := range notebook.Contexts {
    if strings.HasPrefix(currentWorkingDirectory, context) {
        return notebook // Match found
    }
}
```

**Example:**
```
Notebook contexts: ["/home/user/project", "/tmp/work"]
Current directory: "/home/user/project/src"

Match check: strings.HasPrefix("/home/user/project/src", "/home/user/project")
Result: TRUE → Context matches → Return this notebook
```

### 3. Ancestor Search (Tier 3 - Fallback)

If no declared or registered notebooks match, the system performs an ancestor directory search:

1. Start with current working directory
2. Check if `.opennotes.json` exists in current directory
3. If yes: Load and open the notebook → **SUCCESS**
4. If no: Move to parent directory
5. Repeat until reaching filesystem root (`/`) or empty string
6. If root reached: → **NOT FOUND**

## File Locations & Formats

### Global Configuration
**Location:** `~/.config/opennotes/config.json`

```json
{
  "notebooks": [
    "/home/user/work-notebook",
    "/home/user/personal-notebook",
    "/tmp/temp-notebook"
  ]
}
```

### Notebook Configuration
**Location:** `<notebook_directory>/.opennotes.json`

```json
{
  "root": "./notes",
  "name": "Project Notebook",
  "contexts": [
    "/home/user/project",
    "/home/user/project-related"
  ],
  "templates": {
    "default": "# {{.Title}}\n\nDate: {{.Date}}\n\n"
  },
  "groups": [
    {
      "name": "Default",
      "globs": ["**/*.md"],
      "metadata": {}
    }
  ]
}
```

## Key Features

### Deterministic Behavior
- **Clear Priority**: Declared > Registered > Ancestor
- **First Match Wins**: Stops at first successful discovery
- **No Ambiguity**: Priority order prevents conflicts

### Graceful Fallback
- If higher priority method fails, try next tier
- Comprehensive search ensures maximum discovery success
- Returns `nil` only when all methods exhausted

### Context-Aware
- Registered notebooks define active contexts
- Automatically selects appropriate notebook for current work environment
- Supports multiple context paths per notebook

### Efficient Discovery
- Stops immediately upon successful match
- Minimal filesystem operations
- Fast context matching using string prefix comparison

## State Transitions Summary

1. **DECLARED PATH** → Success or Continue to Tier 2
2. **REGISTERED SEARCH** → For each registered notebook:
   - Check exists → Check context match → Success or Continue
3. **ANCESTOR SEARCH** → Walk up directories until found or root
4. **SUCCESS** → Return notebook instance
5. **NOT FOUND** → Return nil

This discovery system ensures OpenNotes works seamlessly across different project environments while maintaining predictable, efficient behavior.
