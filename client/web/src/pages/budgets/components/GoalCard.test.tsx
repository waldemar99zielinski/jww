import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { GoalCard } from './GoalCard'

describe('GoalCard', () => {
  it('renders the goal label and saved/target amounts', () => {
    render(<GoalCard label="Emergency fund" saved={2400} target={10000} />)

    expect(screen.getByText('Emergency fund')).toBeInTheDocument()
  })
})
