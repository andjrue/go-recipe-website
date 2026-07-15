import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageStatus } from '../../../shared/components/PageStatus'
import { recipeApi } from '../api'
import { RecipeCard } from '../components/RecipeCard'
import type { RecipeSummary } from '../types'
import styles from './RecipeListPage.module.css'

export function RecipeListPage() {
  const [recipes, setRecipes] = useState<RecipeSummary[] | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let active = true
    recipeApi.list().then((result) => { if (active) setRecipes(result) }).catch(() => { if (active) setFailed(true) })
    return () => { active = false }
  }, [])

  if (failed) return <PageStatus title="The recipes stayed on the shelf" message="We couldn’t load them right now." />
  if (!recipes) return <PageStatus title="Gathering recipes…" />

  return (
    <section>
      <header className={styles.heading}>
        <div><span>Shared favorites</span><h1>Our recipes</h1></div>
        <Link to="/recipes/new">Add a recipe</Link>
      </header>
      {recipes.length === 0 ? (
        <div className={styles.empty}>
          <h2>The cookbook is ready.</h2>
          <p>Add the first recipe—the one you already know by heart is a good place to start.</p>
          <Link to="/recipes/new">Add the first recipe</Link>
        </div>
      ) : (
        <div className={styles.grid}>{recipes.map((recipe) => <RecipeCard key={recipe.id} recipe={recipe} />)}</div>
      )}
    </section>
  )
}
