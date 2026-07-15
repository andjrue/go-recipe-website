import { createContext, useContext } from 'react'
import type { User } from './types'

export type AuthContextValue = {
  user: User | null
  status: 'loading' | 'ready' | 'error'
  refresh: () => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
