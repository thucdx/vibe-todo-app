import { useQuery } from '@tanstack/react-query'
import { calendarApi } from '../api/client'

export function useCalendar(year: number, month: number) {
  return useQuery({
    queryKey: ['calendar', year, month],
    queryFn: () => calendarApi.summary(year, month).then(r => r.data ?? []),
  })
}
