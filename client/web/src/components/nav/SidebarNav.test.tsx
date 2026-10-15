import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { SidebarNav } from '@/components/nav/SidebarNav'
import { renderWithProviders } from '@/test/renderWithProviders'

describe('SidebarNav', () => {
  it('marks the link matching the current route as active', () => {
    renderWithProviders(<SidebarNav />, { route: '/settings' })

    expect(screen.getByRole('link', { name: /settings/i })).toHaveClass('bg-accent-soft')
    expect(screen.getByRole('link', { name: /home/i })).not.toHaveClass('bg-accent-soft')
  })

  it('marks a nested settings route as active for the Settings link', () => {
    renderWithProviders(<SidebarNav />, { route: '/settings/categories' })
    expect(screen.getByRole('link', { name: /settings/i })).toHaveClass('bg-accent-soft')
  })

  it('renders a standalone Add transaction action', () => {
    renderWithProviders(<SidebarNav />, { route: '/' })
    expect(screen.getByRole('link', { name: /add transaction/i })).toHaveAttribute('href', '/add')
  })
})
