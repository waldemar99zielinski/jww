/**
 * The real Go backend's base URL (not Vite's `VITE_API_BASE_URL`, since Playwright's Node
 * process doesn't load the app's `.env`) — defaults to the same default as `api/client.ts`.
 * Override with `PLAYWRIGHT_API_BASE_URL` if your backend runs elsewhere (see tests/e2e/README.md).
 */
export const API_BASE_URL = process.env.PLAYWRIGHT_API_BASE_URL ?? 'http://localhost:8000'

export interface TestUser {
  username: string
  password: string
}

/**
 * There is no working self-registration endpoint to call from a spec: `POST /api/v1/users` is
 * documented in the root README but isn't actually wired to any server route (`SignupWorkflow`
 * is built in `cmd/server/deps.go` and never mounted — a real bug, left alone for this round, see
 * tests/e2e/README.md). So specs log in as one pre-existing account instead of registering a fresh one
 * per run — provision it once (see tests/e2e/README.md) and point these env vars at it.
 */
export function getTestUser(): TestUser {
  const username = process.env.PLAYWRIGHT_TEST_USERNAME
  const password = process.env.PLAYWRIGHT_TEST_PASSWORD
  if (!username || !password) {
    throw new Error(
      'Set PLAYWRIGHT_TEST_USERNAME and PLAYWRIGHT_TEST_PASSWORD to a pre-existing account before ' +
        'running the e2e suite — see tests/e2e/README.md ("Test user").',
    )
  }
  return { username, password }
}

/** A short, collision-free label for data this spec creates (category/currency names, …) —
 *  keeps specs parallel-safe against the shared DB without needing a fresh user per spec. */
export function uniqueLabel(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1e6)}`
}
