import { useState } from 'react'
import { useAuthStatus, useLogin, useSetupPIN } from '../../hooks/useAuth'

interface Props {
  children: React.ReactNode
}

export default function PinGate({ children }: Props) {
  const { data: auth, isLoading } = useAuthStatus()
  const setupPIN = useSetupPIN()
  const login = useLogin()
  const [pin, setPin] = useState('')
  const [error, setError] = useState('')

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent" />
      </div>
    )
  }

  if (auth?.authenticated) {
    return <>{children}</>
  }

  const isSetup = !auth?.configured
  const isPending = setupPIN.isPending || login.isPending

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      if (isSetup) {
        await setupPIN.mutateAsync(pin)
      } else {
        await login.mutateAsync(pin)
      }
    } catch {
      setError(isSetup ? 'Failed to set PIN. Try again.' : 'Incorrect PIN.')
    }
  }

  return (
    <div className="flex h-screen items-center justify-center bg-gray-50">
      <form
        onSubmit={handleSubmit}
        className="w-80 space-y-5 rounded-2xl bg-white p-8 shadow-md"
      >
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">Todo Vibe</h1>
          <p className="mt-1 text-sm text-gray-500">
            {isSetup
              ? 'Create a PIN to protect your tasks'
              : 'Enter your PIN to continue'}
          </p>
        </div>

        <input
          type="password"
          value={pin}
          onChange={e => setPin(e.target.value)}
          placeholder="PIN (4–8 characters)"
          minLength={4}
          maxLength={8}
          autoFocus
          className="w-full rounded-lg border border-gray-300 px-4 py-2 text-center text-xl tracking-widest focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-200"
        />

        {error && (
          <p className="text-center text-sm text-red-500">{error}</p>
        )}

        <button
          type="submit"
          disabled={isPending || pin.length < 4}
          className="w-full rounded-lg bg-indigo-600 py-2 font-medium text-white transition hover:bg-indigo-700 disabled:opacity-50"
        >
          {isPending ? 'Please wait…' : isSetup ? 'Create PIN' : 'Unlock'}
        </button>
      </form>
    </div>
  )
}
