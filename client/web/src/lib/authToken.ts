/**
 * In-memory + localStorage-backed token store, kept separate from Redux so
 * api/client.ts (a plain module, not a React/Redux consumer) can read the
 * current token without importing the store and creating an import cycle
 * (client.ts -> store -> authSlice -> api/auth.ts -> client.ts).
 */
const STORAGE_KEY = 'auth_token'

let currentToken: string | null = localStorage.getItem(STORAGE_KEY)

export function getAuthToken(): string | null {
  return currentToken
}

export function setAuthToken(token: string | null): void {
  currentToken = token
  if (token) {
    localStorage.setItem(STORAGE_KEY, token)
  } else {
    localStorage.removeItem(STORAGE_KEY)
  }
}
