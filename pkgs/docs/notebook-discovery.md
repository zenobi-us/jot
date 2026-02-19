# Notebook Discovery

Jot automatically discovers and loads notebooks based on the user's current working directory using a sophisticated 3-tier priority system. This document outlines the complete algorithm and provides a visual flowchart of the discovery process.

## Overview

The notebook discovery follows a **5-tier priority system** (first wins):

1. **JOT_NOTEBOOK** (highest priority) - Environment variable
2. **--notebook flag** (CLI override) - Command line argument
3. **Current Directory** (direct check) - `.jot.json` in current working directory
4. **Registered Notebooks** (context matching) - Check each registered notebook for context match
5. **Ancestor Search** (fallback) - Walk up directory tree looking for `.jot.json`

## Discovery Flowchart

```d2 exec="d2 - docs/notebook-discovery.svg" replace="![Notebook Discovery Flowchart](notebook-discovery.svg)"
direction: down

# Start node
Start: {
  label: "Start:\nCurrent Working Directory"
  shape: oval
  style: {
    fill: "#e1f5fe"
    stroke: "#01579b"
    stroke-width: 2
  }
}

# Tier 1: JOT_NOTEBOOK envvar
CheckEnvvar: {
  label: "Check\nJOT_NOTEBOOK\nenvironment variable"
  shape: diamond
  style: {
    fill: "#c8e6c9"
    stroke: "#1b5e20"
    stroke-width: 2
  }
}

HasEnvvarConfig: {
  label: "Has .jot.json\nin JOT_NOTEBOOK\npath?"
  shape: diamond
  style: {
    fill: "#c8e6c9"
    stroke: "#1b5e20"
    stroke-width: 2
  }
}

LoadEnvvar: {
  label: "Load & Open\nEnvvar Notebook\n(Tier 1: HIGHEST PRIORITY)"
  shape: rectangle
  style: {
    fill: "#c8e6c9"
    stroke: "#1b5e20"
    stroke-width: 2
  }
}

# Tier 2: --notebook flag
CheckFlag: {
  label: "Check\n--notebook CLI flag"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

HasFlagConfig: {
  label: "Has .jot.json\nin flag path?"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

LoadFlag: {
  label: "Load & Open\nFlag Notebook\n(Tier 2)"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

# Tier 3: Current Directory
CheckCurrentDir: {
  label: "Check Current\nDirectory for\n.jot.json"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

LoadCurrentDir: {
  label: "Load & Open\nCurrent Directory Notebook\n(Tier 3)"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

# Tier 4: Registered Notebooks
CheckRegistered: {
  label: "Check Registered\nNotebooks from\nglobal config"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

ForEachRegistered: {
  label: "For each registered\nnotebook path"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

HasRegisteredConfig: {
  label: "Has .jot.json\nin registered path?"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

LoadRegisteredConfig: {
  label: "Load notebook config"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

CheckContext: {
  label: "Current directory\nmatches any context\nin notebook?"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

LoadRegistered: {
  label: "Load & Open\nMatched Notebook"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

NextRegistered: {
  label: "Try next\nregistered notebook"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

# Tier 5: Ancestor Search
StartAncestorSearch: {
  label: "Start Ancestor Search\ncurrent = parent(cwd)"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

IsRoot: {
  label: "current == '/' or\nempty string?"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

HasAncestorConfig: {
  label: "Has .jot.json\nin current directory?"
  shape: diamond
  style: {
    fill: "#fff3e0"
    stroke: "#ef6c00"
    stroke-width: 2
  }
}

LoadAncestor: {
  label: "Load & Open\nAncestor Notebook"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

GoToParent: {
  label: "current = parent\ndirectory"
  shape: rectangle
  style: {
    fill: "#f3e5f5"
    stroke: "#4a148c"
    stroke-width: 2
  }
}

# End nodes
Success: {
  label: "Success\nReturn Notebook Instance"
  shape: oval
  style: {
    fill: "#e8f5e8"
    stroke: "#1b5e20"
    stroke-width: 2
  }
}

NotFound: {
  label: "Not Found\nReturn nil"
  shape: oval
  style: {
    fill: "#ffebee"
    stroke: "#c62828"
    stroke-width: 2
  }
}

# Connections
Start -> CheckEnvvar

CheckEnvvar -> HasEnvvarConfig: "Envvar set"
CheckEnvvar -> CheckFlag: "No envvar"

HasEnvvarConfig -> LoadEnvvar: "Yes"
HasEnvvarConfig -> CheckFlag: "No"

CheckFlag -> HasFlagConfig: "Flag provided"
CheckFlag -> CheckCurrentDir: "No flag"

HasFlagConfig -> LoadFlag: "Yes"
HasFlagConfig -> CheckCurrentDir: "No"

CheckCurrentDir -> LoadCurrentDir: "Yes"
CheckCurrentDir -> CheckRegistered: "No"

CheckRegistered -> ForEachRegistered
ForEachRegistered -> HasRegisteredConfig

HasRegisteredConfig -> NextRegistered: "No"
HasRegisteredConfig -> LoadRegisteredConfig: "Yes"

LoadRegisteredConfig -> CheckContext

CheckContext -> LoadRegistered: "Yes"
CheckContext -> NextRegistered: "No"

NextRegistered -> HasRegisteredConfig: "More notebooks"
NextRegistered -> StartAncestorSearch: "No more notebooks"

StartAncestorSearch -> IsRoot

IsRoot -> NotFound: "Yes"
IsRoot -> HasAncestorConfig: "No"

HasAncestorConfig -> LoadAncestor: "Yes"
HasAncestorConfig -> GoToParent: "No"

GoToParent -> IsRoot

LoadEnvvar -> Success
LoadFlag -> Success
LoadCurrentDir -> Success
LoadRegistered -> Success
LoadAncestor -> Success
```

### 1. JOT_NOTEBOOK Environment Variable (Tier 1 - Highest Priority)

The system first checks if the `JOT_NOTEBOOK` environment variable is set:

```bash
export JOT_NOTEBOOK=/path/to/notebook
jot notes list  # Uses the notebook from envvar
```

If the envvar is set:
1. Check if `.jot.json` exists in that path
2. If yes: Load and open the notebook → **SUCCESS**
3. If no: Continue to Tier 2

### 2. --notebook CLI Flag (Tier 2)

The system checks if the `--notebook` flag was provided on the command line:

```bash
jot notes list --notebook /path/to/notebook
```

If a flag path exists:
1. Check if `.jot.json` exists in that path
2. If yes: Load and open the notebook → **SUCCESS**
3. If no: Continue to Tier 3

### 3. Current Directory (Tier 3)

The system checks if `.jot.json` exists in the current working directory:

```bash
cd /home/user/project  # Contains .jot.json
jot notes list   # Auto-discovers notebook in cwd
```

If `.jot.json` exists in current directory:
1. Load and open the notebook → **SUCCESS**
2. If no: Continue to Tier 4

### 4. Registered Notebooks (Tier 4 - Context Matching)

The system checks notebooks registered in the global configuration:

1. Load global config from `~/.config/jot/config.json`
2. For each registered notebook path:
   - Check if `.jot.json` exists
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

### 5. Ancestor Search (Tier 5 - Fallback)

If no environment variable, flag, current directory, or registered notebooks match, the system performs an ancestor directory search:

1. Start with parent directory (not current, as that was checked in Tier 3)
2. Check if `.jot.json` exists in current directory
3. If yes: Load and open the notebook → **SUCCESS**
4. If no: Move to parent directory
5. Repeat until reaching filesystem root (`/`) or empty string
6. If root reached: → **NOT FOUND**

## File Locations & Formats

### Global Configuration
**Location:** `~/.config/jot/config.json`

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
**Location:** `<notebook_directory>/.jot.json`

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

1. **TIER 1: JOT_NOTEBOOK envvar** → Success or Continue to Tier 2
2. **TIER 2: --notebook flag** → Success or Continue to Tier 3
3. **TIER 3: Current Directory** → Success or Continue to Tier 4
4. **TIER 4: REGISTERED SEARCH** → For each registered notebook:
   - Check exists → Check context match → Success or Continue
5. **TIER 5: ANCESTOR SEARCH** → Walk up directories until found or root
6. **SUCCESS** → Return notebook instance
7. **NOT FOUND** → Return nil

This discovery system ensures Jot works seamlessly across different project environments while maintaining predictable, efficient behavior. The priority order follows the principle of least surprise: environment variable (global) → flag (explicit) → auto-detection (implicit).
