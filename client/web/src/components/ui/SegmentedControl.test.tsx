import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { SegmentedControl } from '@/components/ui/SegmentedControl'

const options = [
  { label: 'Expense', value: 'expense' },
  { label: 'Income', value: 'income' },
] as const

describe('SegmentedControl', () => {
  it("calls onChange with the clicked option's value", async () => {
    const onChange = vi.fn()
    render(<SegmentedControl value="expense" options={options} onChange={onChange} />)

    await userEvent.click(screen.getByRole('button', { name: 'Income' }))

    expect(onChange).toHaveBeenCalledWith('income')
  })

  it('renders one button per option', () => {
    render(<SegmentedControl value="expense" options={options} onChange={() => {}} />)
    expect(screen.getAllByRole('button')).toHaveLength(2)
  })
})
