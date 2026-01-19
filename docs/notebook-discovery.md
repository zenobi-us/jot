```
                           ┌─────────────┐
                           │   START     │
                           │ (cwd given) │
                           └──────┬──────┘
                                  │
                                  ▼
                        ┌─────────────────┐
                        │ CHECK DECLARED  │
                        │   NOTEBOOK      │◄──── From CLI flag --notebook
                        │     PATH        │      or env OPENNOTES_NOTEBOOK
                        └─────────┬───────┘
                                  │
                           ┌──────▼──────┐
                           │ Has notebook │
                           │ config file? │
                           └──────┬──────┘
                                  │
                       ┌─────────────────────┐
                       │ YES                 │ NO
                       ▼                     ▼
               ┌───────────────┐    ┌─────────────────┐
               │  LOAD & OPEN  │    │ CHECK REGISTERED │
               │   NOTEBOOK    │    │   NOTEBOOKS     │
               └───────┬───────┘    └─────────┬───────┘
                       │                      │
                       │              ┌───────▼────────┐
                       │              │ For each path  │
                       │              │ in config:     │
                       │              │   notebooks[]  │
                       │              └───────┬────────┘
                       │                      │
                       │              ┌───────▼────────┐
                       │              │ Has notebook   │
                       │              │ config file?   │
                       │              └───────┬────────┘
                       │                      │
                       │           ┌─────────────────────┐
                       │           │ YES                 │ NO
                       │           ▼                     ▼
                       │   ┌───────────────┐     ┌──────────────┐
                       │   │ Load notebook │     │ Try next     │
                       │   │   config      │     │ registered   │
                       │   └───────┬───────┘     │   notebook   │
                       │           │             └──────┬───────┘
                       │           ▼                    │
                       │   ┌───────────────┐           │
                       │   │ Check context │           │
                       │   │    match?     │           │
                       │   └───────┬───────┘           │
                       │           │                   │
                       │    ┌─────────────────┐        │
                       │    │ YES            │ NO     │
                       │    ▼                ▼        │
                       │ ┌─────────┐   ┌──────────────┼────┘
                       │ │  OPEN   │   │ Continue     │
                       │ │NOTEBOOK │   │ searching    │
                       │ └────┬────┘   └──────────────┘
                       │      │
                       │      │        ┌─────────────────┐
                       │      │        │ No matches in   │
                       │      │        │ registered list │
                       │      │        └─────────┬───────┘
                       │      │                  │
                       │      │                  ▼
                       │      │        ┌─────────────────┐
                       │      │        │ START ANCESTOR  │
                       │      │        │     SEARCH      │
                       │      │        └─────────┬───────┘
                       │      │                  │
                       │      │          ┌───────▼────────┐
                       │      │          │ current = cwd  │
                       │      │          └───────┬────────┘
                       │      │                  │
                       │      │           ┌──────▼──────┐
                       │      │           │ current == │
                       │      │           │  "/" or ""? │
                       │      │           └──────┬──────┘
                       │      │                  │
                       │      │         ┌────────────────┐
                       │      │         │ YES           │ NO
                       │      │         ▼               ▼
                       │      │ ┌───────────────┐ ┌─────────────────┐
                       │      │ │  NOT FOUND    │ │ Has notebook    │
                       │      │ │ (return nil)  │ │ config file?    │
                       │      │ └───────────────┘ └─────────┬───────┘
                       │      │                            │
                       │      │                  ┌─────────────────┐
                       │      │                  │ YES            │ NO
                       │      │                  ▼                ▼
                       │      │          ┌───────────────┐ ┌──────────────┐
                       │      │          │  LOAD & OPEN  │ │ current =    │
                       │      │          │   NOTEBOOK    │ │ parent dir   │
                       │      │          └───────┬───────┘ └──────┬───────┘
                       │      │                  │                │
                       │      │                  │                │
                       │      └──────────────────┼────────────────┘
                                │                │
                                ▼                │
                       ┌─────────────────┐       │
                       │     SUCCESS     │       │
                       │ (return notebook│◄──────┘
                       │    instance)    │
                       └─────────────────┘
```

## Context Matching Algorithm

   For notebook with contexts: ["/home/user/project", "/tmp/work"]
   Current working directory: "/home/user/project/src"

   Match check: strings.HasPrefix("/home/user/project/src", "/home/user/project")
   Result: TRUE → Context matches → Return this notebook

### State Transitions

   1. DECLARED PATH → Success or Continue
   2. REGISTERED SEARCH → For each registered notebook
      - Check exists → Check context match → Success or Continue
   3. ANCESTOR SEARCH → Walk up directories until found or root
   4. SUCCESS → Return notebook instance
   5. NOT FOUND → Return nil

### File Locations

   - Global config: ~/.config/opennotes/config.json
     Contains: {"notebooks": ["/path1", "/path2"]}

   - Notebook config: <notebook_dir>/.opennotes.json
     Contains: {"root": "./notes", "name": "Project", "contexts": [...]}

### The notebook discovery follows a 3-tier priority system:

 1. Declared Path (highest priority) - From --notebook flag or OPENNOTES_NOTEBOOK env var
 2. Registered Notebooks (medium priority) - Check each registered notebook for context match
 3. Ancestor Search (fallback) - Walk up directory tree looking for .opennotes.json
