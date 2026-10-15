import { expect, test } from '@playwright/test'
import { loginViaUi } from './support/auth'
import { getTestUser, uniqueLabel } from './support/testUser'

test.describe('manage categories', () => {
  test('can a user create and deactivate a category, and is it reflected in Add Transaction?', async ({ page }) => {
    const user = getTestUser()
    const categoryName = uniqueLabel('e2e-cat')

    await loginViaUi(page, user)
    await page.goto('/settings/categories')

    await page.getByPlaceholder('e.g. Groceries').fill(categoryName)
    await page.getByRole('button', { name: 'Add' }).click()
    await expect(page.getByText(categoryName, { exact: true })).toBeVisible()

    await page.goto('/add')
    await page.getByRole('button', { name: 'Select category' }).click()
    await expect(page.getByRole('button', { name: new RegExp(categoryName, 'i') })).toBeVisible()

    await page.goto('/settings/categories')
    const row = page.getByText(categoryName, { exact: true }).locator('..')
    await row.getByRole('button', { name: /off/i }).click()
    await expect(row.getByRole('button', { name: /on/i })).toBeVisible()

    await page.goto('/add')
    await page.getByRole('button', { name: 'Select category' }).click()
    await expect(page.getByRole('button', { name: new RegExp(categoryName, 'i') })).toHaveCount(0)
  })
})
