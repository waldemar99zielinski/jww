import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { StatusBanner } from '@/components/StatusBanner'

describe('StatusBanner', () => {
  it('renders the "not built yet" message', () => {
    render(<StatusBanner />)
    expect(screen.getByText('Proposed — not built yet')).toBeInTheDocument()
  })
})
