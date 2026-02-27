export type Priority = 'high' | 'medium' | 'low'
export type ViewMode = 'day' | 'week' | 'month'
export type MetricMode = 'count' | 'points'

export interface Task {
  id: string
  title: string
  date: string      // ISO date string from server
  due_time: string | null  // "HH:MM" or null
  priority: Priority
  tags: string[]
  points: number
  done: boolean
  created_at: string
  updated_at: string
}

export interface CreateTaskInput {
  title: string
  date: string           // YYYY-MM-DD
  due_time?: string | null
  priority?: Priority
  tags?: string[]
  points?: number
}

export interface UpdateTaskInput {
  title: string
  date: string
  due_time?: string | null
  priority?: Priority
  tags?: string[]
  points?: number
}

export interface DaySummary {
  date: string   // YYYY-MM-DD
  done: number
  total: number
}

export interface ChartPoint {
  label: string
  value: number
}

export interface AuthStatus {
  configured: boolean
  authenticated: boolean
}
