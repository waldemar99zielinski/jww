import type { ComponentType, SVGProps } from 'react'
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

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>

/**
 * Real categories come from the server as free-text names (no icon/color of
 * their own), unlike the fixed placeholder set in lib/placeholderCategories.ts
 * (budgets-page-only). The server bootstraps a fixed set of category names
 * (see server/migrations/00006_seed_transaction_categories.sql) — map those
 * exactly by name, then fall back to keyword matching and finally a
 * deterministic color + generic tag icon for anything user-created.
 */
const ICONS_BY_NAME: Record<string, IconComponent> = {
  // expense
  groceries: IconBasket,
  transport: IconCar,
  entertainment: IconFilm,
  health: IconHeart,
  rent: IconBuilding,
  clothes: IconShirt,
  subscriptions: IconRepeat,
  travel: IconPlane,
  utilities: IconBolt,
  education: IconGraduationCap,
  'personal care': IconSparkle,
  gifts: IconGift,
  'dining out': IconUtensils,
  baby: IconBabyFace,
  home: IconHome,
  // income
  salary: IconBanknote,
  freelance: IconBriefcase,
  investments: IconTrendingUp,
  refunds: IconUndo,
  bonus: IconStar,
}

const ICONS_BY_KEYWORD: { keywords: string[]; Icon: IconComponent }[] = [
  { keywords: ['grocery', 'groceries', 'food', 'supermarket'], Icon: IconBasket },
  { keywords: ['health', 'medical', 'pharmacy', 'doctor'], Icon: IconHeart },
  { keywords: ['transport', 'car', 'fuel', 'commute', 'uber', 'taxi'], Icon: IconCar },
  { keywords: ['entertainment', 'fun', 'movie', 'cinema', 'subscription'], Icon: IconFilm },
  { keywords: ['rent', 'housing', 'mortgage', 'home', 'bills'], Icon: IconBuilding },
]

const FALLBACK_COLORS = ['var(--cat-1)', 'var(--cat-2)', 'var(--cat-3)', 'var(--cat-4)', 'var(--cat-5)']

function hashString(value: string): number {
  let hash = 0
  for (let i = 0; i < value.length; i++) {
    hash = (hash * 31 + value.charCodeAt(i)) | 0
  }
  return Math.abs(hash)
}

export interface CategoryVisual {
  Icon: IconComponent
  colorVar: string
}

export function getCategoryVisual(name: string): CategoryVisual {
  const normalized = name.toLowerCase()
  const exact = ICONS_BY_NAME[normalized]
  const keywordMatch = ICONS_BY_KEYWORD.find(({ keywords }) => keywords.some((keyword) => normalized.includes(keyword)))
  const colorVar = FALLBACK_COLORS[hashString(normalized) % FALLBACK_COLORS.length]
  return { Icon: exact ?? keywordMatch?.Icon ?? IconTag, colorVar }
}
