import { useCurrencies } from '@/hooks/useCurrencies'

/** The base currency drives reporting and is the sane default for new transactions. */
export function useDefaultCurrency() {
  const currenciesQuery = useCurrencies()
  const currencies = currenciesQuery.data ?? []
  const currency = currencies.find((c) => c.is_base) ?? currencies[0] ?? null

  return { currency, currencies, isLoading: currenciesQuery.isLoading }
}
