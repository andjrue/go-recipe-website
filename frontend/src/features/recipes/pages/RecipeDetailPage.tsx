import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { PageStatus } from '../../../shared/components/PageStatus'
import { recipeApi } from '../api'
import type { Recipe } from '../types'
import styles from './RecipeDetailPage.module.css'

export function RecipeDetailPage() {
  const { recipeId = '' } = useParams()
  const navigate = useNavigate()
  const [recipe, setRecipe] = useState<Recipe | null>(null)
  const [failed, setFailed] = useState(false)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    let active = true
    recipeApi.get(recipeId).then((result) => { if (active) setRecipe(result) }).catch(() => { if (active) setFailed(true) })
    return () => { active = false }
  }, [recipeId])

  async function handleDelete() {
    if (!recipe || !window.confirm(`Delete “${recipe.name}”? This removes it from the cookbook.`)) return
    setDeleting(true)
    try {
      await recipeApi.remove(recipe.id)
      navigate('/recipes')
    } catch {
      setDeleting(false)
      setFailed(true)
    }
  }

  if (failed) return <PageStatus title="We couldn’t open that recipe" message="It may have been removed, or the API may be unavailable." />
  if (!recipe) return <PageStatus title="Opening recipe…" />
  return (
    <article className={styles.recipe}>
      <Link className={styles.back} to="/recipes">← All recipes</Link>
      <header className={styles.hero}>
        <div>
          <div className={styles.eyebrow}>{recipe.recipeType === 'image' ? 'Recipe photos' : 'Structured recipe'}</div>
          <h1>{recipe.name}</h1>
          {recipe.description && <p>{recipe.description}</p>}
          <div className={styles.meta}><span>Cook time</span><strong>{recipe.timeToCook || 'Not listed'}</strong></div>
        </div>
        <div className={styles.actions}>
          <Link to={`/recipes/${recipe.id}/edit`}>Edit</Link>
          <button type="button" disabled={deleting} onClick={() => void handleDelete()}>{deleting ? 'Deleting…' : 'Delete'}</button>
        </div>
      </header>

      {recipe.recipeType === 'image' ? (
        <section className={styles.gallery} aria-label="Recipe photos">
          {recipe.images.map((image, index) => <img key={image.id} src={image.url} alt={`${recipe.name}, page ${index + 1}`} />)}
        </section>
      ) : <div className={styles.content}>
        <section className={styles.ingredients}>
          <span className={styles.sectionNumber}>01</span>
          <h2>Ingredients</h2>
          <ul>{recipe.ingredients.map((ingredient) => (
            <li key={ingredient.id}><strong>{ingredient.quantity}</strong><span>{ingredient.name}</span></li>
          ))}</ul>
        </section>
        <section className={styles.steps}>
          <span className={styles.sectionNumber}>02</span>
          <h2>Method</h2>
          <ol>{recipe.steps.map((step) => (
            <li key={step.id}><span>{String(step.stepNumber).padStart(2, '0')}</span><p>{step.instruction}</p></li>
          ))}</ol>
        </section>
      </div>}
    </article>
  )
}
