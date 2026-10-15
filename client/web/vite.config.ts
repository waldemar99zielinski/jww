import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'
import { configDefaults } from 'vitest/config'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  // Proxy the API through the dev/preview server so a single tunnel (e.g. ngrok) exposes both the
  // app and the API, and so requests are same-origin in dev (no CORS). Falls back to
  // VITE_API_BASE_URL so an existing .env keeps working without a second variable.
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || env.VITE_API_BASE_URL || 'http://localhost:8001'
  const apiProxy = {
    '/api': { target: apiProxyTarget, changeOrigin: true },
  }

  // Hostnames the dev/preview server may answer to. A leading dot matches the domain and all its
  // subdomains; localhost and IP addresses are always allowed regardless. This is what lets an
  // ngrok URL reach the server.
  const allowedHosts = [
    '.ngrok-free.app',
    '.ngrok-free.dev',
    '.ngrok.app',
    '.ngrok.dev',
    '.ngrok.io',
    ...(env.VITE_DEV_ALLOWED_HOST ? [env.VITE_DEV_ALLOWED_HOST] : []),
  ]

  return {
    plugins: [
      react(),
      tailwindcss(),
      // Installability + the offline shell only. The precache is deliberately narrow (see
      // README's "PWA / install" section): everything a phone/webview can reach lazily must
      // stay out of it, and no API response is ever cached. `registerType: 'autoUpdate'` makes
      // a new deploy reload an open app silently.
      VitePWA({
        registerType: 'autoUpdate',
        // Registration is explicit in src/main.tsx (src/lib/pwa.ts), not injected here.
        injectRegister: false,
        // `devOptions` serves a dev manifest/worker so both are inspectable in `npm run dev`;
        // real verification still happens against `npm run preview` of a production build.
        devOptions: { enabled: true, type: 'module' },
        manifest: {
          id: '/',
          name: 'Joint Wallet Watch',
          short_name: 'JWW',
          description: 'Shared income and expense tracking.',
          start_url: '/',
          scope: '/',
          display: 'standalone',
          orientation: 'any',
          // Hex literals, not CSS custom properties — a manifest cannot read var(--paper).
          // Keep in sync with `--paper` in src/index.css.
          theme_color: '#121212',
          background_color: '#121212',
          lang: 'en',
          dir: 'ltr',
          categories: ['finance'],
          // `prefer_related_applications` is intentionally absent (Chromium install criterion).
          icons: [
            { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
            { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
            { src: '/icons/icon-192-maskable.png', sizes: '192x192', type: 'image/png', purpose: 'maskable' },
            { src: '/icons/icon-512-maskable.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
          ],
        },
        workbox: {
          // Shell only. No page/layout chunk, so the "a phone never downloads the desktop
          // chunk" rule holds and a broader glob can't quietly fetch every route at install.
          // The manifest's own icons are precached automatically (includeManifestIcons); only
          // the apple-touch icon, which the manifest does not name, is globbed here.
          globPatterns: ['index.html', 'assets/app-*.js', 'assets/index-*.css', 'favicon.svg', 'icons/apple-touch-icon.png'],
          navigateFallback: 'index.html',
          navigateFallbackDenylist: [/^\/api\//],
          cleanupOutdatedCaches: true,
          // No `runtimeCaching` block at all: that absence is the guarantee that no API
          // response (transactions, session, balances) is ever stored — see the request-class
          // contract in TESTING.md's "PWA install & offline" section.
        },
      }),
    ],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      // Reachable through a tunnel (ngrok) as well as localhost.
      allowedHosts,
      // The app calls same-origin `/api/v1/*` in dev; this forwards it to the backend.
      proxy: apiProxy,
      // When the dev server is reached through a tunnel, HMR's WebSocket must connect on the
      // tunnel's port (e.g. 443), not the local Vite port. Set VITE_DEV_HMR_CLIENT_PORT=443 to
      // keep live reload working over ngrok.
      ...(env.VITE_DEV_HMR_CLIENT_PORT ? { hmr: { clientPort: Number(env.VITE_DEV_HMR_CLIENT_PORT) } } : {}),
    },
    preview: {
      // Same as above, for `npm run preview` of a production build (the PWA verification path).
      allowedHosts,
      proxy: apiProxy,
    },
    build: {
      rollupOptions: {
        output: {
          // A distinct entry filename so the workbox glob below can precache exactly the shell.
          // Every route's `pages/*/index.tsx` also emits `assets/index-*.js`, so a generic
          // `assets/index-*.js` glob would silently precache lazy page chunks too.
          entryFileNames: 'assets/app-[hash].js',
        },
      },
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
