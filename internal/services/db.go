package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	// DuckDB driver
	_ "github.com/duckdb/duckdb-go/v2"
)

// DbService manages DuckDB database connections.
type DbService struct {
	db   *sql.DB
	once sync.Once
	mu   sync.Mutex
	log  zerolog.Logger
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
	defer rows.Close()

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

// Close closes the database connection.
func (d *DbService) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		d.log.Debug().Msg("closing database")
		return d.db.Close()
	}
	return nil
}
