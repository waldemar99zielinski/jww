import { useEffect, useRef, useState, type TouchEvent } from 'react'
import { downloadTransactionAttachment } from '@/api/attachments'
import { IconClose, IconDownload, IconSpinner } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { useScrollLock } from '@/hooks/useScrollLock'
import { formatBytes } from '../formatBytes'
import { previewKindFor } from '../previewKind'
import type { TransactionAttachmentItem } from '../types'

export interface AttachmentPreviewProps {
  transactionId: number
  attachment: TransactionAttachmentItem
  isDownloading: boolean
  onDownload: (attachment: TransactionAttachmentItem) => void
  onClose: () => void
}

/** Don't commit to a vertical drag until it clearly beats the horizontal axis. */
const DIRECTION_LOCK_PX = 12
/** A downward drag must cross this fraction of the viewport height (capped) to dismiss. */
const CLOSE_THRESHOLD_RATIO = 0.2
const CLOSE_THRESHOLD_MAX_PX = 140
/** A hair longer than the exit transition, so it finishes before we unmount. */
const EXIT_MS = 220

/**
 * Full-screen lightbox for one attachment. Images and PDFs render straight from the server's inline
 * `preview_grant` URL (no extra request); video, audio and text are pulled as a blob through the
 * protected download route, because those types don't get a grant. Anything the browser can't render
 * degrades to a "download it instead" message. Closes on the X, Escape, a backdrop tap, or a
 * downward swipe.
 *
 * The overlay is mounted per attachment (the container keys it by id), so it never has to switch
 * sources mid-life and the resolved state can be derived from props + fetch results during render.
 */
export function AttachmentPreview({
  transactionId,
  attachment,
  isDownloading,
  onDownload,
  onClose,
}: AttachmentPreviewProps) {
  const kind = previewKindFor(attachment.mediaType)
  // Images and PDFs are grant-backed, so their source needs no fetch.
  const grantSrc = kind === 'image' || kind === 'pdf' ? attachment.previewUrl : null

  // Blob-backed state, only ever written from the async fetch callbacks below.
  const [blobSrc, setBlobSrc] = useState<string | null>(null)
  const [blobText, setBlobText] = useState<string | null>(null)
  const [failed, setFailed] = useState(false)
  // The source being set isn't the same as the bytes being shown: an <img>/<video>/<iframe> still
  // has to fetch and decode before there's anything to see.
  const [mediaLoaded, setMediaLoaded] = useState(false)

  const src = grantSrc ?? blobSrc
  const text = blobText
  const status: 'loading' | 'ready' | 'error' =
    !kind || failed ? 'error' : src !== null || text !== null ? 'ready' : 'loading'

  const needsMediaLoad = kind === 'image' || kind === 'pdf' || kind === 'video' || kind === 'audio'
  const showLoader = status === 'loading' || (status === 'ready' && needsMediaLoad && !mediaLoaded)

  // Swipe-down-to-dismiss (touch only). A short drag springs back; a long one animates out first.
  const [offsetY, setOffsetY] = useState(0)
  const [dragging, setDragging] = useState(false)
  const [closing, setClosing] = useState(false)
  const gesture = useRef<{ startX: number; startY: number; vertical: boolean } | null>(null)
  const offsetRef = useRef(0)
  const onCloseRef = useRef(onClose)
  useEffect(() => {
    onCloseRef.current = onClose
  })

  useEffect(() => {
    if (!closing) return
    const timer = window.setTimeout(() => onCloseRef.current(), EXIT_MS)
    return () => window.clearTimeout(timer)
  }, [closing])

  function handleTouchStart(event: TouchEvent<HTMLDivElement>) {
    if (closing) return
    const touch = event.touches[0]
    if (!touch) return
    gesture.current = { startX: touch.clientX, startY: touch.clientY, vertical: false }
    offsetRef.current = 0
  }

  function handleTouchMove(event: TouchEvent<HTMLDivElement>) {
    const current = gesture.current
    if (!current || closing) return
    const touch = event.touches[0]
    if (!touch) return
    const dx = touch.clientX - current.startX
    const dy = touch.clientY - current.startY

    if (!current.vertical) {
      if (Math.abs(dx) < DIRECTION_LOCK_PX && Math.abs(dy) < DIRECTION_LOCK_PX) return
      // A mostly-horizontal drag isn't ours; abandon it.
      if (Math.abs(dx) >= Math.abs(dy)) {
        gesture.current = null
        return
      }
      current.vertical = true
      setDragging(true)
    }

    // Only downward movement moves the sheet; upward drags stay pinned.
    const next = Math.max(0, dy)
    offsetRef.current = next
    setOffsetY(next)
  }

  function handleTouchEnd() {
    const current = gesture.current
    gesture.current = null
    setDragging(false)
    if (!current?.vertical) {
      offsetRef.current = 0
      setOffsetY(0)
      return
    }
    const threshold = Math.min(CLOSE_THRESHOLD_MAX_PX, window.innerHeight * CLOSE_THRESHOLD_RATIO)
    if (offsetRef.current > threshold) {
      setClosing(true)
      return
    }
    offsetRef.current = 0
    setOffsetY(0)
  }

  // Fetch the bytes only when there's no grant URL (video/audio/text, plus unlisted image types).
  useEffect(() => {
    if (!kind || grantSrc) return
    let objectUrl: string | null = null
    let cancelled = false
    downloadTransactionAttachment(transactionId, attachment.id)
      .then(async (blob) => {
        if (cancelled) return
        if (kind === 'text') {
          setBlobText(await blob.text())
        } else {
          objectUrl = URL.createObjectURL(blob)
          setBlobSrc(objectUrl)
        }
      })
      .catch(() => {
        if (!cancelled) setFailed(true)
      })
    return () => {
      cancelled = true
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [attachment.id, grantSrc, kind, transactionId])

  // Escape closes; the background page (including the desktop shell's scroll container) is frozen.
  useScrollLock()
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  const transform = closing ? 'translateY(100%)' : dragging ? `translateY(${offsetY}px)` : 'translateY(0px)'

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={`Preview of ${attachment.filename}`}
      className={`fixed inset-0 z-50 flex flex-col bg-black/85 [touch-action:pan-y] ${
        dragging ? 'transition-none' : 'transition-transform duration-200 ease-out'
      }`}
      style={{ transform }}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      onTouchCancel={handleTouchEnd}
    >
      <div className="absolute inset-0" onClick={onClose} role="presentation" />

      <header className="relative z-10 flex items-center justify-between gap-3 px-4 py-3">
        <div className="min-w-0">
          <p className="m-0 truncate text-sm font-semibold text-white">{attachment.filename}</p>
          <p className="m-0 truncate text-xs text-white/60">
            {formatBytes(attachment.byteSize)} · {attachment.mediaType}
          </p>
        </div>
        <div className="flex flex-none items-center gap-2">
          <Button
            variant="custom"
            onClick={() => onDownload(attachment)}
            disabled={isDownloading}
            aria-label={`Download ${attachment.filename}`}
            className="flex items-center gap-1.5 rounded-[10px] border border-white/20 bg-white/10 px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-40"
          >
            <IconDownload className="h-3.5 w-3.5" />
            {isDownloading ? 'Saving…' : 'Download'}
          </Button>
          <Button
            variant="custom"
            aria-label="Close preview"
            onClick={onClose}
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-white/20 bg-white/10 text-white"
          >
            <IconClose className="h-4 w-4" />
          </Button>
        </div>
      </header>

      <div
        className="relative z-10 flex min-h-0 flex-1 items-center justify-center overscroll-contain p-4"
        onClick={onClose}
      >
        {status === 'error' && (
          <p className="m-0 max-w-sm text-center text-sm text-white/80">
            This file can't be previewed in your browser — download it instead.
          </p>
        )}

        {status === 'ready' && kind === 'image' && src && (
          <img
            src={src}
            alt={`Preview of ${attachment.filename}`}
            onLoad={() => setMediaLoaded(true)}
            onError={() => setFailed(true)}
            onClick={(event) => event.stopPropagation()}
            className="max-h-full max-w-full rounded-lg object-contain"
          />
        )}

        {status === 'ready' && kind === 'pdf' && src && (
          <iframe
            src={src}
            title={`Preview of ${attachment.filename}`}
            onLoad={() => setMediaLoaded(true)}
            onClick={(event) => event.stopPropagation()}
            className="h-full w-full rounded-lg bg-white"
          />
        )}

        {status === 'ready' && kind === 'video' && src && (
          <video
            src={src}
            controls
            aria-label={`Preview of ${attachment.filename}`}
            onLoadedData={() => setMediaLoaded(true)}
            onError={() => setFailed(true)}
            onClick={(event) => event.stopPropagation()}
            className="max-h-full max-w-full rounded-lg"
          />
        )}

        {status === 'ready' && kind === 'audio' && src && (
          <audio
            src={src}
            controls
            aria-label={`Preview of ${attachment.filename}`}
            onLoadedData={() => setMediaLoaded(true)}
            onError={() => setFailed(true)}
            onClick={(event) => event.stopPropagation()}
            className="w-full max-w-md"
          />
        )}

        {status === 'ready' && kind === 'text' && (
          <pre
            onClick={(event) => event.stopPropagation()}
            className="m-0 h-full w-full overflow-auto whitespace-pre-wrap break-words rounded-lg bg-white/95 p-4 text-left text-xs leading-relaxed text-ink"
          >
            {text ?? ''}
          </pre>
        )}

        {showLoader && (
          <div
            role="status"
            aria-label="Loading preview"
            className="pointer-events-none absolute inset-0 flex items-center justify-center"
          >
            <IconSpinner className="h-8 w-8 animate-spin text-white/70" />
          </div>
        )}
      </div>
    </div>
  )
}
