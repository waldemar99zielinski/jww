export interface RuntimeConfig {
  apiBaseUrl?: string
}

export function runtimeConfig(): RuntimeConfig {
  return window.__JWW_CONFIG__ ?? {}
}
