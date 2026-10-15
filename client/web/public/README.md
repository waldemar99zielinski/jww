# public/

Static assets served as-is (unprocessed by Vite) — currently just `favicon.svg`. Referenced by
absolute path (e.g. `/favicon.svg`) from `index.html`/CSS, not imported from `src/`.

For icons used inside components, use `src/components/icons.tsx` instead — this folder is only
for assets that must be served verbatim at a fixed URL.

(An unused `icons.svg` — a leftover Vite scaffold sprite with GitHub/Discord/Bluesky icons from
the template, never referenced anywhere in this app — was removed. If you find yourself wanting
a static sprite sheet again, check first whether `components/icons.tsx` already covers it.)
