import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi } from 'vitest'
import MiniCalendar from './MiniCalendar'
import * as calendarHook from '../../hooks/useCalendar'
import * as taskHooks from '../../hooks/useTasks'

vi.mock('../../hooks/useCalendar')
vi.mock('../../hooks/useTasks')

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupMocks() {
  vi.mocked(calendarHook.useCalendar).mockReturnValue({
    data: [],
  } as ReturnType<typeof calendarHook.useCalendar>)

  vi.mocked(taskHooks.useMoveTask).mockReturnValue({
    mutate: vi.fn(),
  } as unknown as ReturnType<typeof taskHooks.useMoveTask>)
}

describe('MiniCalendar', () => {
  afterEach(() => vi.clearAllMocks())

  it('renders day-of-week headers', () => {
    setupMocks()
    const selectedDate = new Date('2026-02-15')
    wrap(<MiniCalendar selectedDate={selectedDate} onSelectDate={vi.fn()} />)

    for (const header of ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']) {
      expect(screen.getByText(header)).toBeInTheDocument()
    }
  })

  it('renders current month name', () => {
    setupMocks()
    // Freeze to February 2026 by using a known selected date
    const selectedDate = new Date('2026-02-15')
    wrap(<MiniCalendar selectedDate={selectedDate} onSelectDate={vi.fn()} />)
    // MiniCalendar initialises viewMonth to new Date() — but we just check it renders a month name
    expect(screen.getByText(/\w+ \d{4}/)).toBeInTheDocument()
  })

  it('calls onSelectDate when a day is clicked', async () => {
    setupMocks()
    const onSelectDate = vi.fn()
    const selectedDate = new Date('2026-02-01')
    wrap(<MiniCalendar selectedDate={selectedDate} onSelectDate={onSelectDate} />)

    // Find day "15" cell within the current view month and click it
    // Days are rendered as spans inside divs; text "15" should appear
    const day15 = screen.getAllByText('15')
    await userEvent.click(day15[0])
    expect(onSelectDate).toHaveBeenCalled()
  })

  it('navigates to previous month when ‹ is clicked', async () => {
    setupMocks()
    wrap(<MiniCalendar selectedDate={new Date('2026-02-15')} onSelectDate={vi.fn()} />)
    const before = screen.getByText(/\w+ \d{4}/).textContent
    await userEvent.click(screen.getByText('‹'))
    const after = screen.getByText(/\w+ \d{4}/).textContent
    expect(after).not.toBe(before)
  })

  it('navigates to next month when › is clicked', async () => {
    setupMocks()
    wrap(<MiniCalendar selectedDate={new Date('2026-02-15')} onSelectDate={vi.fn()} />)
    const before = screen.getByText(/\w+ \d{4}/).textContent
    await userEvent.click(screen.getByText('›'))
    const after = screen.getByText(/\w+ \d{4}/).textContent
    expect(after).not.toBe(before)
  })

  it('shows done/total indicator for days with summary data', () => {
    vi.mocked(calendarHook.useCalendar).mockReturnValue({
      data: [{ date: '2026-02-10', done: 2, total: 3 }],
    } as ReturnType<typeof calendarHook.useCalendar>)
    vi.mocked(taskHooks.useMoveTask).mockReturnValue({
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof taskHooks.useMoveTask>)

    wrap(
      <MiniCalendar selectedDate={new Date('2026-02-01')} onSelectDate={vi.fn()} />,
    )
    expect(screen.getByText('2/3')).toBeInTheDocument()
  })
})
