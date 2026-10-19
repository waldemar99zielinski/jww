# internal/adapter/imageprep

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements `transaction_attachment.ContentNormalizer`.

**Role.** Normalizes uploaded images before object storage. HEIC/HEIF uploads are decoded and re-encoded as bounded PNG so every stored image can be rendered by browsers and fetched/decoded by the configured vision provider. Non-HEIC uploads pass through as the original stream.

## Behavior

- Detection uses the parsed media type (`image/heic`, `image/heif`) and falls back to the `.heic`/`.heif` extension.
- HEIC/HEIF is decoded in pure Go via `github.com/gen2brain/h265/heic`, auto-rotated, and limited to a 48 MP source raster.
- The long edge is scaled to at most **2000 px** using Catmull-Rom sampling, then encoded as PNG.
- The stored filename becomes `*.png` and the stored media type becomes `image/png`.
- Decode/encode failures are returned to the service, which maps them to a validation error; nothing is stored.
- Other files are returned unchanged without buffering or reading their bodies.

## Rules

- Clients upload originals; conversion and resizing are server-owned.
- Never inspect or log attachment bytes here.
- Keep conversion outside the service layer: the service depends only on `ContentNormalizer`.

## Testing

`imageprep_test.go` generates a HEIC in-process, asserts the PNG filename/media type and 2000 px bound, verifies non-HEIC uploads stream through unchanged, and rejects corrupt HEIC input.
