import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ProgressBar } from '@/components/ProgressBar'

describe('ProgressBar', () => {
  it('renders the fill width matching the given percent', () => {
    const { container } = render(<ProgressBar percent={42} />)
    const fill = container.querySelector('.bar-fill') as HTMLElement
    expect(fill.style.width).toBe('42%')
  })

  it('clamps a percent above 100 down to 100', () => {
    const { container } = render(<ProgressBar percent={150} />)
    const fill = container.querySelector('.bar-fill') as HTMLElement
    expect(fill.style.width).toBe('100%')
  })

  it('clamps a negative percent up to 0', () => {
    const { container } = render(<ProgressBar percent={-10} />)
    const fill = container.querySelector('.bar-fill') as HTMLElement
    expect(fill.style.width).toBe('0%')
  })

  it('applies the critical style when critical is true', () => {
    const { container } = render(<ProgressBar percent={90} critical />)
    expect(container.querySelector('.bar-fill.critical')).toBeInTheDocument()
  })
})
