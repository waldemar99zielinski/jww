import { Logo } from '@/components/Logo'

/**
 * Shown while useAppBoot (src/App.tsx) is still waiting on the session check, first-screen
 * chunks, and its own MIN_BOOT_SPLASH_MS floor. The "Horizontal Unfold" <Logo animated /> plays
 * the JWW -> Joint Wallet Watch unfold once and holds there.
 */
export function BootSplash() {
  return (
    <div role="status" aria-label="Loading" className="flex min-h-screen items-center justify-center bg-paper px-6">
      <Logo animated size="lg" className="text-accent-strong" />
    </div>
  )
}
