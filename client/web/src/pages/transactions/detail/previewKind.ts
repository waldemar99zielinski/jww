/** The renderer the attachment lightbox should use, or `null` for types it can't preview. */
export type PreviewKind = 'image' | 'pdf' | 'video' | 'audio' | 'text' | null

/**
 * Classifies a media type into a preview renderer. Everything the browser can natively display —
 * images, PDF, video, audio, and text-like documents — maps to a kind; anything else (office
 * documents, archives, `application/octet-stream`, …) returns `null` and is download-only.
 *
 * Note this is intentionally broader than the server's inline `preview_grant` allowlist (which is
 * only images + PDF): non-grant types are fetched as a blob through the protected download route.
 */
export function previewKindFor(mediaType: string): PreviewKind {
  const type = mediaType.toLowerCase()
  if (type.startsWith('image/')) return 'image'
  if (type === 'application/pdf') return 'pdf'
  if (type.startsWith('video/')) return 'video'
  if (type.startsWith('audio/')) return 'audio'
  if (
    type.startsWith('text/') ||
    type === 'application/json' ||
    type.endsWith('+json') ||
    type === 'application/xml' ||
    type.endsWith('+xml')
  ) {
    return 'text'
  }
  return null
}
