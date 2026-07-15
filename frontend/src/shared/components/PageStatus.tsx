import styles from './PageStatus.module.css'

type PageStatusProps = {
  title: string
  message?: string
  action?: { label: string; onClick: () => void }
}

export function PageStatus({ title, message, action }: PageStatusProps) {
  return (
    <section className={styles.status} aria-live="polite">
      <div className={styles.mark}>✦</div>
      <h1>{title}</h1>
      {message && <p>{message}</p>}
      {action && <button onClick={action.onClick}>{action.label}</button>}
    </section>
  )
}
