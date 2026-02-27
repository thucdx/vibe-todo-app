import axios from 'axios'
import type { AuthStatus, ChartPoint, CreateTaskInput, DaySummary, Task, UpdateTaskInput } from '../types'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api/v1',
  withCredentials: true,
})

export const authApi = {
  status: () => api.get<AuthStatus>('/auth/status'),
  setup:  (pin: string) => api.post('/auth/setup', { pin }),
  login:  (pin: string) => api.post('/auth/login', { pin }),
  logout: () => api.post('/auth/logout'),
}

export const tasksApi = {
  list:       (date: string) => api.get<Task[]>('/tasks', { params: { date } }),
  create:     (data: CreateTaskInput) => api.post<Task>('/tasks', data),
  update:     (id: string, data: UpdateTaskInput) => api.put<Task>(`/tasks/${id}`, data),
  toggleDone: (id: string) => api.patch<Task>(`/tasks/${id}/done`),
  delete:     (id: string) => api.delete(`/tasks/${id}`),
}

export const calendarApi = {
  summary: (year: number, month: number) =>
    api.get<DaySummary[]>('/calendar', { params: { year, month } }),
}

export const statsApi = {
  chart: (view: string, metric: string) =>
    api.get<ChartPoint[]>('/stats', { params: { view, metric } }),
}
