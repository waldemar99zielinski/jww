import { useState, type FormEvent } from 'react'
import { Navigate, useLocation, useNavigate, type Location } from 'react-router-dom'
import { Logo } from '@/components/Logo'
import { Button } from '@/components/ui/Button'
import { TextInput } from '@/components/ui/TextInput'
import { useAuth } from '@/hooks/useAuth'

interface LocationState {
  from?: Location
}

export default function LoginPage() {
  const { status, error, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  if (status === 'authenticated') {
    const redirectTo = (location.state as LocationState | null)?.from?.pathname ?? '/'
    return <Navigate to={redirectTo} replace />
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    try {
      await login(username, password)
      navigate('/', { replace: true })
    } catch {
      // Failure is already surfaced via `error` below.
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-paper px-4 text-ink">
      <Logo size="lg" className="text-accent-strong" />

      <form
        onSubmit={handleSubmit}
        className="flex w-full max-w-sm flex-col gap-4 rounded-2xl border border-line bg-surface p-6 sm:p-8"
      >
        <div>
          <h1 className="m-0 text-xl font-bold">Welcome back</h1>
          <p className="m-0 mt-1 text-xs text-muted">Sign in to your account.</p>
        </div>

        <label className="flex flex-col gap-1.5 text-xs font-semibold text-ink-soft">
          Username
          <TextInput
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            autoFocus
            required
            className="rounded-[10px] px-3 sm:text-[13.5px]"
          />
        </label>

        <label className="flex flex-col gap-1.5 text-xs font-semibold text-ink-soft">
          Password
          <TextInput
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
            className="rounded-[10px] px-3 sm:text-[13.5px]"
          />
        </label>

        {error && <p className="m-0 text-[11.5px] text-critical">{error}</p>}

        <Button
          type="submit"
          disabled={status === 'loading'}
          variant="primary"
          className="py-3 text-[13.5px]"
        >
          {status === 'loading' ? 'Signing in…' : 'Sign in'}
        </Button>
      </form>
    </div>
  )
}
