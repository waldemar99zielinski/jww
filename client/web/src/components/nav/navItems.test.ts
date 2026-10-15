import { describe, expect, it } from 'vitest'
import { isNavItemActive } from '@/components/nav/navItems'

describe('isNavItemActive', () => {
  it('matches only an exact path for an "end" item', () => {
    expect(isNavItemActive('/', { to: '/', end: true })).toBe(true)
    expect(isNavItemActive('/transactions', { to: '/', end: true })).toBe(false)
  })

  it('matches an exact path or any nested path for a non-"end" item', () => {
    expect(isNavItemActive('/settings', { to: '/settings' })).toBe(true)
    expect(isNavItemActive('/settings/categories', { to: '/settings' })).toBe(true)
    expect(isNavItemActive('/settings-other', { to: '/settings' })).toBe(false)
  })

  it('does not match an unrelated path', () => {
    expect(isNavItemActive('/transactions', { to: '/settings' })).toBe(false)
  })
})
