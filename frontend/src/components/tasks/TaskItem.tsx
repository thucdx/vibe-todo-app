import type { Priority, Task } from '../../types'

const PRIORITY_BADGE: Record<Priority, string> = {
  high:   'bg-red-100 text-red-700',
  medium: 'bg-yellow-100 text-yellow-700',
  low:    'bg-gray-100 text-gray-500',
}

interface Props {
  task: Task
  onToggle: (id: string) => void
  onEdit: (task: Task) => void
  onDelete: (id: string) => void
}

export default function TaskItem({ task, onToggle, onEdit, onDelete }: Props) {
  const handleDragStart = (e: React.DragEvent<HTMLDivElement>) => {
    e.dataTransfer.setData('application/task', JSON.stringify(task))
    e.dataTransfer.effectAllowed = 'move'
  }

  return (
    <div
      draggable
      onDragStart={handleDragStart}
      className={[
        'group flex cursor-grab items-center gap-2 rounded-lg px-3 py-2 transition hover:bg-gray-50',
        task.done ? 'opacity-60' : '',
      ].join(' ')}
    >
      {/* Checkbox */}
      <input
        type="checkbox"
        checked={task.done}
        onChange={() => onToggle(task.id)}
        onClick={e => e.stopPropagation()}
        className="h-4 w-4 shrink-0 cursor-pointer rounded accent-indigo-600"
      />

      {/* Title */}
      <span
        className={[
          'flex-1 truncate text-sm',
          task.done ? 'text-gray-400 line-through' : 'text-gray-800',
        ].join(' ')}
      >
        {task.title}
      </span>

      {/* Due time */}
      {task.due_time && (
        <span className="shrink-0 text-xs text-gray-400">
          {task.due_time.slice(0, 5)}
        </span>
      )}

      {/* Priority badge */}
      <span
        className={`shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold ${PRIORITY_BADGE[task.priority]}`}
      >
        {task.priority[0].toUpperCase()}
      </span>

      {/* Tags */}
      {task.tags.map(tag => (
        <span
          key={tag}
          className="shrink-0 rounded bg-indigo-50 px-1.5 py-0.5 text-[10px] text-indigo-600"
        >
          #{tag}
        </span>
      ))}

      {/* Points */}
      <span className="shrink-0 text-xs text-gray-400">{task.points}pt</span>

      {/* Hover actions */}
      <div className="hidden shrink-0 gap-1 group-hover:flex">
        <button
          onClick={() => onEdit(task)}
          title="Edit"
          className="rounded p-0.5 text-gray-400 hover:text-indigo-600 transition"
        >
          ✎
        </button>
        <button
          onClick={() => onDelete(task.id)}
          title="Delete"
          className="rounded p-0.5 text-gray-400 hover:text-red-500 transition"
        >
          🗑
        </button>
      </div>
    </div>
  )
}
