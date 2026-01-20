# md-process-blocks

A Go CLI tool for processing markdown code blocks with custom processors.

## Features

- **Simple key=value syntax**: `exec=true`, `exec="command"`, `replace=true`
- **Stdin/stdout support**: Process from pipes or files
- **Flexible execution modes**: Run block content OR use block as stdin
- **Pure Go**: No external dependencies except your tools (like d2)

## Usage

```bash
# Process file
md-process-blocks -i input.md -o output.md

# Process from stdin
cat input.md | md-process-blocks > output.md
```

## Syntax

### Mode 1: Execute block content as command, append output

```markdown
˙˙˙bash exec=true
echo "Hello World"
˙˙˙
```

Output:
```markdown
˙˙˙bash
echo "Hello World"
˙˙˙

˙˙˙
Hello World
˙˙˙
```

### Mode 2: Execute block content as command, replace block

```markdown
˙˙˙bash exec=true replace=true
echo "Hello World"
˙˙˙
```

Output:
```markdown
Hello World
```

### Mode 3: Execute command with block as stdin, replace block

```markdown
˙˙˙d2 exec="d2 - -" replace=true
x -> y: Hello
˙˙˙
```

Output:
```markdown
<svg xmlns="http://www.w3.org/2000/svg"...>
</svg>
```

### Mode 4: No exec attribute (unchanged)

```markdown
˙˙˙python
print("unchanged")
˙˙˙
```

Output (unchanged):
```markdown
˙˙˙python
print("unchanged")
˙˙˙
```

## Attributes

| Attribute | Values | Description |
|-----------|--------|-------------|
| `exec` | `true` | Execute block content as shell command |
| `exec` | `"command args"` | Execute command with block content as stdin |
| `replace` | `true` | Replace block with output (default: append below) |

## Integration

Used by opennotes documentation build system:
- Source: `pkgs/docs/*.md` (markdown with D2 blocks)
- Output: `docs/*.md` (markdown with embedded SVG)
- Task: `mise run docs-build`

## Building

```bash
go build -o ../../dist/md-process-blocks
```

## Testing

```bash
# Run all tests (unit + e2e)
go test -v

# Run only unit tests
go test -v -run '^Test[^E]'

# Run only e2e tests
go test -v -run '^TestE2E'
```

Test coverage:
- ✅ Code block extraction and parsing
- ✅ Quoted value handling (`exec="command with spaces"`)
- ✅ Multiple attribute parsing
- ✅ Execution modes (exec=true, exec="command", replace=true/false)
- ✅ Error handling (failed commands, missing commands)
- ✅ File I/O operations
- ✅ Multiple blocks in one document
- ✅ Real-world D2 integration (when d2 available)
