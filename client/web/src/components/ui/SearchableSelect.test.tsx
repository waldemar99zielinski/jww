import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { SearchableSelect } from '@/components/ui/SearchableSelect'

const options = [
  { value: 'cat-groceries', label: 'Groceries' },
  { value: 'cat-transport', label: 'Transport' },
]

describe('SearchableSelect', () => {
  it('shows the placeholder when nothing is selected', () => {
    render(<SearchableSelect value="" onChange={() => {}} options={options} placeholder="Select a category" />)
    expect(screen.getByRole('button', { name: 'Select a category' })).toBeInTheDocument()
  })

  it('opens the option list on click and calls onChange when an option is picked', async () => {
    const onChange = vi.fn()
    render(<SearchableSelect value="" onChange={onChange} options={options} placeholder="Select…" />)

    await userEvent.click(screen.getByRole('button', { name: 'Select…' }))
    await userEvent.click(screen.getByRole('button', { name: 'Transport' }))

    expect(onChange).toHaveBeenCalledWith('cat-transport')
  })

  it('filters options by the search query', async () => {
    render(
      <SearchableSelect value="" onChange={() => {}} options={options} placeholder="Select…" searchPlaceholder="Search…" />,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Select…' }))
    await userEvent.type(screen.getByPlaceholderText('Search…'), 'Trans')

    expect(screen.getByRole('button', { name: 'Transport' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Groceries' })).not.toBeInTheDocument()
  })

  it('shows the selected option label as the trigger text', () => {
    render(<SearchableSelect value="cat-groceries" onChange={() => {}} options={options} />)
    expect(screen.getByRole('button', { name: 'Groceries' })).toBeInTheDocument()
  })
})
