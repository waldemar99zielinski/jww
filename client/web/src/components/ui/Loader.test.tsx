import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Loader } from '@/components/ui/Loader'

describe('Loader', () => {
  it('exposes role="status" with an accessible "Loading" label', () => {
    render(<Loader />)
    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument()
  })
})
