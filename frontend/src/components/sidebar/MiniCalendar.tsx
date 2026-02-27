import { useState } from 'react'
import {
  addMonths,
  eachDayOfInterval,
  endOfMonth,
  format,
  getDay,
  isToday,
  isSameDay,
  startOfMonth,
  subMonths,
} from 'date-fns'
import { useCalendar } from '../../hooks/useCalendar'
import { useMoveTask } from '../../hooks/useTasks'
import type { Task } from '../../types'

const DAY_HEADERS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

interface Props {
  selectedDate: Date
  onSelectDate: (d: Date) => void
}

export default function MiniCalendar({ selectedDate, onSelectDate }: Props) {
  const [viewMonth, setViewMonth] = useState<Date>(new Date())
  const { data: summary = [] } = useCalendar(
    viewMonth.getFullYear(),
    viewMonth.getMonth() + 1,
  )
  const moveTask = useMoveTask()

  const days = eachDayOfInterval({
    start: startOfMonth(viewMonth),
    end: endOfMonth(viewMonth),
  })
  const leadingPads = getDay(startOfMonth(viewMonth))

  const summaryMap = new Map(summary.map(s => [s.date, s]))

  const handleDrop = (e: React.DragEvent<HTMLDivElement>, day: Date) => {
    e.preventDefault()
    const raw = e.dataTransfer.getData('application/task')
    if (!raw) return
    const task: Task = JSON.parse(raw)
    const newDate = format(day, 'yyyy-MM-dd')
    if (task.date === newDate) return
    moveTask.mutate({ id: task.id, task, newDate })
  }

  return (
    <div className="p-3 select-none">
      {/* Month navigation */}
      <div className="mb-3 flex items-center justify-between">
        <button
          onClick={() => setViewMonth(m => subMonths(m, 1))}
          className="rounded p-1 text-gray-400 hover:text-indigo-600 transition"
        >
          ‹
        </button>
        <span className="text-sm font-semibold text-gray-700">
          {format(viewMonth, 'MMMM yyyy')}
        </span>
        <button
          onClick={() => setViewMonth(m => addMonths(m, 1))}
          className="rounded p-1 text-gray-400 hover:text-indigo-600 transition"
        >
          ›
        </button>
      </div>

      {/* Day-of-week headers */}
      <div className="mb-1 grid grid-cols-7 text-center text-[10px] font-medium text-gray-400">
        {DAY_HEADERS.map(d => (
          <span key={d}>{d}</span>
        ))}
      </div>

      {/* Calendar grid */}
      <div className="grid grid-cols-7 gap-y-0.5">
        {/* Leading padding */}
        {Array.from({ length: leadingPads }).map((_, i) => (
          <div key={`pad-${i}`} />
        ))}

        {days.map(day => {
          const key = format(day, 'yyyy-MM-dd')
          const s = summaryMap.get(key)
          const isSelected = isSameDay(day, selectedDate)
          const isAllDone = !!s && s.total > 0 && s.done === s.total

          return (
            <div
              key={key}
              onClick={() => onSelectDate(day)}
              onDragOver={e => e.preventDefault()}
              onDrop={e => handleDrop(e, day)}
              className={[
                'flex cursor-pointer flex-col items-center rounded-lg py-1 text-xs transition hover:bg-indigo-50',
                isSelected ? 'bg-indigo-600 text-white hover:bg-indigo-700' : '',
                isToday(day) && !isSelected
                  ? 'ring-1 ring-indigo-400 ring-inset'
                  : '',
              ]
                .filter(Boolean)
                .join(' ')}
            >
              <span className="font-medium leading-none">{format(day, 'd')}</span>
              {s && (
                <span
                  className={[
                    'mt-0.5 text-[9px] leading-none',
                    isSelected
                      ? 'text-indigo-200'
                      : isAllDone
                        ? 'text-green-500'
                        : 'text-gray-400',
                  ].join(' ')}
                >
                  {s.done}/{s.total}
                </span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
