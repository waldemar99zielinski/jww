import { Suspense, type ComponentType, type LazyExoticComponent, type ReactNode } from 'react'
import { PageFallback } from '@/components/PageFallback'
import { useViewport } from '@/hooks/useViewport'

interface ResponsiveViewProps<P extends object> {
  mobile: LazyExoticComponent<ComponentType<P>>
  desktop: LazyExoticComponent<ComponentType<P>>
  props: P
  fallback?: ReactNode
}

/**
 * Renders a page's mobile or desktop layout variant based on viewport width.
 * `mobile`/`desktop` must be `lazy()` components declared at module scope (see
 * src/pages/*\/index.tsx) — React only fetches a variant's chunk once it's
 * actually rendered, so a phone never downloads the desktop layout code and
 * vice versa, even though both lazy() calls happen at import time.
 */
export function ResponsiveView<P extends object>({ mobile, desktop, props, fallback }: ResponsiveViewProps<P>) {
  const viewport = useViewport()
  const Variant = viewport === 'desktop' ? desktop : mobile

  return (
    <Suspense fallback={fallback ?? <PageFallback />}>
      <Variant {...props} />
    </Suspense>
  )
}
