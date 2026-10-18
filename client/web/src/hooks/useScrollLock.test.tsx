import { render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { useScrollLock } from '@/hooks/useScrollLock'

function Consumer({ active = true }: { active?: boolean }) {
  useScrollLock(active)
  return null
}

describe('useScrollLock', () => {
  afterEach(() => {
    document.body.innerHTML = ''
    document.body.style.overflow = ''
    document.documentElement.style.overflow = ''
  })

  it('freezes the document and any data-scroll-lock element, then restores them', () => {
    const main = document.createElement('main')
    main.setAttribute('data-scroll-lock', '')
    main.style.overflow = 'auto'
    document.body.appendChild(main)

    const { unmount } = render(<Consumer />)

    expect(document.documentElement.style.overflow).toBe('hidden')
    expect(document.body.style.overflow).toBe('hidden')
    expect(main.style.overflow).toBe('hidden')

    unmount()

    expect(document.body.style.overflow).toBe('')
    expect(main.style.overflow).toBe('auto')
  })

  it('does nothing while inactive', () => {
    render(<Consumer active={false} />)

    expect(document.body.style.overflow).toBe('')
  })
})
