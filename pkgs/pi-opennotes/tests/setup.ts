/**
 * Test setup and utilities for pi-opennotes
 */

import { beforeAll, afterAll } from "bun:test";

// Re-export mock utilities
export * from "./fixtures/mocks";

// =============================================================================
// Global Test Setup
// =============================================================================

beforeAll(() => {
  // Set test environment
  process.env.NODE_ENV = "test";
});

afterAll(() => {
  // Cleanup
});

// =============================================================================
// Test Utilities
// =============================================================================

/**
 * Create a test timeout wrapper
 */
export function withTimeout<T>(promise: Promise<T>, ms: number = 5000): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`Test timed out after ${ms}ms`));
    }, ms);

    promise
      .then((result) => {
        clearTimeout(timer);
        resolve(result);
      })
      .catch((error) => {
        clearTimeout(timer);
        reject(error);
      });
  });
}

/**
 * Wait for a condition to be true
 */
export async function waitFor(
  condition: () => boolean | Promise<boolean>,
  options: { timeout?: number; interval?: number } = {}
): Promise<void> {
  const { timeout = 5000, interval = 100 } = options;
  const start = Date.now();

  while (Date.now() - start < timeout) {
    if (await condition()) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  throw new Error(`Condition not met within ${timeout}ms`);
}

/**
 * Extract mock call arguments
 */
export function getCallArgs(mockFn: any, callIndex: number = 0): any[] {
  return mockFn.mock?.calls?.[callIndex] ?? [];
}

/**
 * Get number of times a mock was called
 */
export function getCallCount(mockFn: any): number {
  return mockFn.mock?.calls?.length ?? 0;
}
