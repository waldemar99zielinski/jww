import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { DateInput } from '@/components/ui/DateInput'

describe('DateInput', () => {
  it('calls onChange with the new date value', () => {
    const onChange = vi.fn()
    render(<DateInput label="Date" value="2026-01-01" onChange={onChange} />)

    fireEvent.change(screen.getByLabelText('Date'), { target: { value: '2026-02-15' } })

    expect(onChange).toHaveBeenCalledWith('2026-02-15')
  })

  it('focuses the input when clicking anywhere in the field', async () => {
    render(<DateInput label="Date" value="2026-01-01" onChange={() => {}} />)

    const input = screen.getByLabelText('Date')
    await userEvent.click(input)

    expect(document.activeElement).toBe(input)
  })

  it('marks the input invalid via aria-invalid when invalid is true', () => {
    render(<DateInput label="Date" value="2026-01-01" onChange={() => {}} invalid />)
    expect(screen.getByLabelText('Date')).toHaveAttribute('aria-invalid', 'true')
  })
})
