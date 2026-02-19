/**
 * E2E CLI Integration Tests
 * 
 * Tests real Jot CLI integration with actual notebooks
 */

import { describe, it, expect, beforeAll, afterAll } from "bun:test";
import {
  getE2EEnvironment,
  createTestNotebook,
  removeTestNotebook,
  execCli,
  parseCliJson,
  skipIfNoCli,
  describeE2E,
  expectWithinMs,
} from "./setup";

describeE2E("CLI Integration", () => {
  let notebookPath: string;

  beforeAll(async () => {
    const env = await getE2EEnvironment();
    if (!env.cliAvailable) return;
    
    notebookPath = env.testNotebookPath;
    await createTestNotebook(notebookPath);
  });

  afterAll(() => {
    if (notebookPath) {
      removeTestNotebook(notebookPath);
    }
  });

  it("should list notes from notebook", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "list", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    
    // Verify note structure
    const firstNote = notes[0];
    expect(firstNote).toHaveProperty("path");
    expect(firstNote).toHaveProperty("title");
  }));

  it("should search notes by text", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "search", "--query", "Project Alpha", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    
    // Verify it found the right note
    const alphaNote = notes.find((n: any) => n.path?.includes("alpha"));
    expect(alphaNote).toBeDefined();
  }));

  it("should search notes with SQL", skipIfNoCli(async () => {
    const sql = "SELECT path, title FROM notes WHERE data->>'status' = 'active' LIMIT 10";
    const result = await execCli(
      ["notes", "sql", "--query", sql, "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    
    // Verify status filter worked
    const activeNote = notes.find((n: any) => n.path?.includes("alpha"));
    expect(activeNote).toBeDefined();
  }));

  it("should get individual note content", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "get", "projects/alpha.md", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const note = parseCliJson<any>(result.stdout);
    expect(note.title).toBe("Project Alpha");
    expect(note.content).toContain("Goals");
  }));

  it("should execute custom view with parameters", skipIfNoCli(async () => {
    const result = await execCli(
      [
        "notes", "view", "custom-view",
        "--param", "status=active",
        "--param", "limit=5",
        "--output", "json"
      ],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    expect(notes.length).toBeGreaterThan(0);
    expect(notes.length).toBeLessThanOrEqual(5);
  }));

  it("should list notebooks", skipIfNoCli(async () => {
    const result = await execCli(
      ["notebooks", "list", "--output", "json"],
      { timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notebooks = parseCliJson<any[]>(result.stdout);
    expect(notebooks.length).toBeGreaterThan(0);
  }));

  it("should respond within acceptable time", skipIfNoCli(async () => {
    const start = performance.now();
    
    const result = await execCli(
      ["notes", "list", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    const duration = performance.now() - start;
    
    expect(result.code).toBe(0);
    expectWithinMs(duration, 5000, "CLI list command");
  }));

  it("should handle limit and offset", skipIfNoCli(async () => {
    // Get first page
    const page1 = await execCli(
      ["notes", "list", "--limit", "2", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(page1.code).toBe(0);
    const notes1 = parseCliJson<any[]>(page1.stdout);
    expect(notes1.length).toBeLessThanOrEqual(2);

    // Get second page
    const page2 = await execCli(
      ["notes", "list", "--limit", "2", "--offset", "2", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(page2.code).toBe(0);
    const notes2 = parseCliJson<any[]>(page2.stdout);
    
    // Pages should have different notes
    if (notes1.length > 0 && notes2.length > 0) {
      expect(notes1[0].path).not.toBe(notes2[0].path);
    }
  }));

  it("should search with fuzzy matching", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "search", "--query", "Projet", "--fuzzy", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).toBe(0);
    const notes = parseCliJson<any[]>(result.stdout);
    
    // Fuzzy search should find "Project" notes despite typo
    const projectNotes = notes.filter((n: any) => 
      n.title?.toLowerCase().includes("project")
    );
    expect(projectNotes.length).toBeGreaterThan(0);
  }));
});
