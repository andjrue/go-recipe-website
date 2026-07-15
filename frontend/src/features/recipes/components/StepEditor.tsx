import styles from './ListEditor.module.css'

export type StepRow = { key: string; instruction: string }

type Props = {
  rows: StepRow[]
  onChange: (rows: StepRow[]) => void
}

export function StepEditor({ rows, onChange }: Props) {
  return (
    <fieldset className={styles.fieldset}>
      <legend><span>02</span> Method</legend>
      <p>Keep each step focused on one part of the process.</p>
      <div className={styles.rows}>
        {rows.map((row, index) => (
          <div className={styles.stepRow} key={row.key}>
            <b>{String(index + 1).padStart(2, '0')}</b>
            <label><span>Instruction {index + 1}</span><textarea required rows={3} maxLength={5000} value={row.instruction} placeholder="Describe what happens next…" onChange={(event) => onChange(rows.map((item, rowIndex) => rowIndex === index ? { ...item, instruction: event.target.value } : item))} /></label>
            <button type="button" aria-label={`Remove step ${index + 1}`} disabled={rows.length === 1} onClick={() => onChange(rows.filter((_, rowIndex) => rowIndex !== index))}>×</button>
          </div>
        ))}
      </div>
      <button className={styles.add} type="button" onClick={() => onChange([...rows, { key: crypto.randomUUID(), instruction: '' }])}>+ Add step</button>
    </fieldset>
  )
}
