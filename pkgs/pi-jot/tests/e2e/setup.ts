/**
 * E2E Test Setup
 * 
 * Provides real CLI integration for end-to-end testing.
 * Tests will be skipped if Jot CLI is not installed.
 */

import { beforeAll, afterAll, describe, it, expect } from "bun:test";
import { existsSync, mkdirSync, rmSync, writeFileSync } from "fs";
import { join } from "path";
import { tmpdir } from "os";

// =============================================================================
// Test Environment Detection
// =============================================================================

export interface E2EEnvironment {
  cliAvailable: boolean;
  cliPath: string;
  cliVersion: string;
  testNotebookPath: string;
}

let e2eEnv: E2EEnvironment | null = null;

/**
 * Check if Jot CLI is available
 */
async function checkCli(): Promise<{ available: boolean; path: string; version: string }> {
  try {
    const proc = Bun.spawn(["jot", "version"], {
      stdout: "pipe",
      stderr: "pipe",
    });
    
    const stdout = await new Response(proc.stdout).text();
    const exitCode = await proc.exited;
    
    if (exitCode === 0) {
      // Parse version from output like "Jot 0.0.2"
      const versionMatch = stdout.match(/(\d+\.\d+\.\d+)/);
      return {
        available: true,
        path: "jot",
        version: versionMatch ? versionMatch[1] : "unknown",
      };
    }
    return { available: false, path: "", version: "" };
  } catch {
    return { available: false, path: "", version: "" };
  }
}

/**
 * Get or initialize E2E environment
 */
export async function getE2EEnvironment(): Promise<E2EEnvironment> {
  if (!e2eEnv) {
    const cliCheck = await checkCli();
    const testDir = join(tmpdir(), `pi-jot-e2e-${Date.now()}`);
    
    e2eEnv = {
      cliAvailable: cliCheck.available,
      cliPath: cliCheck.path,
      cliVersion: cliCheck.version,
      testNotebookPath: testDir,
    };
  }
  return e2eEnv;
}

// =============================================================================
// Test Notebook Management
// =============================================================================

/**
 * Create a test notebook with sample notes
 */
export async function createTestNotebook(path: string): Promise<void> {
  // Create directory structure
  mkdirSync(join(path, "projects"), { recursive: true });
  mkdirSync(join(path, "tasks"), { recursive: true });
  mkdirSync(join(path, "meetings"), { recursive: true });

  // Create notebook config
  const config = {
    name: "E2E Test Notebook",
    description: "Temporary notebook for E2E testing",
    templates: {},
    views: {
      "custom-view": {
        description: "Custom test view",
        sql: "SELECT * FROM notes WHERE data->>'status' = :status LIMIT :limit",
        parameters: {
          status: { type: "string", default: "active" },
          limit: { type: "number", default: 10 },
        },
      },
    },
  };
  
  writeFileSync(join(path, ".jot.json"), JSON.stringify(config, null, 2));

  // Create sample notes with various content types
  const notes = [
    {
      path: "projects/alpha.md",
      content: `---
title: Project Alpha
tags: [project, active]
status: active
priority: high
created: 2026-01-15T10:00:00Z
---

# Project Alpha

This is a sample project note for E2E testing.

## Goals

- Test search functionality
- Test metadata extraction
- Validate pagination

## Status

Currently in progress.
`,
    },
    {
      path: "projects/beta.md",
      content: `---
title: Project Beta
tags: [project, planning]
status: planning
priority: medium
created: 2026-01-20T14:00:00Z
---

# Project Beta

Another project note for testing multiple results.

## Overview

This project is in planning phase.
`,
    },
    {
      path: "tasks/task-001.md",
      content: `---
title: Implement Feature X
tags: [task, alpha, feature]
status: in-progress
project: alpha
assigned: developer
created: 2026-01-18T09:00:00Z
---

# Task: Implement Feature X

Task for Project Alpha.

## Requirements

1. Build the thing
2. Test the thing
3. Deploy the thing

## Progress

- [x] Design complete
- [ ] Implementation in progress
- [ ] Testing pending
`,
    },
    {
      path: "tasks/task-002.md",
      content: `---
title: Review Documentation
tags: [task, docs]
status: todo
priority: low
created: 2026-01-22T11:30:00Z
---

# Task: Review Documentation

Review and update project documentation.
`,
    },
    {
      path: "meetings/standup-2026-01-28.md",
      content: `---
title: Standup 2026-01-28
tags: [meeting, standup]
date: 2026-01-28
attendees: [alice, bob, charlie]
created: 2026-01-28T09:00:00Z
---

# Standup Meeting - 2026-01-28

## Attendees

- Alice
- Bob
- Charlie

## Updates

### Alice
- Completed feature X
- Starting on feature Y

### Bob
- Working on bug fixes

### Charlie
- Documentation updates

## Action Items

- [ ] Follow up on deployment
`,
    },
    {
      path: "README.md",
      content: `---
title: E2E Test Notebook
tags: [readme, index]
---

# E2E Test Notebook

This notebook contains sample notes for end-to-end testing.

## Contents

- [Projects](projects/) - Project documentation
- [Tasks](tasks/) - Task tracking
- [Meetings](meetings/) - Meeting notes

## Links

See [[projects/alpha]] for the main project.
`,
    },
  ];

  for (const note of notes) {
    writeFileSync(join(path, note.path), note.content);
  }
}

/**
 * Remove test notebook
 */
export function removeTestNotebook(path: string): void {
  if (existsSync(path)) {
    rmSync(path, { recursive: true, force: true });
  }
}

// =============================================================================
// CLI Execution Helpers
// =============================================================================

export interface CliExecResult {
  code: number;
  stdout: string;
  stderr: string;
}

/**
 * Execute Jot CLI command
 */
export async function execCli(
  args: string[],
  options?: { notebook?: string; timeout?: number }
): Promise<CliExecResult> {
  const fullArgs = [...args];
  if (options?.notebook) {
    fullArgs.push("--notebook", options.notebook);
  }

  const proc = Bun.spawn(["jot", ...fullArgs], {
    stdout: "pipe",
    stderr: "pipe",
  });

  let timeoutId: Timer | undefined;
  if (options?.timeout) {
    timeoutId = setTimeout(() => proc.kill(), options.timeout);
  }

  const [stdout, stderr, code] = await Promise.all([
    new Response(proc.stdout).text(),
    new Response(proc.stderr).text(),
    proc.exited,
  ]);

  if (timeoutId) clearTimeout(timeoutId);

  return { code, stdout, stderr };
}

/**
 * Parse JSON output from CLI
 */
export function parseCliJson<T>(stdout: string): T {
  if (!stdout.trim()) return [] as unknown as T;
  return JSON.parse(stdout);
}

// =============================================================================
// Test Skip Helpers
// =============================================================================

/**
 * Skip test if CLI is not available
 */
export function skipIfNoCli(testFn: () => void | Promise<void>): () => void | Promise<void> {
  return async () => {
    const env = await getE2EEnvironment();
    if (!env.cliAvailable) {
      console.log("⏭️  Skipping: Jot CLI not installed");
      return;
    }
    await testFn();
  };
}

/**
 * Create a describe block that skips if CLI is unavailable
 */
export function describeE2E(
  name: string,
  fn: () => void
): void {
  describe(`[E2E] ${name}`, fn);
}

// =============================================================================
// Performance Measurement
// =============================================================================

export interface PerformanceResult {
  durationMs: number;
  memoryUsedMB: number;
}

/**
 * Measure execution performance
 */
export async function measurePerformance<T>(
  fn: () => Promise<T>
): Promise<{ result: T; performance: PerformanceResult }> {
  const startMemory = process.memoryUsage().heapUsed;
  const startTime = performance.now();

  const result = await fn();

  const endTime = performance.now();
  const endMemory = process.memoryUsage().heapUsed;

  return {
    result,
    performance: {
      durationMs: endTime - startTime,
      memoryUsedMB: (endMemory - startMemory) / (1024 * 1024),
    },
  };
}

// =============================================================================
// Assertions
// =============================================================================

export function expectWithinMs(actual: number, maxMs: number, context?: string): void {
  const msg = context ? `${context}: ${actual}ms > ${maxMs}ms` : `${actual}ms exceeds ${maxMs}ms`;
  expect(actual).toBeLessThan(maxMs);
}

export function expectValidJson(str: string): void {
  expect(() => JSON.parse(str)).not.toThrow();
}
