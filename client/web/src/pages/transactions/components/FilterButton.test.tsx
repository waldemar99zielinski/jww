import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { FilterButton } from '@/pages/transactions/components/FilterButton'

describe('FilterButton', () => {
  it('renders without a count badge when no filters are active', () => {
    render(<FilterButton compact={false} activeCount={0} open={false} onClick={vi.fn()} />)

    expect(screen.getByRole('button', { name: /filters/i })).toBeInTheDocument()
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('renders the active filter count when filters are applied', () => {
    render(<FilterButton compact={false} activeCount={3} open={false} onClick={vi.fn()} />)

    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('reflects the open state via aria-expanded', () => {
    render(<FilterButton compact={false} activeCount={0} open onClick={vi.fn()} />)

    expect(screen.getByRole('button', { name: /filters/i })).toHaveAttribute('aria-expanded', 'true')
  })
})
