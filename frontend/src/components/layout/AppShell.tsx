import { useState } from 'react'
import { format } from 'date-fns'
import MiniCalendar from '../sidebar/MiniCalendar'
import TaskList from '../tasks/TaskList'
import StatsChart from '../stats/StatsChart'
import { useLogout } from '../../hooks/useAuth'

export default function AppShell() {
  const [selectedDate, setSelectedDate] = useState<Date>(new Date())
  const logout = useLogout()

  const dateStr = format(selectedDate, 'yyyy-MM-dd')

  return (
    <div className="flex h-screen flex-col">
      {/* Header */}
      <header className="flex items-center justify-between border-b bg-white px-6 py-3 shadow-sm">
        <h1 className="text-lg font-bold text-indigo-600">Todo Vibe</h1>
        <button
          onClick={() => logout.mutate()}
          title="Lock"
          className="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 transition"
        >
          🔒
        </button>
      </header>

      {/* Body */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside className="w-56 shrink-0 overflow-y-auto border-r bg-white">
          <MiniCalendar
            selectedDate={selectedDate}
            onSelectDate={setSelectedDate}
          />
        </aside>

        {/* Main */}
        <main className="flex-1 overflow-y-auto p-6">
          {/* Day heading */}
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-800">
              {format(selectedDate, 'EEEE, MMM d')}
            </h2>
            <button
              onClick={() => setSelectedDate(new Date())}
              className="rounded-lg bg-indigo-50 px-3 py-1 text-sm font-medium text-indigo-600 hover:bg-indigo-100 transition"
            >
              Today
            </button>
          </div>

          <TaskList date={dateStr} />

          <div className="mt-8">
            <StatsChart />
          </div>
        </main>
      </div>
    </div>
  )
}
