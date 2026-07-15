import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../../features/auth/auth-context'
import styles from './AppShell.module.css'

export function AppShell() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login')
  }

  return (
    <div className={styles.app}>
      <header className={styles.header}>
        <NavLink to="/recipes" className={styles.brand}>Our Cookbook</NavLink>
        <nav aria-label="Main navigation">
          <NavLink to="/recipes">Recipes</NavLink>
          <NavLink to="/recipes/new" className={styles.add}>Add recipe</NavLink>
        </nav>
        <div className={styles.account}>
          <span>{user?.alias || user?.email}</span>
          <button type="button" onClick={() => void handleLogout()}>Sign out</button>
        </div>
      </header>
      <main className={styles.main}><Outlet /></main>
    </div>
  )
}
