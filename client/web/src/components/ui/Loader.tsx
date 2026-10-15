import { IconWallet } from '@/components/icons'
import { cx } from './classes'

export interface LoaderProps {
  className?: string
}

/** Default app loader: a coin dropping into a gently bobbing wallet. See index.css for the keyframes. */
export function Loader({ className }: LoaderProps) {
  return (
    <span
      role="status"
      aria-label="Loading"
      className={cx('wallet-loader relative inline-flex h-16 w-16 items-center justify-center', className)}
    >
      <span className="wallet-coin absolute left-1/2 top-0 h-2.5 w-2.5 rounded-full bg-accent" />
      <IconWallet className="wallet-body h-12 w-12 text-accent-strong" strokeWidth={1.5} />
    </span>
  )
}
