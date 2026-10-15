import { lazy, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { CurrencyDto } from '@/api/types'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useCreateCurrency, useCurrencies, useUpdateCurrency } from '@/hooks/useCurrencies'
import type { SettingsCurrenciesViewProps } from './types'

const Mobile = lazy(() => import('./SettingsCurrencies.mobile'))
const Desktop = lazy(() => import('./SettingsCurrencies.desktop'))

export default function SettingsCurrenciesPage() {
  const navigate = useNavigate()
  const { data: currencies, isLoading } = useCurrencies()
  const createCurrency = useCreateCurrency()
  const updateCurrency = useUpdateCurrency()

  const [code, setCode] = useState('')
  const [name, setName] = useState('')

  function handleCreate() {
    const trimmedCode = code.trim().toUpperCase()
    const trimmedName = name.trim()
    if (trimmedCode.length !== 3 || !trimmedName) return
    createCurrency.mutate(
      { code: trimmedCode, name: trimmedName, isBase: (currencies?.length ?? 0) === 0 },
      {
        onSuccess: () => {
          setCode('')
          setName('')
        },
      },
    )
  }

  function handleMakeBase(currency: CurrencyDto) {
    updateCurrency.mutate({ code: currency.code, name: currency.name, isBase: true })
  }

  const props: SettingsCurrenciesViewProps = {
    currencies,
    isLoading,
    onMakeBase: handleMakeBase,
    code,
    onCodeChange: setCode,
    name,
    onNameChange: setName,
    onCreate: handleCreate,
    creating: createCurrency.isPending,
    hasError: createCurrency.isError || updateCurrency.isError,
    onClose: () => navigate(-1),
  }

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
