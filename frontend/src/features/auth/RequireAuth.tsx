import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { PageStatus } from '../../shared/components/PageStatus'
import { useAuth } from './auth-context'

export function RequireAuth() {
  const { user, status, refresh } = useAuth()
  const location = useLocation()

  if (status === 'loading') return <PageStatus title="Opening the cookbook…" />
  if (status === 'error') {
    return <PageStatus title="We couldn’t reach the kitchen" message="Check that the API is running and try again." action={{ label: 'Try again', onClick: () => void refresh() }} />
  }
  if (!user) return <Navigate to="/login" replace state={{ from: location }} />
  return <Outlet />
}
