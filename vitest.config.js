import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    globals: true,
    include: ['pkg/workflow/js/**/*.test.{js,cjs,ts}'],
    testTimeout: 10000,
    hookTimeout: 10000,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['pkg/workflow/js/**/*.{cjs,ts}'],
      exclude: ['pkg/workflow/js/**/*.test.{js,cjs,ts}']
    }
  }
});