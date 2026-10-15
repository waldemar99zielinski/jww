import { expect, test } from '@playwright/test'
import { loginViaUi } from './support/auth'
import { getTestUser } from './support/testUser'

test.describe('login', () => {
  test('can a user log in with valid credentials?', async ({ page }) => {
    await loginViaUi(page, getTestUser())

    await expect(page).toHaveURL('/')
    await expect(page.getByRole('button', { name: /sign in/i })).toHaveCount(0)
  })

  test('can a user NOT log in with the wrong password?', async ({ page }) => {
    const user = getTestUser()
    await loginViaUi(page, { username: user.username, password: `${user.password}-wrong` })

    await expect(page.getByText(/invalid username or password/i)).toBeVisible()
    await expect(page).toHaveURL(/\/login$/)
  })

  test('is an unauthenticated visit to a protected route redirected to login?', async ({ page }) => {
    await page.goto('/transactions')

    await expect(page).toHaveURL(/\/login$/)
  })
})
