import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from '@/test/server'

// RTL's own auto-cleanup only registers against a *global* `afterEach` — this project runs
// Vitest with `globals: false`, so it never fires on its own. Without this, DOM from one test
// leaks into the next within the same file (e.g. two renders both matching the same role/text).
afterEach(() => cleanup())

beforeAll(() => server.listen({ onUnhandledFrame: 'error' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
