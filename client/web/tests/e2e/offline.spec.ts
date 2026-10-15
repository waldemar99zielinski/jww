import { expect, test } from '@playwright/test'
import { loginViaUi } from './support/auth'
import { getTestUser } from './support/testUser'

test.describe('offline', () => {
  test('shows a blocking offline screen, keeps the session, and recovers with focus restored', async ({
    page,
    context,
  }) => {
    await loginViaUi(page, getTestUser())
    await expect(page).toHaveURL('/')

    // Focus a real control so focus restoration is observable after reconnecting.
    const target = page.getByRole('link', { name: 'Settings' }).first()
    await target.focus()

    await context.setOffline(true)

    const dialog = page.getByRole('alertdialog', { name: /offline/i })
    await expect(dialog).toBeVisible()
    await expect(page.getByRole('button', { name: 'Retry' })).toBeFocused()
    // The app underneath is inert, so it is neither clickable nor focusable.
    await expect(page.locator('[inert]')).toHaveCount(1)

    await context.setOffline(false)

    await expect(dialog).toHaveCount(0)
    await expect(target).toBeFocused()
    await expect(page.locator('p[role="status"]')).toHaveText('Back online')
  })

  test('covers the login route too', async ({ page, context }) => {
    await page.goto('/login')

    await context.setOffline(true)

    await expect(page.getByRole('alertdialog', { name: /offline/i })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Retry' })).toBeFocused()
  })
})
