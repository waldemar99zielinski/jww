import { expect, test } from '@playwright/test'
import { getAccessToken } from './support/apiAuth'
import { loginViaUi } from './support/auth'
import { findCategory } from './support/categories'
import { ensureCurrency } from './support/currencies'
import { getTestUser, uniqueLabel } from './support/testUser'
import { createTransaction, getCurrentUserId } from './support/transactions'

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

  test('can a user scan an attachment and get a prefilled form with editable line items?', async ({ page, request }) => {
    const user = getTestUser()
    const token = await getAccessToken(request, user)
    await ensureCurrency(request, token)

    const scannedName = uniqueLabel('e2e-prefill')

    // A real read is a live, non-deterministic vision call, so stub only the extraction route;
    // the upload, category lookup and create still hit the real backend.
    await page.route('**/api/v1/transaction-attachments/extraction', async (route) => {
      const body = route.request().postDataJSON() as { attachment_ids: string[] }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            attachment_id: body.attachment_ids[0],
            attachment_ids: body.attachment_ids,
            type: { value: 'expense', confidence: 0.9 },
            name: { value: scannedName, confidence: 0.9, source: 'merchant header' },
            category_hint: { value: 'Groceries', confidence: 0.9 },
            sub_transactions: [{ name: 'Coffee beans', label: 'pantry', amount: '8.50' }],
          },
        }),
      })
    })

    await loginViaUi(page, user)
    await page.goto('/add')

    // Attach an image, then read it.
    await page.getByRole('button', { name: 'Add attachment' }).click()
    await page
      .locator('input[type="file"]')
      .nth(1)
      .setInputFiles({ name: 'attachment.png', mimeType: 'image/png', buffer: Buffer.from('fake-image-bytes') })
    await page.getByRole('button', { name: 'Scan attachment' }).click()

    // A successful read fills the main form and preloads the extracted line item.
    await expect(page.getByLabel('Transaction name')).toHaveValue(scannedName)
    await expect(page.getByText('1 item added')).toBeVisible()
    await expect(page.getByText('Coffee beans').first()).toBeVisible()
    await expect(page.getByText('= sum of 1 item')).toBeVisible()
    await expect(page.getByRole('button', { name: /groceries/i })).toBeVisible()

    // The drawer is the editing surface: add one more line item, then close it.
    await page.getByRole('button', { name: 'Edit sub-transactions' }).click()
    await page.getByRole('button', { name: 'Add item' }).click()
    await page.getByLabel('New item name').fill('Oat milk')
    await page.getByLabel('New item amount').fill('3,00')
    await page.getByRole('button', { name: 'Add item' }).click()
    await page.getByRole('button', { name: 'Done' }).click()

    await expect(page.getByText('= sum of 2 items')).toBeVisible()
    await expect(page.getByText('11,50', { exact: true }).first()).toBeVisible()

    // The prefilled form is immediately savable and round-trips to the backend.
    await page.getByRole('button', { name: /save transaction/i }).click()

    await expect(page).toHaveURL('/')
    await page.goto('/transactions')
    await expect(page.getByText(scannedName)).toBeVisible()
  })

  test('can a user add a manual line item and see the derived read-only amount?', async ({ page, request }) => {
    // Line items are entered on the transaction detail edit screen; the server recomputes the
    // parent amount as the sum of its items, so the Amount field becomes a read-only derived value.
    // No route stubbing here — every call is real against the backend.
    const user = getTestUser()
    const token = await getAccessToken(request, user)
    const userId = await getCurrentUserId(request, token)
    const currency = await ensureCurrency(request, token)
    const groceries = await findCategory(request, token, 'expense', 'Groceries')

    const transactionName = uniqueLabel('e2e-itemised')
    const transaction = await createTransaction(request, token, {
      type: 'expense',
      amount: 10,
      currency_code: currency.code,
      category_id: groceries.id,
      user_id: userId,
      source: 'manual',
      name: transactionName,
    })

    await loginViaUi(page, user)
    await page.goto(`/transactions/${transaction.id}`)
    await expect(page.getByText(transactionName)).toBeVisible()

    await page.getByRole('button', { name: 'Edit' }).click()

    // A transaction with no items starts with an editable Amount and an empty item list.
    await expect(page.getByLabel('Amount')).toBeVisible()
    await page.getByRole('button', { name: 'Add item' }).click()
    await page.getByLabel('New item name').fill('Coffee beans')
    await page.getByLabel('New item amount').fill('8,50')
    await page.getByRole('button', { name: 'Add item' }).click()

    // The row persists immediately (POST) and the parent amount becomes the derived, read-only Σ.
    await expect(page.getByText('Coffee beans')).toBeVisible()
    await expect(page.getByText('= sum of 1 item')).toBeVisible()
    await expect(page.getByLabel('Amount')).toHaveCount(0)
    await expect(page.getByText('8,50', { exact: true }).first()).toBeVisible()
  })
})
