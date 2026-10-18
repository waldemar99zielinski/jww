# public/

Static assets served as-is (unprocessed by Vite). Referenced by absolute path (e.g. `/icons/icon-192.png`)
from `index.html`, the web app manifest, or CSS — not imported from `src/`.

For icons used inside components, use `src/components/icons.tsx` instead — this folder is only
for assets that must be served verbatim at a fixed URL.

## Icons

- `icons/app-icon.svg` — **the source art** for the `purpose: "any"` icons. It matches
  `favicon.svg`: a warm-white wallet and coin on the rounded `#121212` field.
- `icons/app-icon-maskable.svg` — **the source art** for maskable icons and the Apple touch icon.
  It replaces the baked-in rounded tile with a full-bleed field and scales the mark to `0.9`, keeping
  it within the platform's 40%-radius safe circle. Keep its mark in sync with `app-icon.svg`.
- `favicon.svg` — the canonical browser favicon supplied for the app.
- `icons/icon-192.png`, `icons/icon-512.png` — `purpose: "any"` manifest icons (Chromium requires
  192 and 512).
- `icons/icon-192-maskable.png`, `icons/icon-512-maskable.png` — `purpose: "maskable"`, generated
  from `app-icon-maskable.svg`. Declared as a **separate** manifest entry, never `"any maskable"`
  (Lighthouse flags that).
- `icons/apple-touch-icon.png` — 180×180, opaque and full-bleed (iOS applies its own mask). Not
  named by the manifest, so it is linked manually from `index.html`.

**PNGs are committed; do not add an image-generation step to the build.** Regenerate them with
`node scripts/generate-icons.mjs` — a manual, non-build script that rasterises both masters with
`rsvg-convert` (or Chromium via `@playwright/test` when librsvg is absent) and then verifies sizes,
the mark's strokes, the maskable safe radius and the absence of text, fonts and motion. Run
`node scripts/generate-icons.mjs --check` to verify without writing. A build-time image pipeline
would be a new dependency and a new failure mode for artwork that changes roughly never.

(An unused `icons.svg` — a leftover Vite scaffold sprite with GitHub/Discord/Bluesky icons from
the template, never referenced anywhere in this app — was removed. If you find yourself wanting
a static sprite sheet again, check first whether `components/icons.tsx` already covers it.)
