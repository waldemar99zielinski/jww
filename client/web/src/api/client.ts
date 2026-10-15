import { getAuthToken } from '@/lib/authToken'

// In dev the client calls the Vite dev server on the same origin, and `server.proxy` forwards
// `/api` to the backend (see vite.config.ts). That keeps a single tunnel (e.g. ngrok) sufficient
// and takes CORS off the dev path. Production builds call the real API origin directly.
const API_BASE_URL = import.meta.env.DEV ? '' : (import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8001')
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

  const response = await fetch(`${API_BASE_URL}${API_PREFIX}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
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
