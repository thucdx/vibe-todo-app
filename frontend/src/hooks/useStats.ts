import { useQuery } from '@tanstack/react-query'
import { statsApi } from '../api/client'
import type { MetricMode, ViewMode } from '../types'

export function useStats(view: ViewMode, metric: MetricMode) {
  return useQuery({
    queryKey: ['stats', view, metric],
    queryFn: () => statsApi.chart(view, metric).then(r => r.data ?? []),
  })
}
