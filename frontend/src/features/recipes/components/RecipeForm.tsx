import { useState, type FormEvent } from 'react'
import type { RecipeInput } from '../types'
import { IngredientEditor, type IngredientRow } from './IngredientEditor'
import { StepEditor, type StepRow } from './StepEditor'
import styles from './RecipeForm.module.css'

type Props = {
  initialValue?: RecipeInput
  submitLabel: string
  onSubmit: (recipe: RecipeInput) => Promise<void>
}

const blankRecipe: RecipeInput = {
  name: '', recipeType: 'structured', timeToCook: '', description: '',
  ingredients: [{ name: '', quantity: '' }], steps: [{ instruction: '' }],
}

export function RecipeForm({ initialValue = blankRecipe, submitLabel, onSubmit }: Props) {
  const [name, setName] = useState(initialValue.name)
  const [timeToCook, setTimeToCook] = useState(initialValue.timeToCook)
  const [description, setDescription] = useState(initialValue.description)
  const [ingredients, setIngredients] = useState<IngredientRow[]>(() => initialValue.ingredients.map((item) => ({ ...item, key: crypto.randomUUID() })))
  const [steps, setSteps] = useState<StepRow[]>(() => initialValue.steps.map((item) => ({ ...item, key: crypto.randomUUID() })))
  const [submitting, setSubmitting] = useState(false)
  const [failed, setFailed] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setFailed(false)
    try {
      await onSubmit({
        name, recipeType: 'structured', timeToCook, description,
        ingredients: ingredients.map(({ name: itemName, quantity }) => ({ name: itemName, quantity })),
        steps: steps.map(({ instruction }) => ({ instruction })),
      })
    } catch {
      setFailed(true)
      setSubmitting(false)
    }
  }

  return (
    <form className={styles.form} onSubmit={(event) => void handleSubmit(event)}>
      <section className={styles.basics}>
        <label><span>Recipe name</span><input required autoFocus maxLength={200} value={name} placeholder="The dish everyone asks for" onChange={(event) => setName(event.target.value)} /></label>
        <div className={styles.pair}>
          <label><span>Cook time</span><input maxLength={100} value={timeToCook} placeholder="45 minutes" onChange={(event) => setTimeToCook(event.target.value)} /></label>
          <label><span>Recipe type</span><input value="Structured recipe" disabled /></label>
        </div>
        <label><span>Description</span><textarea rows={4} maxLength={10000} value={description} placeholder="What makes this one worth keeping?" onChange={(event) => setDescription(event.target.value)} /></label>
      </section>
      <IngredientEditor rows={ingredients} onChange={setIngredients} />
      <StepEditor rows={steps} onChange={setSteps} />
      <footer className={styles.footer}>
        {failed && <p role="alert">We couldn’t save this recipe. Check the fields and try again.</p>}
        <button type="submit" disabled={submitting}>{submitting ? 'Saving…' : submitLabel}</button>
      </footer>
    </form>
  )
}
