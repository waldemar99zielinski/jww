import { expect, test } from '@playwright/test'
import { loginViaUi } from './support/auth'
import { getTestUser, uniqueLabel } from './support/testUser'

function uniqueCurrencyCode(): string {
  const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  let code = ''
  for (let i = 0; i < 3; i++) code += letters[Math.floor(Math.random() * letters.length)]
  return code
}

test.describe('manage currencies', () => {
  test('can a user add a currency, and is it available where currency is chosen elsewhere?', async ({ page }) => {
    const user = getTestUser()
    const code = uniqueCurrencyCode()
    const name = uniqueLabel('E2E Currency')

    await loginViaUi(page, user)
    await page.goto('/settings/currencies')

    await page.getByPlaceholder('PLN').fill(code)
    await page.getByPlaceholder('Polish złoty').fill(name)
    await page.getByRole('button', { name: 'Add' }).click()

    await expect(page.getByText(code, { exact: true })).toBeVisible()

    await page.goto('/add')
    await expect(page.getByLabel('Currency').locator(`option[value="${code}"]`)).toHaveCount(1)
  })
})
