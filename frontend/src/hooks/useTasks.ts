import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../api/client'
import type { CreateTaskInput, Task, UpdateTaskInput } from '../types'

export function useTasks(date: string) {
  return useQuery({
    queryKey: ['tasks', date],
    queryFn: () => tasksApi.list(date).then(r => r.data ?? []),
    enabled: !!date,
  })
}

export function useCreateTask(date: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateTaskInput) => tasksApi.create(data).then(r => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks', date] })
      qc.invalidateQueries({ queryKey: ['calendar'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useUpdateTask(date: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTaskInput }) =>
      tasksApi.update(id, data).then(r => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks', date] })
      qc.invalidateQueries({ queryKey: ['calendar'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useToggleDone(date: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => tasksApi.toggleDone(id).then(r => r.data),
    // Optimistic update: flip done immediately
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ['tasks', date] })
      const previous = qc.getQueryData<Task[]>(['tasks', date])
      qc.setQueryData<Task[]>(['tasks', date], old =>
        (old ?? []).map(t => (t.id === id ? { ...t, done: !t.done } : t)),
      )
      return { previous }
    },
    onError: (_err, _id, ctx) => {
      qc.setQueryData(['tasks', date], ctx?.previous)
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ['tasks', date] })
      qc.invalidateQueries({ queryKey: ['calendar'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useDeleteTask(date: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => tasksApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks', date] })
      qc.invalidateQueries({ queryKey: ['calendar'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useMoveTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, task, newDate }: { id: string; task: Task; newDate: string }) =>
      tasksApi.update(id, { ...task, date: newDate }).then(r => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] })
      qc.invalidateQueries({ queryKey: ['calendar'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}
