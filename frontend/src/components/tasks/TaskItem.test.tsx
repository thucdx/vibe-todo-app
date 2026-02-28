import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import TaskItem from './TaskItem'
import type { Task } from '../../types'

const baseTask: Task = {
  id: 'abc-123',
  title: 'Write unit tests',
  date: '2026-02-28',
  due_time: null,
  priority: 'medium',
  tags: [],
  points: 3,
  done: false,
  created_at: '2026-02-28T00:00:00Z',
  updated_at: '2026-02-28T00:00:00Z',
}

function renderTask(overrides?: Partial<Task>, handlers?: {
  onToggle?: (id: string) => void
  onEdit?: (t: Task) => void
  onDelete?: (id: string) => void
}) {
  const task = { ...baseTask, ...overrides }
  const onToggle = handlers?.onToggle ?? vi.fn()
  const onEdit   = handlers?.onEdit   ?? vi.fn()
  const onDelete = handlers?.onDelete ?? vi.fn()
  render(<TaskItem task={task} onToggle={onToggle} onEdit={onEdit} onDelete={onDelete} />)
  return { task, onToggle, onEdit, onDelete }
}

describe('TaskItem', () => {
  it('renders task title', () => {
    renderTask()
    expect(screen.getByText('Write unit tests')).toBeInTheDocument()
  })

  it('renders priority badge', () => {
    renderTask({ priority: 'high' })
    expect(screen.getByText('H')).toBeInTheDocument()
  })

  it('renders points', () => {
    renderTask({ points: 5 })
    expect(screen.getByText('5pt')).toBeInTheDocument()
  })

  it('renders due_time when present', () => {
    renderTask({ due_time: '14:30' })
    expect(screen.getByText('14:30')).toBeInTheDocument()
  })

  it('does not render due_time when null', () => {
    renderTask({ due_time: null })
    expect(screen.queryByText(/\d{2}:\d{2}/)).not.toBeInTheDocument()
  })

  it('renders tags', () => {
    renderTask({ tags: ['work', 'urgent'] })
    expect(screen.getByText('#work')).toBeInTheDocument()
    expect(screen.getByText('#urgent')).toBeInTheDocument()
  })

  it('checkbox is checked when task is done', () => {
    renderTask({ done: true })
    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).toBeChecked()
  })

  it('checkbox is unchecked when task is not done', () => {
    renderTask({ done: false })
    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).not.toBeChecked()
  })

  it('calls onToggle when checkbox is clicked', async () => {
    const onToggle = vi.fn()
    renderTask({}, { onToggle })
    await userEvent.click(screen.getByRole('checkbox'))
    expect(onToggle).toHaveBeenCalledWith('abc-123')
  })

  it('calls onEdit when edit button is clicked', async () => {
    const onEdit = vi.fn()
    renderTask({}, { onEdit })
    await userEvent.click(screen.getByTitle('Edit'))
    expect(onEdit).toHaveBeenCalledWith(expect.objectContaining({ id: 'abc-123' }))
  })

  it('calls onDelete when delete button is clicked', async () => {
    const onDelete = vi.fn()
    renderTask({}, { onDelete })
    await userEvent.click(screen.getByTitle('Delete'))
    expect(onDelete).toHaveBeenCalledWith('abc-123')
  })
})
