import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi } from 'vitest'
import TaskList from './TaskList'
import * as taskHooks from '../../hooks/useTasks'
import type { Task } from '../../types'

// Mock the hooks module so no real HTTP calls are made.
vi.mock('../../hooks/useTasks')

const mockTask: Task = {
  id: 'task-1',
  title: 'Buy groceries',
  date: '2026-02-28',
  due_time: null,
  priority: 'low',
  tags: [],
  points: 1,
  done: false,
  created_at: '2026-02-28T00:00:00Z',
  updated_at: '2026-02-28T00:00:00Z',
}

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupMocks(overrides?: {
  tasks?: Task[]
  isLoading?: boolean
}) {
  const mutate = vi.fn()
  vi.mocked(taskHooks.useTasks).mockReturnValue({
    data: overrides?.tasks ?? [],
    isLoading: overrides?.isLoading ?? false,
  } as ReturnType<typeof taskHooks.useTasks>)

  vi.mocked(taskHooks.useToggleDone).mockReturnValue({ mutate } as ReturnType<typeof taskHooks.useToggleDone>)
  vi.mocked(taskHooks.useDeleteTask).mockReturnValue({ mutate } as ReturnType<typeof taskHooks.useDeleteTask>)

  return { mutate }
}

describe('TaskList', () => {
  afterEach(() => vi.clearAllMocks())

  it('shows loading skeleton while fetching', () => {
    setupMocks({ isLoading: true })
    wrap(<TaskList date="2026-02-28" />)
    // 3 skeleton divs should be present (animate-pulse)
    const skeletons = document.querySelectorAll('.animate-pulse')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('shows empty state message when there are no tasks', () => {
    setupMocks({ tasks: [] })
    wrap(<TaskList date="2026-02-28" />)
    expect(screen.getByText(/No tasks for this day/i)).toBeInTheDocument()
  })

  it('renders task titles when tasks are present', () => {
    setupMocks({ tasks: [mockTask] })
    wrap(<TaskList date="2026-02-28" />)
    expect(screen.getByText('Buy groceries')).toBeInTheDocument()
  })

  it('shows "Add task" button', () => {
    setupMocks()
    wrap(<TaskList date="2026-02-28" />)
    expect(screen.getByText('Add task')).toBeInTheDocument()
  })

  it('opens TaskForm when "Add task" is clicked', async () => {
    // TaskForm uses useCreateTask/useUpdateTask — mock them too
    vi.mocked(taskHooks.useCreateTask).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof taskHooks.useCreateTask>)
    vi.mocked(taskHooks.useUpdateTask).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof taskHooks.useUpdateTask>)

    setupMocks()
    wrap(<TaskList date="2026-02-28" />)
    await userEvent.click(screen.getByText('Add task'))
    expect(screen.getByText('New Task')).toBeInTheDocument()
  })

  it('calls deleteTask.mutate after window.confirm', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const { mutate } = setupMocks({ tasks: [mockTask] })
    wrap(<TaskList date="2026-02-28" />)
    await userEvent.click(screen.getByTitle('Delete'))
    expect(mutate).toHaveBeenCalledWith('task-1')
  })

  it('does not call deleteTask.mutate when confirm is cancelled', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const { mutate } = setupMocks({ tasks: [mockTask] })
    wrap(<TaskList date="2026-02-28" />)
    await userEvent.click(screen.getByTitle('Delete'))
    expect(mutate).not.toHaveBeenCalled()
  })
})
