import { describe, expect, it } from 'vitest'
import {
  IconBabyFace,
  IconBanknote,
  IconBasket,
  IconBolt,
  IconBriefcase,
  IconBuilding,
  IconCar,
  IconFilm,
  IconGift,
  IconGraduationCap,
  IconHeart,
  IconHome,
  IconPlane,
  IconRepeat,
  IconShirt,
  IconSparkle,
  IconStar,
  IconTag,
  IconTrendingUp,
  IconUndo,
  IconUtensils,
} from '@/components/icons'
import { getCategoryVisual } from '@/lib/categoryVisuals'

const BOOTSTRAPPED_EXPENSE_CATEGORIES: [string, unknown][] = [
  ['Groceries', IconBasket],
  ['Transport', IconCar],
  ['Entertainment', IconFilm],
  ['Health', IconHeart],
  ['Rent', IconBuilding],
  ['Clothes', IconShirt],
  ['Subscriptions', IconRepeat],
  ['Travel', IconPlane],
  ['Utilities', IconBolt],
  ['Education', IconGraduationCap],
  ['Personal care', IconSparkle],
  ['Gifts', IconGift],
  ['Dining out', IconUtensils],
  ['Baby', IconBabyFace],
  ['Home', IconHome],
]

const BOOTSTRAPPED_INCOME_CATEGORIES: [string, unknown][] = [
  ['Salary', IconBanknote],
  ['Freelance', IconBriefcase],
  ['Investments', IconTrendingUp],
  ['Refunds', IconUndo],
  ['Bonus', IconStar],
  ['Gifts', IconGift],
]

describe('getCategoryVisual — exact-name matches', () => {
  it.each([...BOOTSTRAPPED_EXPENSE_CATEGORIES, ...BOOTSTRAPPED_INCOME_CATEGORIES])(
    'resolves "%s" to its dedicated icon',
    (name, Icon) => {
      expect(getCategoryVisual(name as string).Icon).toBe(Icon)
    },
  )

  it('is case-insensitive', () => {
    expect(getCategoryVisual('groceries').Icon).toBe(IconBasket)
    expect(getCategoryVisual('GROCERIES').Icon).toBe(IconBasket)
    expect(getCategoryVisual('GrOcErIeS').Icon).toBe(IconBasket)
  })
})

describe('getCategoryVisual — keyword fallback', () => {
  it('matches a user-created category name containing a known keyword', () => {
    expect(getCategoryVisual('Weekly Supermarket Run').Icon).toBe(IconBasket)
    expect(getCategoryVisual('Doctor visit').Icon).toBe(IconHeart)
    expect(getCategoryVisual('Uber rides').Icon).toBe(IconCar)
  })
})

describe('getCategoryVisual — final fallback', () => {
  it('falls back to the generic tag icon for a totally unrecognized name', () => {
    expect(getCategoryVisual('Zzyxw Placeholder Category').Icon).toBe(IconTag)
  })

  it('is deterministic — the same unrecognized name always gets the same color', () => {
    const first = getCategoryVisual('Zzyxw Placeholder Category')
    const second = getCategoryVisual('Zzyxw Placeholder Category')
    expect(first.colorVar).toBe(second.colorVar)
  })

  it('can assign different colors to different unrecognized names', () => {
    const colors = new Set(
      ['Qumbari Ledger', 'Vexolith Tracker', 'Nembrax Account', 'Ploxium Entry', 'Driggsworth Item'].map(
        (name) => getCategoryVisual(name).colorVar,
      ),
    )
    expect(colors.size).toBeGreaterThan(1)
  })
})
