import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { BottomNav } from '@/components/nav/BottomNav'
import { renderWithProviders } from '@/test/renderWithProviders'

describe('BottomNav', () => {
  it('marks the tab matching the current route as active', () => {
    renderWithProviders(<BottomNav />, { route: '/' })

    expect(screen.getByRole('link', { name: /home/i })).toHaveClass('text-accent')
    expect(screen.getByRole('link', { name: /history/i })).toHaveClass('text-muted')
  })

  it('marks a nested route as active for its parent nav item', () => {
    renderWithProviders(<BottomNav />, { route: '/transactions' })

    expect(screen.getByRole('link', { name: /history/i })).toHaveClass('text-accent')
    expect(screen.getByRole('link', { name: /home/i })).toHaveClass('text-muted')
  })

  it('renders a standalone Add action', () => {
    renderWithProviders(<BottomNav />, { route: '/' })
    expect(screen.getByRole('link', { name: 'Add transaction' })).toHaveAttribute('href', '/add')
  })

  it('grows by the shared safe-area gap so the tabs stay clear of the iPhone home indicator', () => {
    renderWithProviders(<BottomNav />, { route: '/' })

    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('h-[var(--bottom-nav-height)]')
    expect(nav).toHaveClass('box-content')
    expect(nav).toHaveClass('pb-[var(--bottom-nav-bottom-gap)]')
  })
})
