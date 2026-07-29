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
        <nav className={styles.navigation} aria-label="Main navigation">
          <NavLink to="/recipes"><span aria-hidden="true">⌂</span><b>Recipes</b></NavLink>
          <NavLink to="/recipes/new" className={styles.add}><span aria-hidden="true">＋</span><b>Add recipe</b></NavLink>
        </nav>
        <div className={styles.account}>
          <span>{user?.alias || user?.email}</span>
          <button type="button" onClick={() => void handleLogout()} aria-label={`Sign out${user?.alias ? ` ${user.alias}` : ''}`}>Sign out</button>
        </div>
      </header>
      <main className={styles.main}><Outlet /></main>
    </div>
  )
}
