# public/

Static assets served as-is (unprocessed by Vite). Referenced by absolute path (e.g. `/icons/icon-192.png`)
from `index.html`, the web app manifest, or CSS — not imported from `src/`.

For icons used inside components, use `src/components/icons.tsx` instead — this folder is only
for assets that must be served verbatim at a fixed URL.

## Icons

- `favicon.svg` — **the source art** for every app icon, and the site favicon. A 512-unit canvas:
  the loader's coin, the `IconWallet` body/flap/dot, and the `JWW` monogram underneath, dark on the
  accent field (mirrors `--on-accent` on `--accent`). If the brand mark changes, edit this file.
- `icons/icon-192.png`, `icons/icon-512.png` — `purpose: "any"` manifest icons (Chromium requires
  192 and 512).
- `icons/icon-192-maskable.png`, `icons/icon-512-maskable.png` — `purpose: "maskable"`. Same art;
  the lockup is drawn inside the 80% safe zone. Declared as a **separate** manifest entry, never
  `"any maskable"` (Lighthouse flags that).
- `icons/apple-touch-icon.png` — 180×180, opaque and unrounded (iOS applies its own mask). Not
  named by the manifest, so it is linked manually from `index.html`.

**PNGs are committed; do not add an image-generation step to the build.** Regenerate them by hand
from `favicon.svg` (e.g. `rsvg-convert -w 192 -h 192 -o public/icons/icon-192.png public/favicon.svg`)
and commit the result. A build-time image pipeline would be a new dependency and a new failure
mode for artwork that changes roughly never.

(An unused `icons.svg` — a leftover Vite scaffold sprite with GitHub/Discord/Bluesky icons from
the template, never referenced anywhere in this app — was removed. If you find yourself wanting
a static sprite sheet again, check first whether `components/icons.tsx` already covers it.)
