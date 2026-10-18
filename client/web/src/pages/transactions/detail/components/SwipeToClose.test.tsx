import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { SwipeToClose } from './SwipeToClose'

function renderSwipe(onClose: () => void) {
  return render(
    <SwipeToClose onClose={onClose}>
      {(close) => <button onClick={close}>Content</button>}
    </SwipeToClose>,
  )
}

/** Drag the (bubbling) content element so the gesture handlers on the wrapper see it. */
function drag(from: { x: number; y: number }, to: { x: number; y: number }) {
  const handle = screen.getByRole('button', { name: 'Content' })
  fireEvent.touchStart(handle, { touches: [{ clientX: from.x, clientY: from.y }] })
  fireEvent.touchMove(handle, { touches: [{ clientX: to.x, clientY: to.y }] })
  fireEvent.touchEnd(handle, { changedTouches: [{ clientX: to.x, clientY: to.y }] })
}

describe('SwipeToClose', () => {
  afterEach(() => vi.useRealTimers())

  it('closes after a drag to the right past the threshold', () => {
    vi.useFakeTimers()
    const onClose = vi.fn()
    renderSwipe(onClose)

    drag({ x: 200, y: 40 }, { x: 380, y: 44 })

    expect(onClose).not.toHaveBeenCalled()
    vi.advanceTimersByTime(300)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('springs back without closing when the drag is short', () => {
    const onClose = vi.fn()
    renderSwipe(onClose)

    drag({ x: 200, y: 40 }, { x: 220, y: 42 })

    expect(onClose).not.toHaveBeenCalled()
  })

  it('abandons the gesture on a mostly vertical drag, so the page keeps scrolling', () => {
    const onClose = vi.fn()
    renderSwipe(onClose)

    drag({ x: 200, y: 40 }, { x: 206, y: 320 })

    expect(onClose).not.toHaveBeenCalled()
  })

  it('ignores drags starting at the screen edge (the OS back gesture owns those)', () => {
    const onClose = vi.fn()
    renderSwipe(onClose)

    drag({ x: 4, y: 40 }, { x: 300, y: 42 })

    expect(onClose).not.toHaveBeenCalled()
  })

  it('animates out and closes when a control calls close()', () => {
    vi.useFakeTimers()
    const onClose = vi.fn()
    renderSwipe(onClose)

    fireEvent.click(screen.getByRole('button', { name: 'Content' }))

    expect(onClose).not.toHaveBeenCalled()
    vi.advanceTimersByTime(300)
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
