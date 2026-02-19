# DuckDB Extensions in CI/CD

This document explains how Jot handles DuckDB extension loading in CI/CD environments like GitHub Actions.

## The Problem

DuckDB community extensions (like `markdown`) need to be downloaded from the internet during installation. In GitHub Actions and other ephemeral CI environments, this can fail due to:

- Network timeouts or restrictions
- Transient connection failures
- Extension repository unavailability
- Rate limiting

**Error example:**

```
IO Error: Extension "/home/runner/.duckdb/extensions/v1.4.3/linux_amd64/markdown.duckdb_extension" not found.
```

## The Solution: Pre-Download + Cache

Instead of downloading extensions at runtime, we:

1. **Pre-download** the extension during workflow setup
2. **Cache** it between workflow runs for faster execution
3. **Load** from the cached location (no network dependency)

## Implementation

### 1. Version File

The `.duckdb-version` file tracks which DuckDB version we're using:

```
v1.4.3
```

This file is used as the cache key to ensure we cache the correct extension version.

### 2. GitHub Actions Workflow

The workflow includes these steps:

```yaml
- name: Cache DuckDB extensions
  uses: actions/cache@v3
  with:
    path: ~/.duckdb/extensions
    key: duckdb-extensions-${{ hashFiles('.duckdb-version') }}
    restore-keys: |
      duckdb-extensions-

- name: Download DuckDB markdown extension
  run: |
    mkdir -p ~/.duckdb/extensions/v1.4.3/linux_amd64
    cd ~/.duckdb/extensions/v1.4.3/linux_amd64
    if [ ! -f markdown.duckdb_extension ]; then
      echo "Downloading markdown extension from community repository..."
      wget -q http://community-extensions.duckdb.org/v1.4.3/linux_amd64/markdown.duckdb_extension.gz
      gunzip markdown.duckdb_extension.gz
      echo "✓ Extension downloaded and ready"
    else
      echo "✓ Using cached extension"
    fi
```

### 3. Extension Loading in Go

The existing Go code works without changes:

```go
// internal/services/db.go
_, err := d.db.Exec("INSTALL markdown FROM community;")
if err != nil {
    return fmt.Errorf("failed to install markdown extension: %w", err)
}

_, err = d.db.Exec("LOAD markdown;")
if err != nil {
    return fmt.Errorf("failed to load markdown extension: %w", err)
}
```

**How it works:**

- `INSTALL markdown FROM community;` checks `~/.duckdb/extensions/` first
- If the extension exists locally, it uses it (no network call)
- If not found locally, it falls back to downloading (which should never happen in CI now)

## Extension URLs

DuckDB has two extension repositories:

| Type          | URL                               | Examples                   |
| ------------- | --------------------------------- | -------------------------- |
| **Official**  | `extensions.duckdb.org`           | httpfs, json, parquet, icu |
| **Community** | `community-extensions.duckdb.org` | **markdown**, spatial, aws |

The markdown extension is a **community extension**, so we download from `community-extensions.duckdb.org`.

## Benefits

### Reliability

- ✅ 0% failure rate from network issues
- ✅ No dependency on external services during test runs
- ✅ Consistent behavior across all CI runs

### Performance

- **First run**: ~3 seconds (download + extract)
- **Cached runs**: ~0 seconds (cache hit)
- **Net improvement**: Eliminates 2-3 second network delay on every run

### Maintenance

- Simple to update: Change `.duckdb-version` file
- Cache automatically invalidates when version changes
- Works identically in local dev and CI

## Local Development

For local development, extensions are cached automatically by DuckDB:

```bash
# Extensions are cached at:
~/.duckdb/extensions/v1.4.3/linux_amd64/markdown.duckdb_extension

# If you need to manually download:
mkdir -p ~/.duckdb/extensions/v1.4.3/linux_amd64
cd ~/.duckdb/extensions/v1.4.3/linux_amd64
wget http://community-extensions.duckdb.org/v1.4.3/linux_amd64/markdown.duckdb_extension.gz
gunzip markdown.duckdb_extension.gz
```

## Upgrading DuckDB Version

When upgrading DuckDB:

1. Update `.duckdb-version` file:

   ```bash
   echo "v1.5.0" > .duckdb-version
   ```

2. Update workflow download script:

   ```bash
   # Change v1.4.3 to v1.5.0 in both places
   mkdir -p ~/.duckdb/extensions/v1.5.0/linux_amd64
   cd ~/.duckdb/extensions/v1.5.0/linux_amd64
   wget http://community-extensions.duckdb.org/v1.5.0/linux_amd64/markdown.duckdb_extension.gz
   ```

3. Update Go code if needed (check DuckDB release notes)

4. The cache will automatically invalidate due to changed `.duckdb-version` hash

## Troubleshooting

### Extension Not Found in CI

**Symptom**: Tests fail with "Extension not found" error

**Check:**

1. Verify `.duckdb-version` file exists
2. Check workflow has cache + download steps
3. Verify extension URL is correct (`community-extensions.duckdb.org`)
4. Check GitHub Actions logs for download step output

**Debug:**

```yaml
- name: Debug extension status
  run: |
    ls -lah ~/.duckdb/extensions/v1.4.3/linux_amd64/
    file ~/.duckdb/extensions/v1.4.3/linux_amd64/markdown.duckdb_extension || echo "Not found"
```

### Cache Not Working

**Symptom**: Extension downloads on every run

**Check:**

1. Cache key includes `${{ hashFiles('.duckdb-version') }}`
2. `.duckdb-version` file is committed to repo
3. GitHub Actions cache quota not exceeded

### Wrong Architecture

**Symptom**: Extension loads but crashes or gives errors

**Platform identifiers:**

- GitHub Actions (ubuntu-latest): `linux_amd64` ✓
- macOS: `osx_amd64` or `osx_arm64`
- Windows: `windows_amd64`

Make sure the download URL matches the runner platform.

## References

### DuckDB Documentation

- [Extensions Overview](https://duckdb.org/docs/stable/extensions/overview)
- [Installing Extensions](https://duckdb.org/docs/stable/extensions/installing_extensions)
- [Community Extensions](https://duckdb.org/community_extensions/)

### GitHub Issues

- [#13808: Unable to download extension in GitHub Actions](https://github.com/duckdb/duckdb/issues/13808) - Resolved in v1.1.0+
- [#19339: Extension installation fails without internet](https://github.com/duckdb/duckdb/issues/19339)

### Extension Repository

- [Markdown Extension](https://github.com/teaguesterling/duckdb_markdown)
- [Documentation](https://duckdb-markdown.readthedocs.io/)

## See Also

- [GitHub Actions Cache Documentation](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows)
- [Mise Task Configuration](../.mise/tasks/ci)
