import type { TransactionCategoryDto } from '@/api/types'
import type { SearchableSelectOption } from '@/components/ui/SearchableSelect'
import { getCategoryVisual } from '@/lib/categoryVisuals'
import { cssVars } from '@/lib/cssVars'

export function toCategoryOptions(categories: TransactionCategoryDto[]): SearchableSelectOption[] {
  return categories.map((category) => {
    const { Icon, colorVar } = getCategoryVisual(category.name)
    return {
      value: category.id,
      label: category.name,
      icon: (
        <span
          className="category-tint flex h-6 w-6 flex-none items-center justify-center rounded-md"
          style={cssVars({ '--tint': colorVar })}
        >
          <Icon className="h-3.5 w-3.5" />
        </span>
      ),
    }
  })
}
