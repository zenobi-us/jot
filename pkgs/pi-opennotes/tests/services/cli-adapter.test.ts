/**
 * Unit tests for CliAdapter
 */

import { describe, it, expect, beforeEach, mock } from "bun:test";
import { CliAdapter } from "../../src/services/cli-adapter";
import { OpenNotesError, ErrorCodes } from "../../src/utils/errors";

describe("CliAdapter", () => {
  let adapter: CliAdapter;
  let mockPi: any;

  beforeEach(() => {
    mockPi = {
      exec: mock(async (cmd: string, args: string[], opts?: any) => ({
        code: 0,
        stdout: "[]",
        stderr: "",
      })),
    };

    adapter = new CliAdapter(mockPi, {
      cliPath: "opennotes",
      defaultTimeout: 30000,
    });

    // Clear cache between tests
    adapter.clearCache();
  });

  describe("exec", () => {
    it("executes command with args", async () => {
      await adapter.exec("opennotes", ["notes", "search"]);

      expect(mockPi.exec).toHaveBeenCalledWith(
        "opennotes",
        ["notes", "search"],
        expect.objectContaining({ timeout: 30000 })
      );
    });

    it("adds notebook flag when provided", async () => {
      await adapter.exec("opennotes", ["notes", "list"], {
        notebook: "/path/to/notebook",
      });

      expect(mockPi.exec).toHaveBeenCalledWith(
        "opennotes",
        ["notes", "list", "--notebook", "/path/to/notebook"],
        expect.any(Object)
      );
    });

    it("uses custom timeout when provided", async () => {
      await adapter.exec("opennotes", ["notes", "list"], {
        timeout: 5000,
      });

      expect(mockPi.exec).toHaveBeenCalledWith(
        "opennotes",
        expect.any(Array),
        expect.objectContaining({ timeout: 5000 })
      );
    });

    it("returns CLI result", async () => {
      mockPi.exec = mock(async () => ({
        code: 0,
        stdout: '{"test": true}',
        stderr: "",
      }));

      const result = await adapter.exec("opennotes", ["version"]);

      expect(result.code).toBe(0);
      expect(result.stdout).toBe('{"test": true}');
      expect(result.stderr).toBe("");
    });

    it("handles non-zero exit code", async () => {
      mockPi.exec = mock(async () => ({
        code: 1,
        stdout: "",
        stderr: "Command failed",
      }));

      const result = await adapter.exec("opennotes", ["invalid"]);

      expect(result.code).toBe(1);
      expect(result.stderr).toBe("Command failed");
    });
  });

  describe("checkInstallation", () => {
    it("returns installed=true when CLI responds", async () => {
      mockPi.exec = mock(async () => ({
        code: 0,
        stdout: "opennotes version 0.10.5",
        stderr: "",
      }));

      const result = await adapter.checkInstallation();

      expect(result.installed).toBe(true);
      expect(result.version).toBe("0.10.5");
      expect(result.path).toBe("opennotes");
    });

    it("returns installed=false when CLI fails", async () => {
      mockPi.exec = mock(async () => ({
        code: 1,
        stdout: "",
        stderr: "command not found",
      }));

      const result = await adapter.checkInstallation();

      expect(result.installed).toBe(false);
      expect(result.version).toBeUndefined();
    });

    it("caches installation check result", async () => {
      mockPi.exec = mock(async () => ({
        code: 0,
        stdout: "opennotes version 1.0.0",
        stderr: "",
      }));

      await adapter.checkInstallation();
      await adapter.checkInstallation();

      // Should only call exec once due to caching
      expect(mockPi.exec.mock.calls.length).toBe(1);
    });

    it("handles version format variations", async () => {
      mockPi.exec = mock(async () => ({
        code: 0,
        stdout: "version 2.0.0-beta",
        stderr: "",
      }));

      const result = await adapter.checkInstallation();

      expect(result.installed).toBe(true);
      expect(result.version).toBe("2.0.0-beta");
    });
  });

  describe("parseJsonOutput", () => {
    it("parses valid JSON", () => {
      const result = adapter.parseJsonOutput<{ name: string }>(
        '{"name": "test"}'
      );

      expect(result).toEqual({ name: "test" });
    });

    it("parses JSON arrays", () => {
      const result = adapter.parseJsonOutput<string[]>('["a", "b", "c"]');

      expect(result).toEqual(["a", "b", "c"]);
    });

    it("returns empty array for empty input", () => {
      const result = adapter.parseJsonOutput<unknown[]>("");

      expect(result).toEqual([]);
    });

    it("returns empty array for whitespace input", () => {
      const result = adapter.parseJsonOutput<unknown[]>("   \n  ");

      expect(result).toEqual([]);
    });

    it("throws OpenNotesError for invalid JSON", () => {
      expect(() => adapter.parseJsonOutput("not json")).toThrow(OpenNotesError);
    });

    it("includes stdout preview in error details", () => {
      try {
        adapter.parseJsonOutput("invalid json content here");
        expect.unreachable("Should have thrown");
      } catch (error) {
        expect(error).toBeInstanceOf(OpenNotesError);
        expect((error as OpenNotesError).code).toBe(ErrorCodes.PARSE_ERROR);
        expect((error as OpenNotesError).details?.stdout).toBeDefined();
      }
    });
  });

  describe("buildNotebookArgs", () => {
    it("returns empty array when no notebook", () => {
      expect(adapter.buildNotebookArgs()).toEqual([]);
    });

    it("returns empty array when notebook is undefined", () => {
      expect(adapter.buildNotebookArgs(undefined)).toEqual([]);
    });

    it("returns notebook args when provided", () => {
      expect(adapter.buildNotebookArgs("/path/to/notebook")).toEqual([
        "--notebook",
        "/path/to/notebook",
      ]);
    });
  });
});
