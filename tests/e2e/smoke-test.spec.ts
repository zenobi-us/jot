import { describe, it, beforeEach, afterEach, expect } from 'vitest';
import { existsSync, mkdirSync, rmSync, writeFileSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';

const OPENNOTES_BIN = join(process.cwd(), 'dist', 'opennotes');

const runCommand = () => {
  const abortController = new globalThis.AbortController();

  const abort = () => abortController.abort();

  const run = async (
    args: string[],
    options?: {
      cwd?: string;
      timeout?: number;
    }
  ) => {
    const proc = Bun.spawn([OPENNOTES_BIN, ...args], {
      cwd: options?.cwd || process.cwd(),
      stdout: 'pipe',
      stderr: 'pipe',
      signal: abortController.signal,
      timeout: options?.timeout || 30000,
    });

    await proc.exited;

    return proc;
  };

  return {
    run,
    abort,
  };
};

describe('opennotes smoke tests', () => {
  let tmpDir: string;
  let notebookDir: string;

  beforeEach(() => {
    // Create temporary directory
    tmpDir = join(tmpdir(), `opennotes-smoke-${Date.now()}-${Math.random()}`);
    notebookDir = join(tmpDir, 'test-notebook');
    mkdirSync(notebookDir, { recursive: true });

    // Create .opennotes.json
    writeFileSync(
      join(notebookDir, '.opennotes.json'),
      JSON.stringify(
        {
          name: 'Test Notebook',
          description: 'Smoke test notebook',
          root: '.',
          contexts: [notebookDir],
        },
        null,
        2
      )
    );

    // Create markdown files
    writeFileSync(join(notebookDir, 'note1.md'), '# Note 1\n\nFirst note');
    writeFileSync(join(notebookDir, 'note2.md'), '# Note 2\n\nSecond note');
    writeFileSync(join(notebookDir, 'note3.md'), '# Note 3\n\nThird note');
  });

  afterEach(() => {
    // Clean up temporary directory
    rmSync(tmpDir, { recursive: true, force: true });
  });

  it('binary should exist', () => {
    expect(existsSync(OPENNOTES_BIN)).toBe(true);
  });

  it('should create notebook with markdown files', () => {
    expect(existsSync(join(notebookDir, '.opennotes.json'))).toBe(true);
    expect(existsSync(join(notebookDir, 'note1.md'))).toBe(true);
    expect(existsSync(join(notebookDir, 'note2.md'))).toBe(true);
    expect(existsSync(join(notebookDir, 'note3.md'))).toBe(true);
  });

  it('should list notes from notebook directory', async () => {
    const proc = runCommand();

    const results = await proc.run([`notes`, `list`], { cwd: notebookDir });

    expect(results.exitCode).toBe(0);
    const noteCount = ((await results.stdout.text()).trim().match(/note\d\.md/g) || []).length;
    expect(noteCount).toBe(3);
  });

  it('should display notebook info', async () => {
    const result = await runCommand().run(['notebook'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    expect(output).toContain('Test Notebook');
  });

  it('should list notebooks', async () => {
    const result = await runCommand().run(['notebook', 'list']);

    expect(result.exitCode).toBe(0);
  });

  it('should search notes by content', async () => {
    const result = await runCommand().run(['notes', 'search', 'First'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    expect(output).toContain('note1.md');
  });

  it('should handle nested markdown files', async () => {
    const nestedDir = join(notebookDir, 'nested', 'folder');
    mkdirSync(nestedDir, { recursive: true });
    writeFileSync(join(nestedDir, 'nested-note.md'), '# Nested Note\n\nThis is nested');

    const result = await runCommand().run(['notes', 'list'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    expect(output).toContain('nested-note.md');
  });

  it('should handle notebooks with special characters in filenames', async () => {
    writeFileSync(join(notebookDir, 'note-with-dashes.md'), '# Note with dashes\n\nContent');
    writeFileSync(
      join(notebookDir, 'note_with_underscores.md'),
      '# Note with underscores\n\nContent'
    );

    const result = await runCommand().run(['notes', 'list'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    expect(output).toContain('note-with-dashes.md');
    expect(output).toContain('note_with_underscores.md');
  });

  it('should handle empty notebooks', async () => {
    const emptyNotebookDir = join(tmpDir, 'empty-notebook');
    mkdirSync(emptyNotebookDir, { recursive: true });

    writeFileSync(
      join(emptyNotebookDir, '.opennotes.json'),
      JSON.stringify(
        {
          name: 'Empty Notebook',
          description: 'Empty notebook for testing',
          root: '.',
          contexts: [emptyNotebookDir],
        },
        null,
        2
      )
    );

    const result = await runCommand().run(['notes', 'list'], {
      cwd: emptyNotebookDir,
    });

    expect(result.exitCode).toBe(0);
  });

  it('should handle notes with frontmatter', async () => {
    const noteWithFrontmatter = `---
title: "Test Note"
tags: ["test", "smoke"]
date: 2024-01-08
---

# Test Note with Frontmatter

This note has frontmatter metadata.`;

    writeFileSync(join(notebookDir, 'frontmatter.md'), noteWithFrontmatter);

    const result = await runCommand().run(['notes', 'list'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    expect(output).toContain('frontmatter.md');
  });

  it('should register notebook in config', async () => {
    const configDir = join(tmpDir, '.config', 'opennotes');
    mkdirSync(configDir, { recursive: true });

    const globalConfig = {
      notebooks: [
        {
          name: 'test-nb',
          path: notebookDir,
        },
      ],
    };

    writeFileSync(join(configDir, 'config.json'), JSON.stringify(globalConfig, null, 2));

    // This should succeed because we're running from the notebook directory
    const result = await runCommand().run(['notebook'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
  });

  it('should handle large notebooks with many files', async () => {
    // Create 10 additional files
    for (let i = 4; i <= 13; i++) {
      writeFileSync(join(notebookDir, `note${i}.md`), `# Note ${i}\n\nContent for note ${i}`);
    }

    const result = await runCommand().run(['notes', 'list'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    const noteCount = (output.match(/note\d+\.md/g) || []).length;
    expect(noteCount).toBe(13); // 3 original + 10 new
  });

  it('should show notebook metadata', async () => {
    const result = await runCommand().run(['notebook'], {
      cwd: notebookDir,
    });

    expect(result.exitCode).toBe(0);
    const output = (await result.stdout.text()).trim();
    expect(output).toContain('Test Notebook');
    expect(output).toContain('Smoke test notebook');
  });
});
