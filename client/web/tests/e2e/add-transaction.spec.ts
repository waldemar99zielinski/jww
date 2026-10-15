import { expect, test } from '@playwright/test'
import { getAccessToken } from './support/apiAuth'
import { loginViaUi } from './support/auth'
import { ensureCurrency } from './support/currencies'
import { getTestUser, uniqueLabel } from './support/testUser'

test.describe('add transaction', () => {
  test('can a user add a transaction, and does it appear in the transaction list?', async ({ page, request }) => {
    const user = getTestUser()
    const token = await getAccessToken(request, user)
    await ensureCurrency(request, token)

    const transactionName = uniqueLabel('e2e-add')

    await loginViaUi(page, user)
    await page.goto('/add')

    await page.getByLabel('Amount').fill('19,99')
    await page.getByLabel('Transaction name').fill(transactionName)

    await page.getByRole('button', { name: 'Select category' }).click()
    await page.getByRole('button', { name: /groceries/i }).click()

    await page.getByRole('button', { name: /save transaction/i }).click()

    await expect(page).toHaveURL('/')

    await page.goto('/transactions')
    await expect(page.getByText(transactionName)).toBeVisible()
  })
})
