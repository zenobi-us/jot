---
id: ca68615f
title: Note Creation Enhancement Specification
created_at: 2026-01-20T21:14:00+10:30
updated_at: 2026-01-20T21:14:00+10:30
status: proposed
epic_id: 3e01c563
phase_id: TBD
---

# Specification: Note Creation Enhancement

## Overview

**Feature Name**: Enhanced `opennotes notes add` Command  
**Epic**: Advanced Note Creation and Search Capabilities (3e01c563)  
**Status**: Proposed - Awaiting Implementation Planning  

### Summary

This specification defines the enhanced `opennotes notes add` command that:
1. **Modernizes the CLI interface** with positional arguments for better UX
2. **Maintains backward compatibility** during v1.x releases with deprecation warnings
3. **Adds metadata support** via `--data` flags for rich note frontmatter
4. **Enhances stdin integration** for piped content workflows
5. **Improves path resolution** with auto-detection of files vs folders

## Command Signature

### New Style (Preferred - v1.x+)
```bash
opennotes notes add <title> [path] [flags]
```

### Old Style (Deprecated - v1.x only)
```bash
opennotes notes add [path] --title "Title" [flags]
```

**Deprecation Timeline**:
- **v0.x**: Both syntaxes work, `--title` flag shows deprecation warning
- **v0.1.0**: `--title` flag removed, only positional title accepted

## Parameters

### Positional Arguments

| Position | Name | Required | Description |
|----------|------|----------|-------------|
| 1 | `<title>` | Conditional* | Note title (becomes frontmatter + H1) |
| 2 | `[path]` | Optional | File or folder path (auto-detected) |

**Required Conditions***: Title must be provided either as positional arg OR via `--title` flag (not both)

### Flags

| Flag | Type | Repeatable | Status | Description |
|------|------|-----------|---------|-------------|
| `--title` | `string` | No | DEPRECATED | Note title (v1.x only, shows warning) |
| `--template` | `string` | No | Existing | Template name from config |
| `--data` | `string` | **Yes** | **NEW** | Set frontmatter field (`field=value`) |

## Behavior Rules

> [!WARNING]
> Below should just be a guidance, because we should largely operate withing the framework 
> of our existing configuration system "Koanf", so how we resolve flags and args should be
> consistent with that system.
>
> Doing this makes Q happy.

### 1. Title Resolution Logic

```go
// Pseudo-code for title resolution
func resolveTitle(args []string, titleFlag string) (string, error) {
    hasPositionalTitle := len(args) > 0 && titleFlag == ""
    hasFlagTitle := titleFlag != ""
    
    if hasPositionalTitle && hasFlagTitle {
        return "", errors.New("cannot specify title both as argument and --title flag")
    }
    
    if !hasPositionalTitle && !hasFlagTitle {
        return "", errors.New("title is required (provide as first argument or --title flag)")
    }
    
    if hasFlagTitle {
        // Show deprecation warning
        Log.Warn("--title flag is deprecated, use positional argument instead. Will be removed in v2.0.0")
        return titleFlag, nil
    }
    
    return args[0], nil
}
```

**Resolution Priority**:
1. If positional title provided: Use it (new style) ✅
2. If `--title` flag provided: Use it + show deprecation warning ⚠️
3. If both provided: Error ❌
4. If neither provided: Error ❌

### 2. Argument Interpretation Logic

The interpretation of arguments depends on which syntax is used:

**New Style (Positional Title)**:
```bash
opennotes notes add "My Note"           # args[0] = title, args[1] = nil
opennotes notes add "My Note" path/     # args[0] = title, args[1] = path
```

**Old Style (Flag Title)**:
```bash
opennotes notes add --title "My Note"  # args[0] = nil (path unspecified)
opennotes notes add myfile.md --title "My Note"  # args[0] = path
```

**Implementation Logic**:
```go
func parseArguments(args []string, titleFlag string) (title, path string, err error) {
    if titleFlag != "" {
        // Old style: args[0] is path (if provided)
        title = titleFlag
        if len(args) > 0 {
            path = args[0]
        }
    } else {
        // New style: args[0] is title, args[1] is path
        if len(args) == 0 {
            return "", "", errors.New("title is required")
        }
        title = args[0]
        if len(args) > 1 {
            path = args[1]
        }
    }
    return title, path, nil
}
```

### 3. Content Priority and Generation

Content is generated with the following priority (highest to lowest):

| Priority | Source | Condition |
|----------|--------|-----------|
| 1 (Highest) | **Stdin** | If stdin is not empty |
| 2 | **Template** | If `--template` flag specified and stdin empty |
| 3 (Lowest) | **Default** | If neither stdin nor template provided |

**Content Generation Logic**:
```go
func generateContent(title string, stdinContent string, templateName string) string {
    // Priority 1: Stdin wins
    if stdinContent != "" {
        return stdinContent
    }
    
    // Priority 2: Template (if specified)
    if templateName != "" {
        return loadTemplate(templateName)
    }
    
    // Priority 3: Default content
    return fmt.Sprintf("# %s\n\n", title)
}
```

**Key Principle**: Stdin content ALWAYS wins over template content to support piping workflows.

### 4. Frontmatter Handling

Frontmatter is constructed from multiple sources:

| Field | Source | Overridable | Notes |
|-------|--------|-------------|-------|
| `created` | Auto-generated | No | RFC3339 timestamp, preserved from current implementation |
| `title` | `<title>` arg or `--title` flag | Yes (via `--data`) | Shows warning if set via `--data` |
| Custom fields | `--data` flags | Yes | Repeatable, supports multiple values |

**Frontmatter Generation Logic**:
```go
func generateFrontmatter(title string, dataFlags []string) map[string]interface{} {
    fm := map[string]interface{}{
        "created": time.Now().Format(time.RFC3339),
        "title":   title,
    }
    
    // Apply custom --data flags
    for _, dataFlag := range dataFlags {
        field, value := parseDataFlag(dataFlag)
        
        // Special case: warn about title override
        if field == "title" {
            Log.Warn("title field is deprecated in --data, use positional argument. Will be removed in v2.0.0")
        }
        
        // Support multiple values for same field (e.g., tags)
        if existing, ok := fm[field]; ok {
            // Convert to array if not already
            switch v := existing.(type) {
            case []interface{}:
                fm[field] = append(v, value)
            default:
                fm[field] = []interface{}{v, value}
            }
        } else {
            fm[field] = value
        }
    }
    
    return fm
}
```

**Data Flag Format**:
```bash
--data field=value       # String value
--data tag=meeting       # First tag
--data tag=sprint        # Second tag (creates array)
--data priority=high     # Any custom field
```

### 5. Path Resolution (Auto-detection)

Path resolution supports multiple formats with automatic detection:

| Input | Type | Output |
|-------|------|--------|
| Omitted | Default | `<notebook-root>/{slugified-title}.md` |
| `folder/` | Folder (ends with `/`) | `<notebook-root>/folder/{slugified-title}.md` |
| `existing/dir` | Folder (existing directory) | `<notebook-root>/existing/dir/{slugified-title}.md` |
| `path/file.md` | Full filepath | `<notebook-root>/path/file.md` |
| `path/file` | No extension | `<notebook-root>/path/file.md` (auto-add `.md`) |

**Path Resolution Algorithm**:
```go
func resolvePath(notebookRoot, inputPath, slugifiedTitle string) string {
    // Case 1: No path specified
    if inputPath == "" {
        return filepath.Join(notebookRoot, slugifiedTitle + ".md")
    }
    
    // Case 2: Ends with "/" - explicit folder
    if strings.HasSuffix(inputPath, "/") {
        return filepath.Join(notebookRoot, inputPath, slugifiedTitle + ".md")
    }
    
    // Case 3: Existing directory - implicit folder
    fullPath := filepath.Join(notebookRoot, inputPath)
    if fileInfo, err := os.Stat(fullPath); err == nil && fileInfo.IsDir() {
        return filepath.Join(fullPath, slugifiedTitle + ".md")
    }
    
    // Case 4: Full filepath with .md extension
    if strings.HasSuffix(inputPath, ".md") {
        return filepath.Join(notebookRoot, inputPath)
    }
    
    // Case 5: Filepath without extension - auto-add .md
    return filepath.Join(notebookRoot, inputPath + ".md")
}
```

**Title Slugification**:
- Convert title to lowercase
- Replace spaces with hyphens
- Remove special characters (keep alphanumeric and hyphens)
- Examples:
  - `"Quick Thought"` → `quick-thought.md`
  - `"Meeting Notes"` → `meeting-notes.md`
  - `"Bug #456"` → `bug-456.md`

## Examples

### Example 1: Minimal (New Style)
```bash
opennotes notes add "Quick Thought"
```

**Result**:
- File: `quick-thought.md` (in notebook root)
- Frontmatter:
  ```yaml
  ---
  created: 2026-01-20T21:14:00+10:30
  title: Quick Thought
  ---
  ```
- Content: `# Quick Thought\n\n`

### Example 2: With Folder Path
```bash
opennotes notes add "Meeting Notes" meetings/
```

**Result**:
- File: `meetings/meeting-notes.md`
- Frontmatter:
  ```yaml
  ---
  created: 2026-01-20T21:14:00+10:30
  title: Meeting Notes
  ---
  ```
- Content: `# Meeting Notes\n\n`

### Example 3: With Full Filepath
```bash
opennotes notes add "Meeting Notes" meetings/2024-01-20.md
```

**Result**:
- File: `meetings/2024-01-20.md` (exact path used)
- Frontmatter:
  ```yaml
  ---
  created: 2026-01-20T21:14:00+10:30
  title: Meeting Notes
  ---
  ```
- Content: `# Meeting Notes\n\n`

### Example 4: With Metadata
```bash
opennotes notes add "Sprint Planning" meetings/ \
  --data tag=meeting \
  --data tag=sprint \
  --data attendees=team \
  --data priority=high
```

**Result**:
- File: `meetings/sprint-planning.md`
- Frontmatter:
  ```yaml
  ---
  created: 2026-01-20T21:14:00+10:30
  title: Sprint Planning
  tag:
    - meeting
    - sprint
  attendees: team
  priority: high
  ---
  ```
- Content: `# Sprint Planning\n\n`

### Example 5: With Stdin (Highest Priority)
```bash
echo "## Agenda\n- Item 1\n- Item 2" | \
  opennotes notes add "Sprint Planning" meetings/2024-01-20.md \
  --data tag=meeting
```

**Result**:
- File: `meetings/2024-01-20.md`
- Frontmatter:
  ```yaml
  ---
  created: 2026-01-20T21:14:00+10:30
  title: Sprint Planning
  tag: meeting
  ---
  ```
- Content: (from stdin)
  ```markdown
  ## Agenda
  - Item 1
  - Item 2
  ```

### Example 6: With Template (No Stdin)
```bash
opennotes notes add "Bug #456" bugs/ \
  --template bug \
  --data priority=high \
  --data assignee=alice
```

**Result**:
- File: `bugs/bug-456.md`
- Frontmatter:
  ```yaml
  ---
  created: 2026-01-20T21:14:00+10:30
  title: Bug #456
  priority: high
  assignee: alice
  ---
  ```
- Content: (from `bug` template, e.g.)
  ```markdown
  # Bug #456
  
  ## Description
  
  ## Steps to Reproduce
  
  ## Expected Behavior
  
  ## Actual Behavior
  ```

### Example 7: Old Style with Warning (Deprecated)
```bash
opennotes notes add --title "My Note"
```

**Output**:
```
⚠️  Warning: --title flag is deprecated, use positional argument instead. Will be removed in v2.0.0
```

**Result**:
- File: `my-note.md`
- Frontmatter and content as expected
- Warning shown to stderr

### Example 8: Old Style with Path (Deprecated)
```bash
opennotes notes add my-file.md --title "My Note"
```

**Output**:
```
⚠️  Warning: --title flag is deprecated, use positional argument instead. Will be removed in v2.0.0
```

**Result**:
- File: `my-file.md`
- Warning shown to stderr

### Example 9: Error Cases

**Both positional and flag title**:
```bash
opennotes notes add "Title" --title "Other Title"
```
**Output**: `Error: cannot specify title both as argument and --title flag`

**No title provided**:
```bash
opennotes notes add
```
**Output**: `Error: title is required (provide as first argument or --title flag)`

**Deprecated data.title usage**:
```bash
opennotes notes add "My Note" --data title="Override Title"
```
**Output**:
```
⚠️  Warning: title field is deprecated in --data, use positional argument. Will be removed in v2.0.0
```

## Implementation Changes

### Files to Modify

| File | Type | Changes Required |
|------|------|------------------|
| `cmd/notes_add.go` | Command | Core implementation changes |
| `internal/services/note.go` | Service | Path resolution, content generation |
| `internal/services/note_test.go` | Tests | Test coverage for new behavior |

### Command Implementation (`cmd/notes_add.go`)

**Current Structure**:
```go
var notesAddCmd = &cobra.Command{
    Use:   "add",
    Short: "Add a new note to the notebook",
    Args:  cobra.ExactArgs(0), // Current: no positional args
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
    },
}

func init() {
    notesAddCmd.Flags().StringVar(&titleFlag, "title", "", "Note title")
    notesAddCmd.Flags().StringVar(&templateFlag, "template", "", "Template name")
    notesAddCmd.MarkFlagRequired("title")
}
```

**New Structure** (v1.x with backward compatibility):
```go
var notesAddCmd = &cobra.Command{
    Use:   "add <title> [path]",
    Short: "Add a new note to the notebook",
    Long: `Add a new note to the notebook with optional metadata.

SYNTAX:
  opennotes notes add <title> [path] [flags]          # New style (recommended)
  opennotes notes add [path] --title "Title" [flags]  # Old style (deprecated)

EXAMPLES:
  # Create note in root
  opennotes notes add "Quick Thought"
  
  # Create note in folder
  opennotes notes add "Meeting Notes" meetings/
  
  # Create note with metadata
  opennotes notes add "Sprint Planning" meetings/ \
    --data tag=meeting --data priority=high
  
  # Pipe content from stdin
  echo "# Content" | opennotes notes add "My Note"
`,
    Args:  cobra.MaximumNArgs(2), // Changed: allow 0-2 args for backward compat
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Resolve title (positional vs flag)
        title, path, err := parseArguments(args, titleFlag)
        if err != nil {
            return err
        }
        
        // 2. Show deprecation warning if --title used
        if titleFlag != "" {
            Log.Warn("--title flag is deprecated, use positional argument instead. Will be removed in v2.0.0")
        }
        
        // 3. Check for stdin content
        stdinContent, err := readStdin()
        if err != nil {
            return fmt.Errorf("reading stdin: %w", err)
        }
        
        // 4. Parse --data flags
        frontmatterData, err := parseDataFlags(dataFlags)
        if err != nil {
            return fmt.Errorf("parsing --data flags: %w", err)
        }
        
        // 5. Generate content (stdin > template > default)
        content := generateContent(title, stdinContent, templateFlag)
        
        // 6. Resolve path (auto-detect file vs folder)
        notebook := getCurrentNotebook()
        finalPath := resolvePath(notebook.Path, path, slugify(title))
        
        // 7. Create note with frontmatter
        err = createNoteWithFrontmatter(finalPath, title, content, frontmatterData)
        if err != nil {
            return fmt.Errorf("creating note: %w", err)
        }
        
        fmt.Printf("Created note: %s\n", finalPath)
        return nil
    },
}

func init() {
    // Keep --title flag for backward compatibility (v1.x only)
    notesAddCmd.Flags().StringVar(&titleFlag, "title", "", "Note title (DEPRECATED: use positional argument)")
    notesAddCmd.Flags().MarkDeprecated("title", "use positional argument instead, will be removed in v2.0.0")
    
    // Existing template flag
    notesAddCmd.Flags().StringVar(&templateFlag, "template", "", "Template name from config")
    
    // NEW: --data flag (repeatable)
    notesAddCmd.Flags().StringArrayVar(&dataFlags, "data", []string{}, "Set frontmatter field (repeatable, format: field=value)")
}
```

### Service Implementation

**New Functions Required in `internal/services/note.go`**:

```go
// parseDataFlags parses --data flags in "field=value" format
func parseDataFlags(dataFlags []string) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    
    for _, dataFlag := range dataFlags {
        parts := strings.SplitN(dataFlag, "=", 2)
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid --data format: %s (expected field=value)", dataFlag)
        }
        
        field := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        
        // Warn about deprecated title usage
        if field == "title" {
            Log.Warn("title field is deprecated in --data, use positional argument. Will be removed in v2.0.0")
        }
        
        // Support multiple values (convert to array)
        if existing, ok := result[field]; ok {
            switch v := existing.(type) {
            case []interface{}:
                result[field] = append(v, value)
            default:
                result[field] = []interface{}{v, value}
            }
        } else {
            result[field] = value
        }
    }
    
    return result, nil
}

// readStdin reads content from stdin if available
func readStdin() (string, error) {
    stat, err := os.Stdin.Stat()
    if err != nil {
        return "", err
    }
    
    // Check if stdin is piped
    if (stat.Mode() & os.ModeCharDevice) != 0 {
        return "", nil // No stdin
    }
    
    bytes, err := io.ReadAll(os.Stdin)
    if err != nil {
        return "", err
    }
    
    return string(bytes), nil
}

// slugify converts title to filesystem-safe name
func slugify(title string) string {
    // Convert to lowercase
    slug := strings.ToLower(title)
    
    // Replace spaces with hyphens
    slug = strings.ReplaceAll(slug, " ", "-")
    
    // Remove special characters (keep alphanumeric and hyphens)
    var builder strings.Builder
    for _, r := range slug {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
            builder.WriteRune(r)
        }
    }
    
    return builder.String()
}

// createNoteWithFrontmatter creates note with frontmatter and content
func createNoteWithFrontmatter(path, title, content string, customData map[string]interface{}) error {
    // Generate frontmatter
    frontmatter := map[string]interface{}{
        "created": time.Now().Format(time.RFC3339),
        "title":   title,
    }
    
    // Merge custom data
    for k, v := range customData {
        frontmatter[k] = v
    }
    
    // Serialize frontmatter to YAML
    fmBytes, err := yaml.Marshal(frontmatter)
    if err != nil {
        return fmt.Errorf("marshaling frontmatter: %w", err)
    }
    
    // Construct final content
    finalContent := fmt.Sprintf("---\n%s---\n\n%s", string(fmBytes), content)
    
    // Ensure directory exists
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("creating directory: %w", err)
    }
    
    // Write file
    if err := os.WriteFile(path, []byte(finalContent), 0644); err != nil {
        return fmt.Errorf("writing file: %w", err)
    }
    
    return nil
}
```

### Testing Requirements

**Test Coverage Goals**: ≥85% for all new functionality

**Test Cases Required**:

1. **Title Resolution Tests**:
   - ✅ Positional title only (new style)
   - ✅ Flag title only (old style with warning)
   - ✅ Both positional and flag title (error)
   - ✅ Neither positional nor flag title (error)

2. **Argument Interpretation Tests**:
   - ✅ New style: title only
   - ✅ New style: title + path
   - ✅ Old style: title flag only
   - ✅ Old style: title flag + path arg

3. **Path Resolution Tests**:
   - ✅ No path (default to root + slugified title)
   - ✅ Folder path ending with `/`
   - ✅ Existing directory path
   - ✅ Full filepath with `.md`
   - ✅ Filepath without extension
   - ✅ Nested directories (auto-create)

4. **Content Priority Tests**:
   - ✅ Stdin content overrides template
   - ✅ Template used when stdin empty
   - ✅ Default content when no stdin or template

5. **Frontmatter Tests**:
   - ✅ Auto-generated `created` field
   - ✅ Title from positional arg
   - ✅ Single custom field via `--data`
   - ✅ Multiple custom fields via `--data`
   - ✅ Repeated `--data` field (array creation)
   - ✅ Deprecated `--data title=...` (warning)

6. **Data Flag Parsing Tests**:
   - ✅ Valid format `field=value`
   - ✅ Invalid format (error)
   - ✅ Multiple values for same field
   - ✅ Special characters in values

7. **Slugification Tests**:
   - ✅ Spaces to hyphens
   - ✅ Special characters removed
   - ✅ Lowercase conversion
   - ✅ Numbers preserved

8. **Deprecation Warning Tests**:
   - ✅ Warning shown for `--title` flag
   - ✅ Warning shown for `--data title=...`
   - ✅ No warning for new style

9. **Integration Tests**:
   - ✅ End-to-end: create note with all features
   - ✅ Stdin piping workflow
   - ✅ Template + metadata workflow
   - ✅ Cross-platform path handling

## Migration Timeline

### Phase 1: v1.x Releases (Current)

**Status**: Dual syntax support with deprecation warnings

**Behavior**:
- ✅ Both old and new syntax work
- ⚠️ `--title` flag shows deprecation warning
- ⚠️ `--data title=...` shows deprecation warning
- ✅ All features available

**User Communication**:
- Release notes document new syntax
- Deprecation warnings guide users
- Documentation updated to show new style first
- Migration guide provided

**Testing Strategy**:
- Test both syntaxes in CI
- Validate deprecation warnings
- Ensure zero regressions

### Phase 2: v2.0.0 Release (Future)

**Status**: Breaking change - old syntax removed

**Behavior**:
- ❌ `--title` flag removed entirely
- ✅ Only positional title syntax supported
- ❌ `--data title=...` becomes error (not just warning)

**User Communication**:
- Advance notice in v1.x release notes
- Migration guide prominently featured
- Clear error messages for removed flags

**Migration Path**:
```bash
# Old (v1.x)
opennotes notes add --title "My Note"

# New (v2.0.0+)
opennotes notes add "My Note"
```

## Breaking Changes Summary

### v1.x → v2.0.0

| Feature | v1.x Behavior | v2.0.0 Behavior |
|---------|---------------|-----------------|
| `--title` flag | Works with warning | Removed (error) |
| Positional title | Works (recommended) | Works (required) |
| `--data title=...` | Works with warning | Error |
| Path resolution | Same | Same |
| `--data` flags | Same | Same |
| Template support | Same | Same |
| Stdin priority | Same | Same |

**Migration Effort**: LOW - Simple command syntax change
**Tooling Impact**: MEDIUM - Scripts using `--title` must be updated
**Risk**: LOW - Deprecation warnings provide advance notice

## Error Handling

### Error Messages

**Missing Title**:
```
Error: title is required (provide as first argument or --title flag)
Usage: opennotes notes add <title> [path] [flags]
```

**Conflicting Title Arguments**:
```
Error: cannot specify title both as argument and --title flag
Usage: opennotes notes add <title> [path] [flags]
```

**Invalid Data Flag Format**:
```
Error: invalid --data format: "fieldvalue" (expected field=value)
```

**File Already Exists**:
```
Error: note already exists at path/to/note.md
Use --force to overwrite (future flag)
```

**Directory Creation Failed**:
```
Error: creating directory: permission denied
Path: /path/to/directory
```

### Validation Sequence

1. **Pre-flight Checks**:
   - ✅ Notebook context resolved
   - ✅ Title provided (positional or flag)
   - ✅ No conflicting title arguments

2. **Flag Validation**:
   - ✅ `--data` flags in valid format
   - ✅ Template name exists (if specified)

3. **Path Validation**:
   - ✅ Target directory is writable
   - ✅ File doesn't already exist (unless --force)

4. **Content Validation**:
   - ✅ Stdin content is valid (if provided)
   - ✅ Template content is valid (if used)

5. **Write Operation**:
   - ✅ Directory creation successful
   - ✅ File write successful
   - ✅ File permissions correct

## Performance Considerations

### Performance Goals

| Operation | Target | Notes |
|-----------|--------|-------|
| Command parsing | <5ms | Cobra flag parsing |
| Stdin reading | <10ms | For typical note content |
| Path resolution | <1ms | Filesystem checks |
| Content generation | <5ms | Template or default |
| Frontmatter serialization | <5ms | YAML marshaling |
| File write | <10ms | Typical SSD performance |
| **Total** | **<50ms** | End-to-end command execution |

### Optimization Notes

- **Lazy Loading**: Only read stdin when needed
- **Minimal Filesystem**: Single directory check for path resolution
- **Efficient Slugification**: Single-pass string transformation
- **YAML Library**: Use fast YAML serializer (e.g., `gopkg.in/yaml.v3`)

## Security Considerations

### Input Validation

1. **Path Traversal Protection**:
   - ✅ Validate paths are within notebook root
   - ✅ Reject `..` path components
   - ✅ Canonicalize paths before writing

2. **Filename Sanitization**:
   - ✅ Slugification removes dangerous characters
   - ✅ Limit filename length (max 255 characters)
   - ✅ Reject hidden files (starting with `.`)

3. **Frontmatter Injection Protection**:
   - ✅ YAML library handles escaping
   - ✅ No raw string interpolation
   - ✅ Validate field names (alphanumeric + underscore)

4. **Stdin Content**:
   - ✅ No size limit (user responsibility)
   - ✅ Content treated as untrusted markdown
   - ✅ No code execution during note creation

### Recommended Validations

```go
// validatePath ensures path is safe and within notebook
func validatePath(notebookRoot, targetPath string) error {
    // Canonicalize paths
    absNotebook, err := filepath.Abs(notebookRoot)
    if err != nil {
        return err
    }
    
    absTarget, err := filepath.Abs(targetPath)
    if err != nil {
        return err
    }
    
    // Ensure target is within notebook
    if !strings.HasPrefix(absTarget, absNotebook) {
        return errors.New("path must be within notebook root")
    }
    
    // Reject hidden files
    if strings.HasPrefix(filepath.Base(absTarget), ".") {
        return errors.New("hidden files not allowed")
    }
    
    return nil
}

// validateFieldName ensures frontmatter field names are safe
func validateFieldName(field string) error {
    if field == "" {
        return errors.New("field name cannot be empty")
    }
    
    // Allow only alphanumeric and underscore
    for _, r := range field {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
            return fmt.Errorf("invalid field name: %s (use alphanumeric and underscore only)", field)
        }
    }
    
    return nil
}
```

## Documentation Updates

### Files to Update

| File | Section | Changes |
|------|---------|---------|
| `README.md` | Quick Start | Update syntax to new style |
| `docs/cli-reference.md` | `notes add` | Full command documentation |
| CLI Help | `--help` output | Updated examples and flag descriptions |
| `docs/migration-guide.md` | **NEW** | Migration from old to new syntax |

### Documentation Checklist

- [ ] Update README with new syntax examples
- [ ] Create migration guide for v1.x → v2.0.0
- [ ] Document `--data` flag with multiple examples
- [ ] Document stdin piping workflows
- [ ] Add path resolution examples
- [ ] Document deprecation timeline
- [ ] Update CLI help text
- [ ] Add troubleshooting section for common errors

## Acceptance Criteria

### Functional Requirements

- [ ] Positional title argument works as expected
- [ ] `--title` flag still works with deprecation warning (v1.x)
- [ ] Both positional and flag title causes error
- [ ] Path resolution auto-detects files vs folders
- [ ] Stdin content overrides template content
- [ ] `--data` flags set frontmatter fields correctly
- [ ] Repeated `--data` fields create arrays
- [ ] Title slugification works for all test cases
- [ ] Deprecation warnings shown at correct times

### Non-Functional Requirements

- [ ] Command executes in <50ms for typical use
- [ ] Test coverage ≥85% for new code
- [ ] No regressions in existing functionality
- [ ] All error messages are clear and actionable
- [ ] Documentation is comprehensive and accurate
- [ ] Cross-platform compatibility (Linux, macOS, Windows)

### Quality Gates

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Linter passes with zero warnings
- [ ] Manual testing on 3 platforms
- [ ] Code review approved
- [ ] Documentation review approved

## Future Enhancements (Out of Scope)

These features are NOT included in this specification but may be considered for future epics:

1. **Force Overwrite Flag**: `--force` to overwrite existing notes
2. **Interactive Mode**: Prompt for metadata fields interactively
3. **Batch Creation**: Create multiple notes from a single command
4. **Auto-tagging**: Automatically infer tags from title or content
5. **Smart Templates**: Template selection based on path or metadata
6. **Validation Hooks**: Run custom validation before note creation
7. **Git Integration**: Auto-commit created notes
8. **Editor Launch**: Automatically open created note in editor

## References

### Related Documents

- Epic: `.memory/epic-3e01c563-advanced-note-operations.md`
- Research: `.memory/research-3e01c563-advanced-operations.md`
- Architecture: `.memory/knowledge-codemap.md`
- Learning: `.memory/learning-5e4c3f2a-codebase-architecture.md`

### External References

- Cobra Documentation: https://github.com/spf13/cobra
- YAML v3: https://github.com/go-yaml/yaml
- Go CLI Best Practices: https://cli.guide/
- Semantic Versioning: https://semver.org/

### Similar Implementations

- **kubectl create**: Positional resource type, flag-based configuration
- **docker run**: Positional image name, flag-based options
- **gh issue create**: Interactive prompts with flag overrides
- **taskwarrior add**: Positional description, flag-based attributes

## Appendix A: Command Structure Comparison

### Before (Current)
```bash
# Only one syntax supported
opennotes notes add --title "My Note"
opennotes notes add --title "My Note" --template meeting

# Verbose, flag-heavy
```

### After (v1.x)
```bash
# New style (recommended)
opennotes notes add "My Note"
opennotes notes add "My Note" meetings/
opennotes notes add "My Note" --template meeting --data tag=important

# Old style (deprecated, shows warning)
opennotes notes add --title "My Note"
```

### Future (v2.0.0)
```bash
# Only new style supported
opennotes notes add "My Note"
opennotes notes add "My Note" meetings/
opennotes notes add "My Note" --template meeting --data tag=important

# Old style removed (error)
```

## Appendix B: Implementation Checklist

### Pre-Implementation

- [ ] Review specification with stakeholders
- [ ] Validate design decisions
- [ ] Confirm test coverage goals
- [ ] Set up feature branch

### Implementation Tasks

- [ ] Update `cmd/notes_add.go` command structure
- [ ] Implement title resolution logic
- [ ] Implement argument interpretation logic
- [ ] Implement path resolution algorithm
- [ ] Implement content priority logic
- [ ] Implement frontmatter generation
- [ ] Implement `--data` flag parsing
- [ ] Implement stdin reading
- [ ] Implement slugification
- [ ] Add deprecation warnings
- [ ] Add error handling

### Testing Tasks

- [ ] Write unit tests for title resolution
- [ ] Write unit tests for path resolution
- [ ] Write unit tests for data flag parsing
- [ ] Write unit tests for slugification
- [ ] Write integration tests for full workflow
- [ ] Test deprecation warnings
- [ ] Test error messages
- [ ] Test cross-platform behavior

### Documentation Tasks

- [ ] Update CLI help text
- [ ] Update README examples
- [ ] Create migration guide
- [ ] Document `--data` flag usage
- [ ] Add troubleshooting section
- [ ] Update architecture docs

### Quality Assurance

- [ ] Code review
- [ ] Documentation review
- [ ] Manual testing on Linux
- [ ] Manual testing on macOS
- [ ] Manual testing on Windows
- [ ] Performance benchmarking
- [ ] Security review

### Release Preparation

- [ ] Update CHANGELOG
- [ ] Tag release notes
- [ ] Prepare migration announcement
- [ ] Update version number
- [ ] Create release branch

---

**End of Specification**

**Specification Version**: 1.0  
**Created**: 2026-01-20T21:14:00+10:30  
**Status**: Proposed - Awaiting Implementation Planning  
**Next Steps**: Human review → Implementation planning → Task breakdown
