import { apiRequest } from '../../shared/api/client'
import type { User } from './types'

export const authApi = {
  currentUser: () => apiRequest<User>('/api/me'),
  logout: () => apiRequest<void>('/api/auth/logout', { method: 'POST' }),
}
