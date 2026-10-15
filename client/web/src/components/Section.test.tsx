import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { Section } from '@/components/Section'

describe('Section', () => {
  it('renders the title and children', () => {
    render(
      <MemoryRouter>
        <Section title="Recent transactions">
          <p>Content</p>
        </Section>
      </MemoryRouter>,
    )

    expect(screen.getByText('Recent transactions')).toBeInTheDocument()
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('renders an action link when given', () => {
    render(
      <MemoryRouter>
        <Section title="Recent transactions" action={{ label: 'See all', to: '/transactions' }}>
          <p>Content</p>
        </Section>
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'See all' })).toHaveAttribute('href', '/transactions')
  })

  it('renders no action link when none is given', () => {
    render(
      <MemoryRouter>
        <Section title="Recent transactions">
          <p>Content</p>
        </Section>
      </MemoryRouter>,
    )

    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})
