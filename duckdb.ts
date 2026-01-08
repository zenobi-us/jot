#!/usr/bin/env bun

import * as duckdb from '@duckdb/duckdb-wasm';
import path from 'path';
import { Worker } from 'worker_threads';

// 1. Define paths to the local WASM and worker files
const DUCKDB_DIST = path.resolve('node_modules/@duckdb/duckdb-wasm/dist');

const MANUAL_BUNDLES = {
  mvp: {
    mainModule: path.join(DUCKDB_DIST, 'duckdb-mvp.wasm'),
    mainWorker: path.join(DUCKDB_DIST, 'duckdb-node-mvp.worker.cjs'),
  },
  eh: {
    mainModule: path.join(DUCKDB_DIST, 'duckdb-eh.wasm'),
    mainWorker: path.join(DUCKDB_DIST, 'duckdb-node-eh.worker.cjs'),
  },
};

// 2. Instantiate DuckDB
const logger = new duckdb.ConsoleLogger();
const worker = new Worker(MANUAL_BUNDLES.mvp.mainWorker);
const db = new duckdb.AsyncDuckDB(logger, worker);

await db.instantiate(MANUAL_BUNDLES.mvp.mainModule);

// 3. Connect and Query
const conn = await db.connect();
const result = await conn.query('SELECT 42 AS answer');
console.log(result.toArray());

// 4. Cleanup
await conn.close();
await db.terminate();
await worker.terminate();
