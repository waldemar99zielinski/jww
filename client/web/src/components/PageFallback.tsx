import { Loader } from '@/components/ui/Loader'

export function PageFallback() {
  return (
    <div className="pointer-events-none fixed inset-0 z-30 flex items-center justify-center">
      <Loader />
    </div>
  )
}
