import { expect, test } from '@playwright/test'
import { getAccessToken } from './support/apiAuth'
import { loginViaUi } from './support/auth'
import { findCategory } from './support/categories'
import { ensureCurrency } from './support/currencies'
import { getTestUser, uniqueLabel } from './support/testUser'
import { createTransaction, getCurrentUserId } from './support/transactions'

test.describe('transactions', () => {
  test('can a user list/view their transactions, and does a filter narrow the list?', async ({ page, request }) => {
    const user = getTestUser()
    const token = await getAccessToken(request, user)
    const userId = await getCurrentUserId(request, token)
    const currency = await ensureCurrency(request, token)
    const groceries = await findCategory(request, token, 'expense', 'Groceries')
    const salary = await findCategory(request, token, 'income', 'Salary')

    const expenseName = uniqueLabel('e2e-expense')
    const incomeName = uniqueLabel('e2e-income')

    await createTransaction(request, token, {
      type: 'expense',
      amount: 12.34,
      currency_code: currency.code,
      category_id: groceries.id,
      user_id: userId,
      source: 'manual',
      name: expenseName,
    })
    await createTransaction(request, token, {
      type: 'income',
      amount: 500,
      currency_code: currency.code,
      category_id: salary.id,
      user_id: userId,
      source: 'manual',
      name: incomeName,
    })

    await loginViaUi(page, user)
    await page.goto('/transactions')

    await expect(page.getByText(expenseName)).toBeVisible()
    await expect(page.getByText(incomeName)).toBeVisible()

    await page.getByRole('button', { name: /filters/i }).click()
    await page.getByLabel('Name').fill(incomeName)
    await page.getByRole('button', { name: /apply/i }).click()

    await expect(page.getByText(incomeName)).toBeVisible()
    await expect(page.getByText(expenseName)).toHaveCount(0)
  })
})
