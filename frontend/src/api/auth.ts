import { api, request } from './client'
import type { User } from '@/types/api'

export const authApi = {
  requestOtp: (email: string) => api.post<void>('/auth/otp/request', { email }),
  verifyOtp: (email: string, code: string) =>
    api.post<{ user: User }>('/auth/otp/verify', { email, code }),
  logout: () => api.post<void>('/auth/logout'),
  me: () => request<{ user: User }>('GET', '/auth/me', { silent401: true }),
}
