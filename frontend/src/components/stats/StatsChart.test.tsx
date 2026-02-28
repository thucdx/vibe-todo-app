import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi } from 'vitest'
import StatsChart from './StatsChart'
import * as statsHook from '../../hooks/useStats'

vi.mock('../../hooks/useStats')

// recharts ResponsiveContainer requires a measured DOM container.
// In jsdom there are no real dimensions, so we stub it to render children directly.
vi.mock('recharts', async () => {
  const actual = await vi.importActual<typeof import('recharts')>('recharts')
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="recharts-container">{children}</div>
    ),
  }
})

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('StatsChart', () => {
  afterEach(() => vi.clearAllMocks())

  it('shows loading skeleton while fetching', () => {
    vi.mocked(statsHook.useStats).mockReturnValue({
      data: [],
      isLoading: true,
    } as ReturnType<typeof statsHook.useStats>)

    wrap(<StatsChart />)
    const skeleton = document.querySelector('.animate-pulse')
    expect(skeleton).toBeInTheDocument()
  })

  it('renders view toggle buttons', () => {
    vi.mocked(statsHook.useStats).mockReturnValue({
      data: [],
      isLoading: false,
    } as ReturnType<typeof statsHook.useStats>)

    wrap(<StatsChart />)
    expect(screen.getByRole('button', { name: /day/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /week/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /month/i })).toBeInTheDocument()
  })

  it('renders metric toggle buttons', () => {
    vi.mocked(statsHook.useStats).mockReturnValue({
      data: [],
      isLoading: false,
    } as ReturnType<typeof statsHook.useStats>)

    wrap(<StatsChart />)
    expect(screen.getByRole('button', { name: /count/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /points/i })).toBeInTheDocument()
  })

  it('renders the chart container when not loading', () => {
    vi.mocked(statsHook.useStats).mockReturnValue({
      data: [{ label: 'Mon', value: 5 }],
      isLoading: false,
    } as ReturnType<typeof statsHook.useStats>)

    wrap(<StatsChart />)
    expect(screen.getByTestId('recharts-container')).toBeInTheDocument()
  })

  it('calls useStats with new view when toggle is clicked', async () => {
    const mockUseStats = vi.mocked(statsHook.useStats).mockReturnValue({
      data: [],
      isLoading: false,
    } as ReturnType<typeof statsHook.useStats>)

    wrap(<StatsChart />)
    await userEvent.click(screen.getByRole('button', { name: /week/i }))

    // After clicking week, useStats should have been called with view='week'
    const calls = mockUseStats.mock.calls
    const lastCall = calls[calls.length - 1]
    expect(lastCall[0]).toBe('week')
  })

  it('calls useStats with new metric when points toggle is clicked', async () => {
    const mockUseStats = vi.mocked(statsHook.useStats).mockReturnValue({
      data: [],
      isLoading: false,
    } as ReturnType<typeof statsHook.useStats>)

    wrap(<StatsChart />)
    await userEvent.click(screen.getByRole('button', { name: /points/i }))

    const calls = mockUseStats.mock.calls
    const lastCall = calls[calls.length - 1]
    expect(lastCall[1]).toBe('points')
  })
})
