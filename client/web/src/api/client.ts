import { getAuthToken } from '@/lib/authToken'
import { runtimeConfig } from '@/lib/runtimeConfig'

// In dev the client calls the Vite dev server on the same origin, and `server.proxy` forwards
// `/api` to the backend (see vite.config.ts). That keeps a single tunnel (e.g. ngrok) sufficient
// and takes CORS off the dev path. Production builds call the real API origin directly.
const BUILD_TIME_API_BASE_URL = import.meta.env.DEV ? '' : (import.meta.env.VITE_API_BASE_URL ?? '')
const API_PREFIX = '/api/v1'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getAuthToken()

  // A FormData body needs the browser's own Content-Type (which includes the multipart
  // boundary), so only default to JSON when the caller isn't sending a form.
  const isFormData = init?.body instanceof FormData

  const response = await fetch(apiUrl(path), {
    ...init,
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  })

  if (!response.ok) {
    throw new ApiError(response.status, await readErrorMessage(response))
  }

  if (response.status === 204) {
    return undefined as T
  }

  const body = (await response.json()) as { data: T }
  return body.data
}

/**
 * Full URL for a path under the API prefix. Use this for raw browser-driven requests that can't
 * go through `apiFetch` — e.g. an `<img src>` (which can't carry the auth header, so it must use a
 * grant URL) or a `target="_blank"` preview link.
 */
export function apiUrl(path: string): string {
  // An explicitly empty runtime value means same-origin. Nullish coalescing is
  // intentional: the container emits "" by default and must override Vite's
  // standalone-build fallback without rebuilding the image.
  const apiBaseUrl = runtimeConfig().apiBaseUrl ?? BUILD_TIME_API_BASE_URL
  return `${apiBaseUrl.replace(/\/$/, '')}${API_PREFIX}${path}`
}

/**
 * Downloads raw bytes. Attachments (and only attachments) are the documented exception to the JSON
 * `{data}` envelope, so this mirrors `apiFetch`'s auth header and error handling but returns the
 * body as a `Blob` instead of unwrapping JSON.
 */
export async function apiDownload(path: string): Promise<Blob> {
  const token = getAuthToken()
  const response = await fetch(apiUrl(path), {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!response.ok) {
    throw new ApiError(response.status, await readErrorMessage(response))
  }
  return response.blob()
}

async function readErrorMessage(response: Response): Promise<string> {
  const text = await response.text()
  if (!text) return response.statusText
  try {
    const body = JSON.parse(text) as { error?: { message?: string } }
    return body.error?.message ?? text
  } catch {
    return text
  }
}

export function toQueryString(params: object): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params as Record<string, string | number | boolean | undefined>)) {
    if (value !== undefined) search.set(key, String(value))
  }
  const query = search.toString()
  return query ? `?${query}` : ''
}
