import { Navigate } from 'react-router-dom'
import { useAuth } from '../auth-context'
import styles from './LoginPage.module.css'

export function LoginPage() {
  const { user, status } = useAuth()
  if (status === 'ready' && user) return <Navigate to="/recipes" replace />

  return (
    <main className={styles.page}>
      <section className={styles.card}>
        <div className={styles.eyebrow}>A private collection</div>
        <h1>Recipes worth<br />coming back to.</h1>
        <p>One shared place for weeknight favorites, family recipes, and the dishes we’re still perfecting.</p>
        <a className={styles.login} href="/api/auth/google/login">Continue with Google</a>
        <small>Access is limited to invited email addresses.</small>
      </section>
      <aside className={styles.note} aria-hidden="true">
        <span>Tonight’s idea</span>
        <strong>Something warm,<br />simple, and good.</strong>
        <div className={styles.rule} />
        <em>Our Cookbook</em>
      </aside>
    </main>
  )
}
