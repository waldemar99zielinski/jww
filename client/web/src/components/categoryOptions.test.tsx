import { isValidElement } from 'react'
import { describe, expect, it } from 'vitest'
import type { TransactionCategoryDto } from '@/api/types'
import { toCategoryOptions } from './categoryOptions'

const categories: TransactionCategoryDto[] = [
  { id: 'cat-1', type: 'expense', name: 'Groceries', active: true, sort_order: 0 },
  { id: 'cat-2', type: 'expense', name: 'A Totally Unknown Category', active: true, sort_order: 1 },
]

describe('toCategoryOptions', () => {
  it('maps each category to a value/label option with an icon element', () => {
    const options = toCategoryOptions(categories)

    expect(options).toHaveLength(2)
    expect(options[0]).toMatchObject({ value: 'cat-1', label: 'Groceries' })
    expect(options[1]).toMatchObject({ value: 'cat-2', label: 'A Totally Unknown Category' })
    expect(isValidElement(options[0].icon)).toBe(true)
    expect(isValidElement(options[1].icon)).toBe(true)
  })

  it('returns an empty array for no categories', () => {
    expect(toCategoryOptions([])).toEqual([])
  })
})
