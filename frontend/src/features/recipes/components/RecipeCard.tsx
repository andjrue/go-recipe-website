import { Link } from 'react-router-dom'
import type { RecipeSummary } from '../types'
import styles from './RecipeCard.module.css'

export function RecipeCard({ recipe }: { recipe: RecipeSummary }) {
  return (
    <article className={styles.card}>
      <Link to={`/recipes/${recipe.id}`} aria-label={`View ${recipe.name}`}>
        <div className={styles.visual}>
          <span>{recipe.recipeType === 'image' ? 'Recipe card' : 'From the kitchen'}</span>
          <b>{recipe.name.slice(0, 1).toUpperCase()}</b>
        </div>
        <div className={styles.body}>
          <div className={styles.meta}>{recipe.timeToCook || 'Time not listed'}</div>
          <h2>{recipe.name}</h2>
          <p>{recipe.description || 'A recipe waiting for its story.'}</p>
          <span className={styles.open}>Open recipe <span aria-hidden="true">→</span></span>
        </div>
      </Link>
    </article>
  )
}
