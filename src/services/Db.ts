import { DuckDBConnection } from '@duckdb/node-api';
import { DuckDBInstance } from '@duckdb/node-api';
import { Logger } from './LoggerService';

const Log = Logger.child({ namespace: 'DbService' });

export type DbService = ReturnType<typeof createDbService>;
export function createDbService() {
  let db: DuckDBInstance | null = null;

  async function getDb(): Promise<DuckDBConnection> {
    if (db !== null) {
      return await db.connect();
    }

    db = await DuckDBInstance.create(':memory:');
    const connection = await db.connect();

    try {
      // Install and load the markdown extension
      connection.run('INSTALL markdown FROM community;');
      connection.run('LOAD markdown;');
    } catch (error) {
      Log.error('Failed to initialize markdown extension: %s', error);
      throw new Error(
        `Failed to initialize markdown extension: ${error instanceof Error ? error.message : String(error)}`
      );
    }

    return connection;
  }

  Log.debug('initialized');
  return {
    getDb,
  };
}
