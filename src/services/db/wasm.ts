import * as duckdb from '@duckdb/duckdb-wasm';
import { Logger } from '../LoggerService.ts';
import path from 'path';
import type { DbConnection, IDbService } from '../../db/interface.ts';

const Log = Logger.child({ namespace: 'DbService/Wasm' });

// Define paths to the WASM and worker files using require.resolve for absolute paths
const getManualBundles = () => {
  const packagePath = import.meta.resolve('@duckdb/duckdb-wasm/package.json');
  const duckdbDir = path.dirname(packagePath);
  const distDir = path.join(duckdbDir, 'dist');

  return {
    eh: {
      mainModule: path.join(distDir, 'duckdb-eh.wasm'),
      mainWorker: path.join(distDir, 'duckdb-node-eh.worker.cjs'),
    },
  };
};

const MANUAL_BUNDLES = getManualBundles();

// Log bundle paths for debugging
Logger.child({ namespace: 'DbService/Wasm' }).debug('WASM bundles: %o', MANUAL_BUNDLES);

/**
 * Adapter class wrapping DuckDB's AsyncDuckDBConnection to match DbConnection interface.
 * DuckDB already provides the methods we need; we just type them correctly.
 */
class WasmConnection implements DbConnection {
  constructor(private connection: duckdb.AsyncDuckDBConnection) {}

  async prepare(query: string) {
    const preparedStatement = await this.connection.prepare(query);
    // Return the prepared statement directly - it has send() method
    return preparedStatement;
  }

  async query(query: string): Promise<unknown[] | null> {
    const arrowResult = await this.connection.query(query);
    if (!arrowResult) {
      return null;
    }
    // Convert Arrow result to JSON objects
    return arrowResult.toArray().map((row) => row.toJSON());
  }
}

export function createDbService(): IDbService {
  let db: duckdb.AsyncDuckDB | null = null;

  async function getDb(): Promise<DbConnection> {
    if (db !== null) {
      Log.debug('getDb: reuse');
      const connection = await db.connect();
      return new WasmConnection(connection);
    }

    Log.debug('getDb: initialize new instance');

    // Create worker with node worker bundle - convert path to file:// URL for Bun
    const workerPath = MANUAL_BUNDLES.eh.mainWorker;
    const workerUrl = workerPath.startsWith('/') ? `file://${workerPath}` : workerPath;
    Log.debug('Creating worker with URL: %s', workerUrl);

    // eslint-disable-next-line no-undef
    const worker = new Worker(workerUrl);
    const logger = new duckdb.ConsoleLogger();

    // Initialize AsyncDuckDB
    db = new duckdb.AsyncDuckDB(logger, worker);
    await db.instantiate(MANUAL_BUNDLES.eh.mainModule);

    const connection = await db.connect();

    try {
      // Install and load the markdown extension
      await connection.query('INSTALL markdown FROM community;');
      await connection.query('LOAD markdown;');
    } catch (error) {
      Log.error('Failed to initialize markdown extension: %s', error);
      throw new Error(
        `Failed to initialize markdown extension: ${error instanceof Error ? error.message : String(error)}`
      );
    }

    Log.debug('initialized');
    return new WasmConnection(connection);
  }

  Log.debug('initialized');
  return {
    getDb,
  };
}
