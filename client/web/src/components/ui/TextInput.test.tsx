import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createRef } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { TextInput } from '@/components/ui/TextInput'

describe('TextInput', () => {
  it('reflects typed input via onChange', async () => {
    const onChange = vi.fn()
    render(<TextInput aria-label="Name" value="" onChange={onChange} />)

    await userEvent.type(screen.getByRole('textbox', { name: 'Name' }), 'a')

    expect(onChange).toHaveBeenCalled()
  })

  it('renders a leading icon when leftIcon is given', () => {
    render(<TextInput aria-label="Amount" value="" onChange={() => {}} leftIcon={<svg data-testid="left-icon" />} />)
    expect(screen.getByTestId('left-icon')).toBeInTheDocument()
  })

  it('forwards the ref to the underlying input element', () => {
    const ref = createRef<HTMLInputElement>()
    render(<TextInput aria-label="Name" value="" onChange={() => {}} ref={ref} />)
    expect(ref.current).toBeInstanceOf(HTMLInputElement)
  })
})
