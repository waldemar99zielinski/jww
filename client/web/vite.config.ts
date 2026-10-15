import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'
import { configDefaults } from 'vitest/config'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      // Set VITE_DEV_ALLOWED_HOST in .env (e.g. to a tunnel hostname) to reach
      // the dev server from another device without editing this file.
      allowedHosts: env.VITE_DEV_ALLOWED_HOST ? [env.VITE_DEV_ALLOWED_HOST] : undefined,
    },
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      globals: false,
      // `tests/e2e/` is Playwright's own test runner (`playwright test`), not Vitest — Vitest's
      // default include glob would otherwise also pick up `tests/e2e/*.spec.ts` and crash trying
      // to run Playwright's `test.describe()` outside Playwright's runner.
      exclude: [...configDefaults.exclude, 'tests/e2e/**'],
      coverage: {
        provider: 'v8',
        reporter: ['text', 'html'],
      },
    },
  }
})
