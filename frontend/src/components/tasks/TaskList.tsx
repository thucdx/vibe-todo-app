import { useState } from 'react'
import { useDeleteTask, useTasks, useToggleDone } from '../../hooks/useTasks'
import type { Task } from '../../types'
import TaskForm from './TaskForm'
import TaskItem from './TaskItem'

interface Props {
  date: string
}

export default function TaskList({ date }: Props) {
  const { data: tasks = [], isLoading } = useTasks(date)
  const toggleDone = useToggleDone(date)
  const deleteTask = useDeleteTask(date)
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const handleDelete = (id: string) => {
    if (window.confirm('Delete this task?')) {
      deleteTask.mutate(id)
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[1, 2, 3].map(i => (
          <div key={i} className="h-10 animate-pulse rounded-lg bg-gray-100" />
        ))}
      </div>
    )
  }

  return (
    <div>
      {tasks.length === 0 ? (
        <p className="py-6 text-center text-sm text-gray-400">
          No tasks for this day. Add one!
        </p>
      ) : (
        <div className="space-y-1">
          {tasks.map(task => (
            <TaskItem
              key={task.id}
              task={task}
              onToggle={id => toggleDone.mutate(id)}
              onEdit={setEditingTask}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}

      <button
        onClick={() => setShowCreate(true)}
        className="mt-3 flex w-full items-center gap-1 rounded-lg px-3 py-2 text-sm text-gray-400 hover:bg-gray-100 hover:text-gray-700 transition"
      >
        <span className="text-lg leading-none">+</span>
        <span>Add task</span>
      </button>

      {showCreate && (
        <TaskForm
          defaultDate={date}
          onClose={() => setShowCreate(false)}
        />
      )}

      {editingTask && (
        <TaskForm
          task={editingTask}
          defaultDate={date}
          onClose={() => setEditingTask(null)}
        />
      )}
    </div>
  )
}
