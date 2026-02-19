/**
 * E2E Multi-Notebook Tests
 * 
 * Tests handling multiple notebooks and notebook switching
 */

import { describe, it, expect, beforeAll, afterAll } from "bun:test";
import { mkdirSync, writeFileSync } from "fs";
import { join } from "path";
import { tmpdir } from "os";
import {
  getE2EEnvironment,
  createTestNotebook,
  removeTestNotebook,
  execCli,
  parseCliJson,
  skipIfNoCli,
  describeE2E,
} from "./setup";

describeE2E("Multi-Notebook", () => {
  let notebook1Path: string;
  let notebook2Path: string;

  beforeAll(async () => {
    const env = await getE2EEnvironment();
    if (!env.cliAvailable) return;
    
    const tmpBase = tmpdir();
    notebook1Path = join(tmpBase, `pi-jot-nb1-${Date.now()}`);
    notebook2Path = join(tmpBase, `pi-jot-nb2-${Date.now()}`);
    
    // Create two different notebooks
    await createTestNotebook(notebook1Path);
    
    // Create second notebook with different content
    mkdirSync(join(notebook2Path, "docs"), { recursive: true });
    
    const nb2Config = {
      name: "Notebook 2",
      description: "Second test notebook",
      views: {
        "docs-only": {
          description: "Only documentation",
          sql: "SELECT * FROM notes WHERE path LIKE 'docs/%'",
        },
      },
    };
    
    writeFileSync(
      join(notebook2Path, ".jot.json"),
      JSON.stringify(nb2Config, null, 2)
    );

    writeFileSync(
      join(notebook2Path, "docs/guide.md"),
      `---
title: User Guide
tags: [docs, guide]
---

# User Guide

This is from notebook 2.
`
    );

    writeFileSync(
      join(notebook2Path, "docs/api.md"),
      `---
title: API Reference
tags: [docs, api]
---

# API Reference

API documentation from notebook 2.
`
    );
  });

  afterAll(() => {
    if (notebook1Path) removeTestNotebook(notebook1Path);
    if (notebook2Path) removeTestNotebook(notebook2Path);
  });

  it("should list notes from notebook 1", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "list", "--output", "json"],
      { notebook: notebook1Path, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    
    // Should have Project Alpha from notebook 1
    const hasAlpha = notes.some((n: any) => n.path?.includes("alpha"));
    expect(hasAlpha).toBe(true);
  }));

  it("should list notes from notebook 2", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "list", "--output", "json"],
      { notebook: notebook2Path, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    
    // Should have docs from notebook 2
    const hasDocs = notes.some((n: any) => n.path?.includes("docs/"));
    expect(hasDocs).toBe(true);
    
    // Should NOT have Project Alpha
    const hasAlpha = notes.some((n: any) => n.path?.includes("alpha"));
    expect(hasAlpha).toBe(false);
  }));

  it("should execute view from notebook 1", skipIfNoCli(async () => {
    const result = await execCli(
      [
        "notes", "view", "custom-view",
        "--param", "status=active",
        "--output", "json"
      ],
      { notebook: notebook1Path, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
  }));

  it("should execute view from notebook 2", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "view", "docs-only", "--output", "json"],
      { notebook: notebook2Path, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    
    // All results should be from docs/
    notes.forEach((note: any) => {
      expect(note.path).toMatch(/docs\//);
    });
  }));

  it("should not find notebook 1 views in notebook 2", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "view", "custom-view", "--output", "json"],
      { notebook: notebook2Path, timeout: 5000 }
    );

    // Should fail - custom-view is only in notebook 1
    expect(result.code).not.toBe(0);
  }));

  it("should search only within specified notebook", skipIfNoCli(async () => {
    // Search in notebook 1
    const result1 = await execCli(
      ["notes", "search", "--query", "Project", "--output", "json"],
      { notebook: notebook1Path, timeout: 5000 }
    );

    expect(result1.code).toBe(0);
    const notes1 = parseCliJson<any[]>(result1.stdout);
    const hasProjects = notes1.some((n: any) => n.title?.includes("Project"));
    expect(hasProjects).toBe(true);

    // Search in notebook 2
    const result2 = await execCli(
      ["notes", "search", "--query", "Project", "--output", "json"],
      { notebook: notebook2Path, timeout: 5000 }
    );

    // Might return 0 results or empty array
    if (result2.code === 0) {
      const notes2 = parseCliJson<any[]>(result2.stdout);
      const hasProjects2 = notes2.some((n: any) => n.title?.includes("Project"));
      expect(hasProjects2).toBe(false);
    }
  }));

  it("should get note from correct notebook", skipIfNoCli(async () => {
    // Get from notebook 1
    const result1 = await execCli(
      ["notes", "get", "projects/alpha.md", "--output", "json"],
      { notebook: notebook1Path, timeout: 5000 }
    );

    expect(result1.code).toBe(0);
    const note1 = parseCliJson<any>(result1.stdout);
    expect(note1.title).toBe("Project Alpha");

    // Try same path in notebook 2 (should fail)
    const result2 = await execCli(
      ["notes", "get", "projects/alpha.md", "--output", "json"],
      { notebook: notebook2Path, timeout: 5000 }
    );

    expect(result2.code).not.toBe(0);
  }));

  it("should handle notebook without views", skipIfNoCli(async () => {
    // Create minimal notebook
    const minimalPath = join(tmpdir(), `pi-jot-minimal-${Date.now()}`);
    mkdirSync(minimalPath, { recursive: true });
    
    writeFileSync(
      join(minimalPath, ".jot.json"),
      JSON.stringify({ name: "Minimal" }, null, 2)
    );

    writeFileSync(
      join(minimalPath, "note.md"),
      "# Simple Note\n\nContent here."
    );

    try {
      // List notes should work
      const listResult = await execCli(
        ["notes", "list", "--output", "json"],
        { notebook: minimalPath, timeout: 5000 }
      );

      expect(listResult.code).toBe(0);

      // View list should return empty or fail gracefully
      const viewResult = await execCli(
        ["notes", "views", "--output", "json"],
        { notebook: minimalPath, timeout: 5000 }
      );

      if (viewResult.code === 0) {
        const views = parseCliJson<any[]>(viewResult.stdout);
        expect(views.length).toBe(0);
      }
    } finally {
      removeTestNotebook(minimalPath);
    }
  }));

  it("should list all configured notebooks", skipIfNoCli(async () => {
    const result = await execCli(
      ["notebooks", "list", "--output", "json"],
      { timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notebooks = parseCliJson<any[]>(result.stdout);
    
    // Should have at least some notebooks
    expect(notebooks.length).toBeGreaterThan(0);
    
    // Each notebook should have required fields
    notebooks.forEach((nb: any) => {
      expect(nb).toHaveProperty("name");
      expect(nb).toHaveProperty("path");
    });
  }));

  it("should isolate SQL queries to notebook", skipIfNoCli(async () => {
    // Query notebook 1
    const result1 = await execCli(
      [
        "notes", "sql",
        "--query", "SELECT COUNT(*) as count FROM notes",
        "--output", "json"
      ],
      { notebook: notebook1Path, timeout: 5000 }
    );

    expect(result1.code).toBe(0);
    const count1 = parseCliJson<any[]>(result1.stdout);

    // Query notebook 2
    const result2 = await execCli(
      [
        "notes", "sql",
        "--query", "SELECT COUNT(*) as count FROM notes",
        "--output", "json"
      ],
      { notebook: notebook2Path, timeout: 5000 }
    );

    expect(result2.code).toBe(0);
    const count2 = parseCliJson<any[]>(result2.stdout);

    // Counts should be different
    expect(count1[0].count).not.toBe(count2[0].count);
  }));
});
