import { useEffect, useState } from 'react'
import { useCreateTask, useUpdateTask } from '../../hooks/useTasks'
import type { Priority, Task } from '../../types'

const PRIORITIES: Priority[] = ['high', 'medium', 'low']

interface Props {
  task?: Task
  defaultDate: string
  onClose: () => void
}

export default function TaskForm({ task, defaultDate, onClose }: Props) {
  const isEdit = !!task
  const createTask = useCreateTask(defaultDate)
  const updateTask = useUpdateTask(defaultDate)

  const [title, setTitle]       = useState(task?.title ?? '')
  const [date, setDate]         = useState(task?.date ? task.date.slice(0, 10) : defaultDate)
  const [dueTime, setDueTime]   = useState(task?.due_time ?? '')
  const [priority, setPriority] = useState<Priority>(task?.priority ?? 'medium')
  const [tagInput, setTagInput] = useState(task?.tags.join(', ') ?? '')
  const [points, setPoints]     = useState(task?.points ?? 1)
  const [error, setError]       = useState('')

  // Close on Escape key
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [onClose])

  const parsedTags = tagInput
    .split(',')
    .map(t => t.trim())
    .filter(Boolean)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!title.trim()) { setError('Title is required.'); return }

    const payload = {
      title: title.trim(),
      date,
      due_time: dueTime || null,
      priority,
      tags: parsedTags,
      points,
    }

    try {
      if (isEdit) {
        await updateTask.mutateAsync({ id: task.id, data: payload })
      } else {
        await createTask.mutateAsync(payload)
      }
      onClose()
    } catch {
      setError('Failed to save task. Please try again.')
    }
  }

  const isPending = createTask.isPending || updateTask.isPending

  return (
    /* Backdrop */
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      onClick={onClose}
    >
      {/* Card — stop propagation so clicks inside don't close */}
      <form
        onSubmit={handleSubmit}
        onClick={e => e.stopPropagation()}
        className="w-full max-w-md space-y-4 rounded-2xl bg-white p-6 shadow-xl"
      >
        <h2 className="text-lg font-semibold text-gray-900">
          {isEdit ? 'Edit Task' : 'New Task'}
        </h2>

        {/* Title */}
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Title *</label>
          <input
            autoFocus
            value={title}
            onChange={e => setTitle(e.target.value)}
            placeholder="What needs to be done?"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-200"
          />
        </div>

        {/* Date + Due time */}
        <div className="flex gap-3">
          <div className="flex-1">
            <label className="mb-1 block text-xs font-medium text-gray-500">Date *</label>
            <input
              type="date"
              value={date}
              onChange={e => setDate(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
          </div>
          <div className="flex-1">
            <label className="mb-1 block text-xs font-medium text-gray-500">Due time</label>
            <input
              type="time"
              value={dueTime}
              onChange={e => setDueTime(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
          </div>
        </div>

        {/* Priority */}
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Priority</label>
          <div className="flex gap-2">
            {PRIORITIES.map(p => (
              <button
                key={p}
                type="button"
                onClick={() => setPriority(p)}
                className={[
                  'flex-1 rounded-lg border py-1 text-xs font-medium capitalize transition',
                  priority === p
                    ? p === 'high'
                      ? 'border-red-400 bg-red-50 text-red-700'
                      : p === 'medium'
                        ? 'border-yellow-400 bg-yellow-50 text-yellow-700'
                        : 'border-gray-400 bg-gray-50 text-gray-700'
                    : 'border-gray-200 text-gray-400 hover:border-gray-300',
                ].join(' ')}
              >
                {p}
              </button>
            ))}
          </div>
        </div>

        {/* Tags */}
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">
            Tags <span className="font-normal text-gray-400">(comma-separated)</span>
          </label>
          <input
            value={tagInput}
            onChange={e => setTagInput(e.target.value)}
            placeholder="work, personal, health"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-200"
          />
          {parsedTags.length > 0 && (
            <div className="mt-1.5 flex flex-wrap gap-1">
              {parsedTags.map(t => (
                <span
                  key={t}
                  className="rounded bg-indigo-50 px-1.5 py-0.5 text-[11px] text-indigo-600"
                >
                  #{t}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Points */}
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Points</label>
          <input
            type="number"
            min={1}
            max={100}
            value={points}
            onChange={e => setPoints(Math.max(1, Math.min(100, Number(e.target.value))))}
            className="w-24 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-200"
          />
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        {/* Actions */}
        <div className="flex gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 rounded-lg border border-gray-300 py-2 text-sm text-gray-600 hover:bg-gray-50 transition"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="flex-1 rounded-lg bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50 transition"
          >
            {isPending ? 'Saving…' : isEdit ? 'Save' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  )
}
