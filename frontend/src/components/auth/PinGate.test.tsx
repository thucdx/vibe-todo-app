import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi } from 'vitest'
import PinGate from './PinGate'
import * as authHooks from '../../hooks/useAuth'

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('PinGate', () => {
  it('shows loading spinner while checking auth', () => {
    vi.spyOn(authHooks, 'useAuthStatus').mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof authHooks.useAuthStatus>)

    wrap(<PinGate><div>App</div></PinGate>)
    expect(screen.queryByText('App')).not.toBeInTheDocument()
  })

  it('shows create-PIN form when PIN not yet configured', () => {
    vi.spyOn(authHooks, 'useAuthStatus').mockReturnValue({
      data: { configured: false, authenticated: false },
      isLoading: false,
    } as ReturnType<typeof authHooks.useAuthStatus>)

    wrap(<PinGate><div>App</div></PinGate>)
    expect(screen.getByText('Create PIN')).toBeInTheDocument()
  })

  it('shows unlock form when PIN is configured', () => {
    vi.spyOn(authHooks, 'useAuthStatus').mockReturnValue({
      data: { configured: true, authenticated: false },
      isLoading: false,
    } as ReturnType<typeof authHooks.useAuthStatus>)

    wrap(<PinGate><div>App</div></PinGate>)
    expect(screen.getByText('Unlock')).toBeInTheDocument()
  })

  it('renders children when authenticated', () => {
    vi.spyOn(authHooks, 'useAuthStatus').mockReturnValue({
      data: { configured: true, authenticated: true },
      isLoading: false,
    } as ReturnType<typeof authHooks.useAuthStatus>)

    wrap(<PinGate><div>App Content</div></PinGate>)
    expect(screen.getByText('App Content')).toBeInTheDocument()
  })

  it('disables the submit button when PIN is too short', async () => {
    vi.spyOn(authHooks, 'useAuthStatus').mockReturnValue({
      data: { configured: true, authenticated: false },
      isLoading: false,
    } as ReturnType<typeof authHooks.useAuthStatus>)

    wrap(<PinGate><div>App</div></PinGate>)
    const btn = screen.getByRole('button', { name: /unlock/i })
    expect(btn).toBeDisabled()

    await userEvent.type(screen.getByPlaceholderText(/PIN/i), '12')
    expect(btn).toBeDisabled()

    await userEvent.type(screen.getByPlaceholderText(/PIN/i), '34')
    expect(btn).not.toBeDisabled()
  })
})
