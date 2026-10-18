function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

export function toISODate(date: Date): string {
  return date.toISOString().slice(0, 10)
}

/**
 * The date's calendar day *in the viewer's timezone*, as `YYYY-MM-DD`. Unlike `toISODate` (UTC,
 * used for "today" buckets), this is what a native date `<input>` must be seeded with so an
 * existing timestamp renders and round-trips on the day the user actually sees.
 */
export function toLocalISODate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

export function formatDayBucket(isoDateTime: string): string {
  const date = new Date(isoDateTime)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)

  if (isSameDay(date, today)) return 'Today'
  if (isSameDay(date, yesterday)) return 'Yesterday'
  return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })
}

export function formatTime(isoDateTime: string): string {
  return new Date(isoDateTime).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
}

export function formatShortDate(isoDate: string): string {
  return new Date(`${isoDate}T00:00:00`).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })
}

/**
 * Combines the date ("YYYY-MM-DD") and time ("HH:MM") picker values into an
 * RFC3339 instant, reading them as local wall-clock time. An empty time falls
 * back to midnight.
 */
export function toISODateTime(date: string, time: string): string {
  return new Date(`${date}T${time || '00:00'}:00`).toISOString()
}
