/**
 * Common interface for database connections and operations.
 * This interface abstracts the underlying database implementation,
 * allowing for different backends (WASM, Native, etc).
 */

/**
 * Represents a prepared DuckDB statement that can be parameterized and executed.
 */
export interface DbPreparedStatement {
  /**
   * Send the prepared statement with optional parameters.
   * @param params - Optional parameter bindings for the statement
   * @returns A request object that can be used to read results
   */
  send(params?: Record<string, unknown>): Promise<DbStatementRequest>;
}

/**
 * Represents the result of a sent statement, allowing iterative or bulk reading.
 */
export interface DbStatementRequest {
  /**
   * Read all available rows from the result set.
   * @returns Array of row objects, or null if no results
   */
  readAll(): Promise<unknown[] | null>;
}

/**
 * Represents an active database connection with query execution capabilities.
 */
export interface DbConnection {
  /**
   * Prepare a SQL statement for execution with optional parameters.
   * @param query - The SQL query string
   * @returns A prepared statement that can be executed
   */
  prepare(query: string): Promise<DbPreparedStatement>;

  /**
   * Execute a raw query (convenience method, equivalent to prepare + send + readAll).
   * @param query - The SQL query string
   * @returns Array of result rows, or null if no results
   */
  query(query: string): Promise<unknown[] | null>;
}

/**
 * Database service that manages connection lifecycle and instance management.
 */
export interface IDbService {
  /**
   * Get or create a database connection.
   * Implementations may reuse connections or create new ones as needed.
   * @returns A ready-to-use database connection
   */
  getDb(): Promise<DbConnection>;
}
