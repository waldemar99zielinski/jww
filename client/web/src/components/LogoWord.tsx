const TAIL_EASE = 'cubic-bezier(0.65, 0, 0.25, 1)'
const TAIL_MS = 500

export interface LogoWordProps {
  word: string
  resolved: boolean
  last: boolean
}

/**
 * "Horizontal Unfold" (one-shot): each word is its first letter (always visible, never
 * shrinks) plus the rest — a "tail" that's either permanently expanded (every static reuse)
 * or transitioning from collapsed to expanded, simultaneously with the other two words (no
 * stagger), after a brief hold. Unlike the original spec, it does not fold back — once
 * unfolded, "Joint Wallet Watch" stays. aria-hidden since the parent's aria-label already
 * announces the full name once. Private to `Logo` — not meant to be used standalone.
 */
export function LogoWord({ word, resolved, last }: LogoWordProps) {
  const [first, ...rest] = word
  const restText = rest.join('')

  return (
    <span
      aria-hidden="true"
      className="inline-flex items-baseline"
      style={{
        marginRight: !last && resolved ? '0.25em' : 0,
        transition: `margin-right ${TAIL_MS}ms ease-in-out`,
      }}
    >
      <b className="shrink-0">{first}</b>
      <span
        className="inline-block overflow-hidden whitespace-nowrap"
        style={{
          maxWidth: resolved ? '2.3em' : 0,
          opacity: resolved ? 1 : 0,
          transform: resolved ? 'translateX(0)' : 'translateX(-2px)',
          transition: `max-width ${TAIL_MS}ms ${TAIL_EASE}, opacity ${TAIL_MS}ms ${TAIL_EASE}, transform ${TAIL_MS}ms ${TAIL_EASE}`,
        }}
      >
        {restText}
      </span>
    </span>
  )
}
