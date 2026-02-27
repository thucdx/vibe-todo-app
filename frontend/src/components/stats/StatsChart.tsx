import { useState } from 'react'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { useStats } from '../../hooks/useStats'
import type { MetricMode, ViewMode } from '../../types'

const VIEWS: ViewMode[]   = ['day', 'week', 'month']
const METRICS: MetricMode[] = ['count', 'points']

export default function StatsChart() {
  const [view, setView]     = useState<ViewMode>('day')
  const [metric, setMetric] = useState<MetricMode>('count')
  const { data = [], isLoading } = useStats(view, metric)

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-600">Productivity</h3>
        <div className="flex gap-2">
          {/* View toggles */}
          <div className="flex gap-1">
            {VIEWS.map(v => (
              <button
                key={v}
                onClick={() => setView(v)}
                className={[
                  'btn-toggle capitalize',
                  v === view ? 'btn-toggle-active' : 'btn-toggle-inactive',
                ].join(' ')}
              >
                {v}
              </button>
            ))}
          </div>

          {/* Metric toggles */}
          <div className="flex gap-1">
            {METRICS.map(m => (
              <button
                key={m}
                onClick={() => setMetric(m)}
                className={[
                  'btn-toggle capitalize',
                  m === metric ? 'btn-toggle-active' : 'btn-toggle-inactive',
                ].join(' ')}
              >
                {m}
              </button>
            ))}
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="h-48 animate-pulse rounded-xl bg-gray-100" />
      ) : (
        <ResponsiveContainer width="100%" height={200}>
          <AreaChart data={data} margin={{ top: 5, right: 10, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%"  stopColor="#6366f1" stopOpacity={0.25} />
                <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#f3f4f6" />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 11, fill: '#9ca3af' }}
              axisLine={false}
              tickLine={false}
            />
            <YAxis
              allowDecimals={false}
              tick={{ fontSize: 11, fill: '#9ca3af' }}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip
              contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 2px 8px rgba(0,0,0,.1)' }}
              formatter={(v: number) => [v, metric === 'count' ? 'Tasks done' : 'Points earned']}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke="#6366f1"
              strokeWidth={2}
              fill="url(#chartGrad)"
              dot={false}
              activeDot={{ r: 4, fill: '#6366f1' }}
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
