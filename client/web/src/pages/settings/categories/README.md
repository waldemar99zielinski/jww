# pages/settings/categories/

Manage transaction categories: switch between expense/income type, list categories (including
inactive ones — `useCategories(type, false)`), toggle a category active/inactive, and create a
new one. Standard container/view structure (see `pages/README.md`).

## Things to know

- `handleToggleActive` re-sends the category's existing `name`/`sort_order` alongside the
  flipped `active` flag, since the update API takes the full editable field set rather than a
  partial patch — see `hooks/useCategories.ts`'s `useUpdateCategory`.
- Category type (`expense`/`income`) is local `useState` in `index.tsx`, not part of the
  category list query's result — switching it re-queries `useCategories` with the new type.

## Testing

- Container: MSW-mock `useCategories`/`useCreateCategory`/`useUpdateCategory`; the two behaviors
  worth pinning down with a test are exactly the two called out above — `handleToggleActive`
  resending the existing `name`/`sort_order` alongside the flipped `active` flag, and switching
  the expense/income type re-querying with the new `type` rather than filtering client-side.
- This page owns the **"can a user manage categories"** Playwright flow from `TESTING.md`
  (create/rename/deactivate, then see it reflected in the Add Transaction category picker).

## Don't

- Don't pass a category's `type` into `useUpdateCategory`'s mutation input — it isn't part of
  what gets updated (category type can't change after creation); the hook derives everything it
  needs from the mutation response.
