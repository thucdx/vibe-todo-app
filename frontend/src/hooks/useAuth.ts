import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authApi } from '../api/client'

export function useAuthStatus() {
  return useQuery({
    queryKey: ['auth', 'status'],
    queryFn: () => authApi.status().then(r => r.data),
    retry: false,
  })
}

export function useSetupPIN() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (pin: string) => authApi.setup(pin),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth'] }),
  })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (pin: string) => authApi.login(pin),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth'] }),
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => qc.clear(),
  })
}
