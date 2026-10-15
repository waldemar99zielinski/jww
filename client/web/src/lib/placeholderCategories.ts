import type { ComponentType, SVGProps } from 'react'
import { IconBasket, IconBuilding, IconCar, IconFilm, IconHeart } from '@/components/icons'

/**
 * Fixed placeholder category set for the not-yet-real Budgets page (see
 * data/sampleBudgets.ts). Unlike lib/categoryVisuals.ts — which resolves an
 * icon/color for the server's free-text category names — these ids never
 * come from the API and should not be reached for anything outside
 * pages/budgets/.
 */
export type PlaceholderCategoryId = 'groceries' | 'transport' | 'rent' | 'entertainment' | 'health'

export interface PlaceholderCategory {
  id: PlaceholderCategoryId
  label: string
  colorVar: string
  Icon: ComponentType<SVGProps<SVGSVGElement>>
}

export const placeholderCategories: PlaceholderCategory[] = [
  { id: 'groceries', label: 'Groceries', colorVar: 'var(--cat-1)', Icon: IconBasket },
  { id: 'health', label: 'Health', colorVar: 'var(--cat-2)', Icon: IconHeart },
  { id: 'transport', label: 'Transport', colorVar: 'var(--cat-3)', Icon: IconCar },
  { id: 'entertainment', label: 'Fun', colorVar: 'var(--cat-4)', Icon: IconFilm },
  { id: 'rent', label: 'Rent', colorVar: 'var(--cat-5)', Icon: IconBuilding },
]

export function getPlaceholderCategory(id: PlaceholderCategoryId): PlaceholderCategory {
  const category = placeholderCategories.find((c) => c.id === id)
  if (!category) {
    throw new Error(`Unknown category: ${id}`)
  }
  return category
}
