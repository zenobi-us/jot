# Logging Configuration

Jot supports flexible logging configuration via environment variables.

## Environment Variables

### LOG_LEVEL

Controls which log messages are displayed based on severity.

**Values:**

- `trace` - Most verbose, includes all log levels
- `debug` - Debug information for troubleshooting
- `info` - General informational messages (default)
- `warn` - Warning messages only
- `error` - Error messages only
- `fatal` - Fatal errors only
- `panic` - Panic-level errors only

**Examples:**

```bash
# Show only errors during CI
LOG_LEVEL=error mise run ci

# Enable debug logging for troubleshooting
LOG_LEVEL=debug jot notes list

# Quiet mode - errors only
LOG_LEVEL=error jot notes search "query"
```

### LOG_FORMAT

Controls the output format of log messages.

**Values:**

#### `compact` (default)

Clean, readable format with short timestamps (HH:MM:SS).

```
22:36:45 DBG loading config namespace=ConfigService path=/home/user/.config/jot/config.json
22:36:45 DBG database initialized namespace=DbService
```

**Best for:** Interactive CLI use, development

#### `console`

Standard colorized console format with 12-hour time.

```
10:36PM DBG loading config namespace=ConfigService path=/home/user/.config/jot/config.json
10:36PM DBG database initialized namespace=DbService
```

**Best for:** Traditional console logging feel

#### `json`

Structured JSON output for log aggregation and parsing.

```json
{"level":"debug","namespace":"ConfigService","path":"/home/user/.config/jot/config.json","time":"2026-01-21T22:36:45+10:30","message":"loading config"}
{"level":"debug","namespace":"DbService","time":"2026-01-21T22:36:45+10:30","message":"database initialized"}
```

**Best for:** Log aggregation systems, parsing with `jq`, automated processing

#### `ci`

Non-colorized format with full ISO 8601 timestamps for CI/CD.

```
2026-01-21T22:36:45+10:30 DBG loading config namespace=ConfigService path=/home/user/.config/jot/config.json
2026-01-21T22:36:45+10:30 DBG database initialized namespace=DbService
```

**Best for:** CI/CD pipelines, log files, no ANSI color support

**Examples:**

```bash
# JSON output for processing with jq
LOG_FORMAT=json jot notes list 2>&1 | jq 'select(.level=="error")'

# CI-friendly output
LOG_FORMAT=ci LOG_LEVEL=info mise run ci

# Compact format (default, explicit)
LOG_FORMAT=compact jot notes search "test"
```

### DEBUG (Legacy)

Legacy flag that sets LOG_LEVEL to `debug`.

```bash
# These are equivalent:
DEBUG=1 jot notes list
LOG_LEVEL=debug jot notes list
```

**Note:** `LOG_LEVEL` takes precedence over `DEBUG` if both are set.

## Combining Options

Environment variables can be combined for fine-grained control:

```bash
# Quiet CI mode - only errors, JSON format
LOG_LEVEL=error LOG_FORMAT=json mise run ci

# Verbose debugging with full timestamps
LOG_LEVEL=debug LOG_FORMAT=ci jot notes list

# Production-friendly: info level, JSON output
LOG_LEVEL=info LOG_FORMAT=json jot notes search "incident"
```

## Usage in .mise/tasks

Tasks can set logging defaults:

```toml
[tasks.ci]
description = "Run all tests"
env = { LOG_LEVEL = "error", LOG_FORMAT = "ci" }
run = """
mise lint
mise test
mise build
"""
```

## Recommended Presets

| Scenario        | LOG_LEVEL | LOG_FORMAT | Use Case                           |
| --------------- | --------- | ---------- | ---------------------------------- |
| **Development** | `debug`   | `compact`  | Local development, troubleshooting |
| **CI/CD**       | `error`   | `ci`       | Automated testing, minimal output  |
| **Production**  | `info`    | `json`     | Log aggregation, parsing           |
| **Debugging**   | `debug`   | `console`  | Detailed troubleshooting           |
| **Silent**      | `error`   | `compact`  | Scripts, automation                |

## Examples

### Interactive Development

```bash
# Default: compact format, info level
jot notes list

# Debug mode for troubleshooting
LOG_LEVEL=debug jot notes search "query"
```

### CI/CD Pipeline

```bash
# Minimal output, only errors
LOG_LEVEL=error LOG_FORMAT=ci mise run ci
```

### Log Processing

```bash
# Extract all error messages with jq
LOG_FORMAT=json jot notes list 2>&1 | jq 'select(.level=="error")'

# Count log messages by level
LOG_FORMAT=json mise run test 2>&1 | jq -r '.level' | sort | uniq -c
```

### Production Monitoring

```bash
# JSON output for Datadog/Splunk/etc
LOG_LEVEL=info LOG_FORMAT=json jot notes search "incident" | your-log-aggregator
```

## Migration from Previous Versions

If you were using `DEBUG=1`, you can now use:

```bash
# Old way
DEBUG=1 jot notes list

# New way (more control)
LOG_LEVEL=debug LOG_FORMAT=compact jot notes list
```

The `DEBUG` variable still works but `LOG_LEVEL` provides more granular control.
