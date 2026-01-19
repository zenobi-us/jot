package services

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	// DuckDB driver
	_ "github.com/duckdb/duckdb-go/v2"
)

// Compiled regex patterns for glob detection
var (
	globPatternRegex   *regexp.Regexp
	readMarkdownRegex  *regexp.Regexp
)

func init() {
	// Match quoted strings containing glob patterns (* or ?)
	globPatternRegex = regexp.MustCompile(`(['"])(.*[\*\?].*?)(['"])`)
	
	// Match read_markdown function calls with file paths
	readMarkdownRegex = regexp.MustCompile(`read_markdown\s*\(\s*(['"])(.*?)(['"])`)
}

// DbService manages DuckDB database connections.
type DbService struct {
	db       *sql.DB
	readOnly *sql.DB
	once     sync.Once
	roOnce   sync.Once
	mu       sync.Mutex
	log      zerolog.Logger
}

// NewDbService creates a new database service.
func NewDbService() *DbService {
	return &DbService{
		log: Log("DbService"),
	}
}

// GetDB returns an initialized database connection.
// The connection is lazily initialized on first call and reused thereafter.
func (d *DbService) GetDB(ctx context.Context) (*sql.DB, error) {
	var initErr error

	d.once.Do(func() {
		d.log.Debug().Msg("initializing database")

		// Open in-memory database
		db, err := sql.Open("duckdb", "")
		if err != nil {
			initErr = fmt.Errorf("failed to open database: %w", err)
			return
		}
		d.db = db

		// Install and load markdown extension
		d.log.Debug().Msg("installing markdown extension")
		if _, err := db.ExecContext(ctx, "INSTALL markdown FROM community"); err != nil {
			initErr = fmt.Errorf("failed to install markdown extension: %w", err)
			return
		}

		d.log.Debug().Msg("loading markdown extension")
		if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
			initErr = fmt.Errorf("failed to load markdown extension: %w", err)
			return
		}

		d.log.Debug().Msg("database initialized")
	})

	if initErr != nil {
		return nil, initErr
	}

	return d.db, nil
}

// GetReadOnlyDB returns a separate read-only database connection.
// This is used for executing user-provided SQL queries safely.
// The connection is lazily initialized on first call and reused thereafter.
func (d *DbService) GetReadOnlyDB(ctx context.Context) (*sql.DB, error) {
	var initErr error

	d.roOnce.Do(func() {
		d.log.Debug().Msg("initializing read-only database connection")

		// Open separate in-memory database
		db, err := sql.Open("duckdb", "")
		if err != nil {
			initErr = fmt.Errorf("failed to open read-only database: %w", err)
			return
		}

		// Install and load markdown extension
		d.log.Debug().Msg("installing markdown extension on read-only connection")
		if _, err := db.ExecContext(ctx, "INSTALL markdown FROM community"); err != nil {
			initErr = fmt.Errorf("failed to install markdown extension on read-only connection: %w", err)
			if closeErr := db.Close(); closeErr != nil {
				d.log.Warn().Err(closeErr).Msg("failed to close db after install error")
			}
			return
		}

		d.log.Debug().Msg("loading markdown extension on read-only connection")
		if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
			initErr = fmt.Errorf("failed to load markdown extension on read-only connection: %w", err)
			if closeErr := db.Close(); closeErr != nil {
				d.log.Warn().Err(closeErr).Msg("failed to close db after load error")
			}
			return
		}

		d.readOnly = db
		d.log.Debug().Msg("read-only database initialized")
	})

	if initErr != nil {
		return nil, initErr
	}

	return d.readOnly, nil
}

// Query executes a query and returns results as maps.
func (d *DbService) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	db, err := d.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			d.log.Warn().Err(err).Msg("failed to close rows")
		}
	}()

	return rowsToMaps(rows)
}

// rowsToMaps converts sql.Rows to a slice of maps.
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// preprocessSQL processes SQL query to resolve glob patterns relative to notebook root.
// This ensures that patterns like "**/*.md" resolve consistently from the notebook root
// directory instead of the current working directory.
func (d *DbService) preprocessSQL(query string, notebookRoot string) (string, error) {
	d.log.Debug().
		Str("originalQuery", query).
		Str("notebookRoot", notebookRoot).
		Msg("preprocessing SQL query")

	// Keep track of any errors during replacement
	var lastErr error

	// First, validate all read_markdown file paths for security
	readMarkdownRegex.ReplaceAllStringFunc(query, func(match string) string {
		// Extract file path from read_markdown('path')
		submatches := readMarkdownRegex.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			filePath := submatches[2]
			cleanPath := filepath.Clean(filePath)
			
			// Check for path traversal patterns
			if strings.Contains(cleanPath, "..") {
				lastErr = fmt.Errorf("path traversal not allowed in file path: %s", filePath)
				return match
			}
			
			// Check for absolute paths (should be relative to notebook)
			if filepath.IsAbs(cleanPath) {
				lastErr = fmt.Errorf("path traversal not allowed in file path: %s", filePath)
				return match
			}
		}
		return match
	})

	// Return early if path traversal detected
	if lastErr != nil {
		return "", fmt.Errorf("SQL preprocessing failed: %w", lastErr)
	}

	// Then process glob patterns normally
	processed := globPatternRegex.ReplaceAllStringFunc(query, func(match string) string {
		// Extract the pattern from the quoted string
		// The regex captures: quote + pattern + quote
		if len(match) < 2 {
			return match
		}

		quote := match[0:1]        // First character (quote)
		pattern := match[1 : len(match)-1] // Everything except first and last char
		endQuote := match[len(match)-1:]   // Last character (quote)

		// Only process if quotes match
		if quote != endQuote {
			return match
		}

		// Resolve pattern to absolute path
		resolvedPath, err := d.resolveGlobPattern(pattern, notebookRoot)
		if err != nil {
			d.log.Warn().
				Err(err).
				Str("pattern", pattern).
				Msg("failed to resolve glob pattern")
			lastErr = err
			return match // Return original on error
		}

		// Validate path is within notebook directory
		if err := d.validateNotebookPath(resolvedPath, notebookRoot); err != nil {
			d.log.Warn().
				Err(err).
				Str("resolvedPath", resolvedPath).
				Str("pattern", pattern).
				Msg("security validation failed for glob pattern")
			lastErr = err
			return match // Return original on security failure, but remember the error
		}

		result := quote + resolvedPath + endQuote
		d.log.Debug().
			Str("pattern", pattern).
			Str("resolvedPath", resolvedPath).
			Msg("glob pattern resolved")

		return result
	})

	// If any security validation failed, return error
	if lastErr != nil {
		return "", fmt.Errorf("SQL preprocessing failed: %w", lastErr)
	}

	if processed != query {
		d.log.Debug().
			Str("processedQuery", processed).
			Msg("query preprocessing completed")
	}

	return processed, nil
}

// resolveGlobPattern converts a relative glob pattern to an absolute path
// anchored at the notebook root directory.
func (d *DbService) resolveGlobPattern(pattern string, notebookRoot string) (string, error) {
	// Clean the pattern to handle any path traversal attempts
	cleanPattern := filepath.Clean(pattern)

	// Check for path traversal attempts
	if strings.Contains(cleanPattern, "..") {
		return "", fmt.Errorf("path traversal not allowed in glob pattern: %s", pattern)
	}

	// If pattern is already absolute, validate it's within notebook
	if filepath.IsAbs(cleanPattern) {
		return cleanPattern, nil
	}

	// Convert relative pattern to absolute path
	absolutePath := filepath.Join(notebookRoot, cleanPattern)

	return absolutePath, nil
}

// validateNotebookPath ensures the resolved path stays within the notebook directory.
// This prevents path traversal attacks that could access files outside the notebook.
func (d *DbService) validateNotebookPath(resolvedPath, notebookRoot string) error {
	// Get absolute paths for comparison
	absResolved, err := filepath.Abs(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for resolved pattern: %w", err)
	}

	absNotebook, err := filepath.Abs(notebookRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for notebook root: %w", err)
	}

	// Ensure resolved path starts with notebook root
	// Use filepath.Clean to normalize paths before comparison
	cleanResolved := filepath.Clean(absResolved)
	cleanNotebook := filepath.Clean(absNotebook)

	// Check if the resolved path is within the notebook directory
	if !strings.HasPrefix(cleanResolved, cleanNotebook) {
		return fmt.Errorf("path traversal detected: resolved path %s is outside notebook directory %s", cleanResolved, cleanNotebook)
	}

	return nil
}

// Close closes both database connections.
func (d *DbService) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var errs []error

	if d.db != nil {
		d.log.Debug().Msg("closing main database")
		if err := d.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if d.readOnly != nil {
		d.log.Debug().Msg("closing read-only database")
		if err := d.readOnly.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close database(s): %v", errs)
	}

	return nil
}
