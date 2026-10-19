/**
 * Triggers a browser save for already-fetched bytes. The download API routes are protected, so
 * callers fetch a Blob with the auth header and hand it here rather than using a plain `<a href>`.
 */
export function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  // Revoke on the next tick so the browser has started reading the blob before it's freed.
  setTimeout(() => URL.revokeObjectURL(url), 0)
}
