import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { ApiError } from '../../shared/api/client'
import { authApi } from './api'
import { AuthContext, type AuthContextValue } from './auth-context'
import type { User } from './types'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [status, setStatus] = useState<AuthContextValue['status']>('loading')

  const refresh = useCallback(async () => {
    setStatus('loading')
    try {
      setUser(await authApi.currentUser())
      setStatus('ready')
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setUser(null)
        setStatus('ready')
        return
      }
      setStatus('error')
    }
  }, [])

  useEffect(() => { void refresh() }, [refresh])

  const logout = useCallback(async () => {
    await authApi.logout()
    setUser(null)
  }, [])

  const value = useMemo(() => ({ user, status, refresh, logout }), [user, status, refresh, logout])
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
