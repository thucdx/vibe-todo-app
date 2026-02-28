import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi } from 'vitest'
import TaskForm from './TaskForm'
import * as taskHooks from '../../hooks/useTasks'
import type { Task } from '../../types'

vi.mock('../../hooks/useTasks')

const editTask: Task = {
  id: 'task-edit',
  title: 'Existing task',
  date: '2026-02-28',
  due_time: '09:00',
  priority: 'high',
  tags: ['work'],
  points: 5,
  done: false,
  created_at: '2026-02-28T00:00:00Z',
  updated_at: '2026-02-28T00:00:00Z',
}

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupHookMocks(opts?: { isPending?: boolean; createError?: boolean; updateError?: boolean }) {
  const createMutateAsync = opts?.createError
    ? vi.fn().mockRejectedValue(new Error('network'))
    : vi.fn().mockResolvedValue({})

  const updateMutateAsync = opts?.updateError
    ? vi.fn().mockRejectedValue(new Error('network'))
    : vi.fn().mockResolvedValue({})

  vi.mocked(taskHooks.useCreateTask).mockReturnValue({
    mutateAsync: createMutateAsync,
    isPending: opts?.isPending ?? false,
  } as unknown as ReturnType<typeof taskHooks.useCreateTask>)

  vi.mocked(taskHooks.useUpdateTask).mockReturnValue({
    mutateAsync: updateMutateAsync,
    isPending: opts?.isPending ?? false,
  } as unknown as ReturnType<typeof taskHooks.useUpdateTask>)

  return { createMutateAsync, updateMutateAsync }
}

describe('TaskForm — create mode', () => {
  afterEach(() => vi.clearAllMocks())

  it('renders "New Task" heading', () => {
    setupHookMocks()
    const onClose = vi.fn()
    wrap(<TaskForm defaultDate="2026-02-28" onClose={onClose} />)
    expect(screen.getByText('New Task')).toBeInTheDocument()
  })

  it('renders "Create" submit button', () => {
    setupHookMocks()
    wrap(<TaskForm defaultDate="2026-02-28" onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument()
  })

  it('shows validation error when title is empty', async () => {
    setupHookMocks()
    wrap(<TaskForm defaultDate="2026-02-28" onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: /create/i }))
    expect(screen.getByText('Title is required.')).toBeInTheDocument()
  })

  it('calls createTask.mutateAsync on valid submit', async () => {
    const { createMutateAsync } = setupHookMocks()
    const onClose = vi.fn()
    wrap(<TaskForm defaultDate="2026-02-28" onClose={onClose} />)

    await userEvent.type(screen.getByPlaceholderText(/what needs to be done/i), 'My new task')
    await userEvent.click(screen.getByRole('button', { name: /create/i }))

    expect(createMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'My new task', date: '2026-02-28' }),
    )
    expect(onClose).toHaveBeenCalled()
  })

  it('shows error message on createTask failure', async () => {
    setupHookMocks({ createError: true })
    wrap(<TaskForm defaultDate="2026-02-28" onClose={vi.fn()} />)

    await userEvent.type(screen.getByPlaceholderText(/what needs to be done/i), 'Failing task')
    await userEvent.click(screen.getByRole('button', { name: /create/i }))

    expect(screen.getByText(/failed to save task/i)).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    setupHookMocks()
    const onClose = vi.fn()
    wrap(<TaskForm defaultDate="2026-02-28" onClose={onClose} />)
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalled()
  })

  it('calls onClose when Escape key is pressed', async () => {
    setupHookMocks()
    const onClose = vi.fn()
    wrap(<TaskForm defaultDate="2026-02-28" onClose={onClose} />)
    await userEvent.keyboard('{Escape}')
    expect(onClose).toHaveBeenCalled()
  })
})

describe('TaskForm — edit mode', () => {
  afterEach(() => vi.clearAllMocks())

  it('renders "Edit Task" heading', () => {
    setupHookMocks()
    wrap(<TaskForm task={editTask} defaultDate="2026-02-28" onClose={vi.fn()} />)
    expect(screen.getByText('Edit Task')).toBeInTheDocument()
  })

  it('pre-fills title field with existing task title', () => {
    setupHookMocks()
    wrap(<TaskForm task={editTask} defaultDate="2026-02-28" onClose={vi.fn()} />)
    const input = screen.getByPlaceholderText(/what needs to be done/i) as HTMLInputElement
    expect(input.value).toBe('Existing task')
  })

  it('calls updateTask.mutateAsync on submit', async () => {
    const { updateMutateAsync } = setupHookMocks()
    const onClose = vi.fn()
    wrap(<TaskForm task={editTask} defaultDate="2026-02-28" onClose={onClose} />)

    await userEvent.click(screen.getByRole('button', { name: /save/i }))

    expect(updateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'task-edit' }),
    )
    expect(onClose).toHaveBeenCalled()
  })
})
