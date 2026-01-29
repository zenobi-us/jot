/**
 * E2E Error Scenario Tests
 * 
 * Tests error handling with CLI not installed, missing notebooks, etc.
 */

import { describe, it, expect, beforeAll, afterAll } from "bun:test";
import {
  getE2EEnvironment,
  createTestNotebook,
  removeTestNotebook,
  execCli,
  skipIfNoCli,
  describeE2E,
} from "./setup";

describeE2E("Error Scenarios", () => {
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

  it("should handle missing notebook gracefully", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "list", "--output", "json"],
      { notebook: "/nonexistent/path", timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
    expect(result.stderr.toLowerCase()).toContain("notebook");
  }));

  it("should reject invalid SQL queries", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "sql", "--query", "DROP TABLE notes", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should handle malformed SQL", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "sql", "--query", "SELECT * FROM WHERE", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should handle missing note file", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "get", "nonexistent.md", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should handle invalid view name", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "view", "nonexistent-view", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should validate view parameters", skipIfNoCli(async () => {
    const result = await execCli(
      [
        "notes", "view", "custom-view",
        "--param", "limit=not-a-number",
        "--output", "json"
      ],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should handle SQL injection attempts", skipIfNoCli(async () => {
    const maliciousSql = "SELECT * FROM notes; DROP TABLE notes; --";
    const result = await execCli(
      ["notes", "sql", "--query", maliciousSql, "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    // Should either reject or safely handle
    if (result.code === 0) {
      // If it executes, verify it didn't do damage
      const verifyResult = await execCli(
        ["notes", "list", "--output", "json"],
        { notebook: notebookPath, timeout: 5000 }
      );
      expect(verifyResult.code).toBe(0);
    } else {
      // Rejected - good
      expect(result.stderr).toBeTruthy();
    }
  }));

  it("should handle path traversal attempts", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "get", "../../etc/passwd", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should handle empty search query", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "search", "--query", "", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    // Either reject or return all notes
    if (result.code === 0) {
      // Should return results
      expect(result.stdout).toBeTruthy();
    } else {
      expect(result.stderr).toBeTruthy();
    }
  }));

  it("should handle missing required flags", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "sql"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr).toBeTruthy();
  }));

  it("should timeout on very long queries", skipIfNoCli(async () => {
    // This is a meta-test - verify our timeout mechanism works
    const start = performance.now();
    
    try {
      await execCli(
        ["notes", "sql", "--query", "SELECT * FROM notes"],
        { notebook: notebookPath, timeout: 100 } // Very short timeout
      );
    } catch (error) {
      // Timeout expected
    }

    const duration = performance.now() - start;
    
    // Should have timed out quickly
    expect(duration).toBeLessThan(500);
  }, { timeout: 1000 });

  it("should provide helpful error messages", skipIfNoCli(async () => {
    const result = await execCli(
      ["notes", "get", "missing-note.md", "--output", "json"],
      { notebook: notebookPath, timeout: 5000 }
    );

    expect(result.code).not.toBe(0);
    expect(result.stderr.length).toBeGreaterThan(0);
    
    // Error should mention the file
    expect(result.stderr.toLowerCase()).toMatch(/not found|missing|does not exist/);
  }));
});

describeE2E("CLI Not Installed", () => {
  it("should detect when CLI is not available", async () => {
    const env = await getE2EEnvironment();
    
    if (!env.cliAvailable) {
      expect(env.cliPath).toBe("");
      expect(env.cliVersion).toBe("");
      console.log("✓ Correctly detected CLI not installed");
    } else {
      console.log(`✓ CLI detected: version ${env.cliVersion}`);
    }
  });
});
