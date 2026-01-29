/**
 * E2E Performance Tests
 * 
 * Tests response times, memory usage, and large datasets
 */

import { describe, it, expect, beforeAll, afterAll } from "bun:test";
import { writeFileSync } from "fs";
import { join } from "path";
import {
  getE2EEnvironment,
  createTestNotebook,
  removeTestNotebook,
  execCli,
  parseCliJson,
  skipIfNoCli,
  describeE2E,
  expectWithinMs,
  measurePerformance,
} from "./setup";

describeE2E("Performance", () => {
  let notebookPath: string;

  beforeAll(async () => {
    const env = await getE2EEnvironment();
    if (!env.cliAvailable) return;
    
    notebookPath = env.testNotebookPath;
    await createTestNotebook(notebookPath);
    
    // Create additional notes for performance testing
    for (let i = 1; i <= 50; i++) {
      const content = `---
title: Performance Test Note ${i}
tags: [test, performance, batch-${Math.floor(i / 10)}]
status: ${i % 2 === 0 ? 'active' : 'inactive'}
priority: ${i % 3 === 0 ? 'high' : i % 3 === 1 ? 'medium' : 'low'}
created: 2026-01-${String(i % 28 + 1).padStart(2, '0')}T10:00:00Z
---

# Performance Test Note ${i}

This is a performance test note with some content.

## Section 1

Lorem ipsum dolor sit amet, consectetur adipiscing elit.

## Section 2

Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

## Links

See [[projects/alpha]] for more information.

## Tags

#performance #test #batch-${Math.floor(i / 10)}
`;
      writeFileSync(join(notebookPath, `perf-${i}.md`), content);
    }
  });

  afterAll(() => {
    if (notebookPath) {
      removeTestNotebook(notebookPath);
    }
  });

  it("should list large dataset quickly", skipIfNoCli(async () => {
    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        ["notes", "list", "--output", "json"],
        { notebook: notebookPath, timeout: 10000 }
      );
    });

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(50);
    
    // Should complete within 5 seconds
    expectWithinMs(performance.durationMs, 5000, "List all notes");
    
    console.log(`  Listed ${notes.length} notes in ${performance.durationMs.toFixed(0)}ms`);
  }));

  it("should search large dataset efficiently", skipIfNoCli(async () => {
    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        ["notes", "search", "--query", "performance", "--output", "json"],
        { notebook: notebookPath, timeout: 10000 }
      );
    });

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    
    // Search should be fast
    expectWithinMs(performance.durationMs, 3000, "Search notes");
    
    console.log(`  Found ${notes.length} notes in ${performance.durationMs.toFixed(0)}ms`);
  }));

  it("should handle SQL queries efficiently", skipIfNoCli(async () => {
    const sql = "SELECT path, title, data->>'status' as status FROM notes WHERE data->>'status' = 'active' ORDER BY data->>'created' DESC LIMIT 20";
    
    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        ["notes", "sql", "--query", sql, "--output", "json"],
        { notebook: notebookPath, timeout: 10000 }
      );
    });

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    
    // Complex SQL should still be fast
    expectWithinMs(performance.durationMs, 5000, "SQL query");
    
    console.log(`  Executed SQL query in ${performance.durationMs.toFixed(0)}ms`);
  }));

  it("should paginate efficiently", skipIfNoCli(async () => {
    const pageSize = 10;
    let totalNotes = 0;
    let totalTime = 0;

    // Fetch first 3 pages
    for (let page = 0; page < 3; page++) {
      const { result, performance } = await measurePerformance(async () => {
        return await execCli(
          [
            "notes", "list",
            "--limit", String(pageSize),
            "--offset", String(page * pageSize),
            "--output", "json"
          ],
          { notebook: notebookPath, timeout: 5000 }
        );
      });

      expect(result.code).toBe(0);
      const notes = parseCliJson<any[]>(result.stdout);
      totalNotes += notes.length;
      totalTime += performance.durationMs;
      
      // Each page should be fast
      expectWithinMs(performance.durationMs, 3000, `Page ${page + 1}`);
    }

    console.log(`  Fetched ${totalNotes} notes across 3 pages in ${totalTime.toFixed(0)}ms`);
    console.log(`  Average per page: ${(totalTime / 3).toFixed(0)}ms`);
  }));

  it("should handle fuzzy search on large dataset", skipIfNoCli(async () => {
    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        ["notes", "search", "--query", "preformance", "--fuzzy", "--output", "json"],
        { notebook: notebookPath, timeout: 10000 }
      );
    });

    expect(result.code).toBe(0);
    
    // Fuzzy search is slower but should still complete in reasonable time
    expectWithinMs(performance.durationMs, 8000, "Fuzzy search");
    
    console.log(`  Fuzzy search completed in ${performance.durationMs.toFixed(0)}ms`);
  }));

  it("should sort efficiently", skipIfNoCli(async () => {
    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        [
          "notes", "list",
          "--sort", "modified",
          "--order", "desc",
          "--limit", "20",
          "--output", "json"
        ],
        { notebook: notebookPath, timeout: 5000 }
      );
    });

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    
    // Sorting should not significantly impact performance
    expectWithinMs(performance.durationMs, 4000, "Sorted list");
    
    console.log(`  Sorted ${notes.length} notes in ${performance.durationMs.toFixed(0)}ms`);
  }));

  it("should handle multiple rapid queries", skipIfNoCli(async () => {
    const queries = [
      ["notes", "list", "--limit", "10", "--output", "json"],
      ["notes", "search", "--query", "test", "--output", "json"],
      ["notes", "list", "--limit", "10", "--offset", "10", "--output", "json"],
    ];

    const startTime = performance.now();
    
    // Execute queries in parallel
    const results = await Promise.all(
      queries.map(args => 
        execCli(args, { notebook: notebookPath, timeout: 5000 })
      )
    );

    const totalTime = performance.now() - startTime;

    // All should succeed
    results.forEach((result, i) => {
      expect(result.code).toBe(0);
    });

    // Total time should be reasonable (parallel execution)
    expectWithinMs(totalTime, 10000, "Parallel queries");
    
    console.log(`  Executed ${queries.length} parallel queries in ${totalTime.toFixed(0)}ms`);
  }));

  it("should handle complex SQL with aggregations", skipIfNoCli(async () => {
    const sql = `
      SELECT 
        data->>'status' as status,
        COUNT(*) as count,
        AVG(LENGTH(content)) as avg_length
      FROM notes
      WHERE data->>'tags' LIKE '%performance%'
      GROUP BY data->>'status'
      ORDER BY count DESC
    `;

    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        ["notes", "sql", "--query", sql, "--output", "json"],
        { notebook: notebookPath, timeout: 10000 }
      );
    });

    expect(result.code).toBe(0);
    const results = parseCliJson<any[]>(result.stdout);
    expect(results.length).toBeGreaterThan(0);
    
    // Aggregations might be slower
    expectWithinMs(performance.durationMs, 8000, "SQL aggregation");
    
    console.log(`  SQL aggregation completed in ${performance.durationMs.toFixed(0)}ms`);
  }));

  it("should stay within memory limits", skipIfNoCli(async () => {
    const { result, performance } = await measurePerformance(async () => {
      return await execCli(
        ["notes", "list", "--output", "json"],
        { notebook: notebookPath, timeout: 10000 }
      );
    });

    expect(result.code).toBe(0);
    
    // Memory usage should be reasonable (< 50MB for this operation)
    expect(Math.abs(performance.memoryUsedMB)).toBeLessThan(50);
    
    console.log(`  Memory delta: ${performance.memoryUsedMB.toFixed(2)}MB`);
  }));
});

describeE2E("Budget Management", () => {
  it("should fit results within 75% of typical context budget", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "list", "--limit", "50", "--output", "json"],
      { timeout: 5000 }
    );

    expect(result.code).toBe(0);
    
    const outputSize = Buffer.byteLength(result.stdout, 'utf8');
    // Typical context window: 200k tokens ≈ 800KB
    // 75% of 800KB = 600KB
    const maxSize = 600 * 1024;
    
    expect(outputSize).toBeLessThan(maxSize);
    
    console.log(`  Output size: ${(outputSize / 1024).toFixed(2)}KB (max: ${(maxSize / 1024).toFixed(0)}KB)`);
  }));
});
