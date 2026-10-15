import { lazy, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TransactionCategoryDto, TransactionType } from '@/api/types'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useCategories, useCreateCategory, useUpdateCategory } from '@/hooks/useCategories'
import type { SettingsCategoriesViewProps } from './types'

const Mobile = lazy(() => import('./SettingsCategories.mobile'))
const Desktop = lazy(() => import('./SettingsCategories.desktop'))

export default function SettingsCategoriesPage() {
  const navigate = useNavigate()
  const [type, setType] = useState<TransactionType>('expense')
  const [name, setName] = useState('')

  const { data: categories, isLoading } = useCategories(type, false)
  const createCategory = useCreateCategory()
  const updateCategory = useUpdateCategory()

  function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) return
    createCategory.mutate({ type, name: trimmed }, { onSuccess: () => setName('') })
  }

  function handleToggleActive(category: TransactionCategoryDto) {
    updateCategory.mutate({
      id: category.id,
      name: category.name,
      active: !category.active,
      sort_order: category.sort_order,
    })
  }

  const props: SettingsCategoriesViewProps = {
    type,
    onSelectType: setType,
    categories,
    isLoading,
    onToggleActive: handleToggleActive,
    name,
    onNameChange: setName,
    onCreate: handleCreate,
    creating: createCategory.isPending,
    createError: createCategory.isError,
    onClose: () => navigate(-1),
  }

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
