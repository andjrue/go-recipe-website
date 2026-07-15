import styles from './ListEditor.module.css'

export type IngredientRow = { key: string; name: string; quantity: string }

type Props = {
  rows: IngredientRow[]
  onChange: (rows: IngredientRow[]) => void
}

export function IngredientEditor({ rows, onChange }: Props) {
  function update(index: number, field: 'name' | 'quantity', value: string) {
    onChange(rows.map((row, rowIndex) => rowIndex === index ? { ...row, [field]: value } : row))
  }

  return (
    <fieldset className={styles.fieldset}>
      <legend><span>01</span> Ingredients</legend>
      <p>Add these in the order you’d reach for them.</p>
      <div className={styles.rows}>
        {rows.map((row, index) => (
          <div className={styles.ingredientRow} key={row.key}>
            <label><span>Quantity</span><input value={row.quantity} maxLength={100} placeholder="2 cups" onChange={(event) => update(index, 'quantity', event.target.value)} /></label>
            <label><span>Ingredient</span><input required value={row.name} maxLength={200} placeholder="All-purpose flour" onChange={(event) => update(index, 'name', event.target.value)} /></label>
            <button type="button" aria-label={`Remove ingredient ${index + 1}`} disabled={rows.length === 1} onClick={() => onChange(rows.filter((_, rowIndex) => rowIndex !== index))}>×</button>
          </div>
        ))}
      </div>
      <button className={styles.add} type="button" onClick={() => onChange([...rows, { key: crypto.randomUUID(), name: '', quantity: '' }])}>+ Add ingredient</button>
    </fieldset>
  )
}
