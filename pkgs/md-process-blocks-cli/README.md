# md-process-blocks

A Go CLI tool for processing markdown code blocks with custom processors.

## Features

- **mdsh-compatible syntax**: `˙˙˙lang > $ command args`
- **Stdin/stdout support**: Process from pipes or files
- **Automatic D2 diagram generation**: Built-in support for D2
- **Extensible**: Add custom processors via command metadata
- **Simple**: No Haskell dependencies, pure Go

## Usage

```bash
# Process file, replace code blocks with output
md-process-blocks -i input.md -o output.md

# Process from stdin
cat input.md | md-process-blocks > output.md

# Don't replace blocks (extract mode)
md-process-blocks -i input.md --replace=false
```

## Syntax

### mdsh-style commands

Code blocks with commands in the opening fence are processed:

```markdown
˙˙˙d2 > $ d2 - -
x -> y: Hello
˙˙˙
```

Output (code block replaced with command output):
```markdown
<svg xmlns="http://www.w3.org/2000/svg"...>
</svg>
```

### No auto-processing

Code blocks without explicit commands are left unchanged:

```markdown
˙˙˙d2
x -> y
˙˙˙
```

This is returned as-is (no processing).

## Integration

Used by opennotes documentation build system:
- Source: `pkgs/docs/*.md` (markdown with D2 blocks)
- Output: `docs/*.md` (markdown with embedded SVG)
- Task: `mise run docs-build`

## Building

```bash
go build -o ../../dist/md-process-blocks
```
