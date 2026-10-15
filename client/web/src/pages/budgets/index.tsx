import { lazy } from 'react'
import { ResponsiveView } from '@/components/ResponsiveView'

const Mobile = lazy(() => import('./Budgets.mobile'))
const Desktop = lazy(() => import('./Budgets.desktop'))

export default function BudgetsPage() {
  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={{}} />
}
