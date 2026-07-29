import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { PageStatus } from '../../../shared/components/PageStatus'
import { recipeApi } from '../api'
import { normalizeImageForUpload } from '../image-processing'
import { RecipeForm } from '../components/RecipeForm'
import type { Recipe, RecipeInput } from '../types'
import styles from './RecipeEditorPage.module.css'

export function RecipeEditorPage() {
  const { recipeId } = useParams()
  const navigate = useNavigate()
  const [recipe, setRecipe] = useState<Recipe | null>(null)
  const [failed, setFailed] = useState(false)
  const editing = Boolean(recipeId)

  useEffect(() => {
    if (!recipeId) return
    let active = true
    recipeApi.get(recipeId).then((result) => { if (active) setRecipe(result) }).catch(() => { if (active) setFailed(true) })
    return () => { active = false }
  }, [recipeId])

  if (failed) return <PageStatus title="We couldn’t open the editor" message="The recipe may no longer exist." />
  if (editing && !recipe) return <PageStatus title="Preparing the editor…" />

  const initialValue: RecipeInput | undefined = recipe ? {
    name: recipe.name,
    recipeType: recipe.recipeType,
    timeToCook: recipe.timeToCook,
    description: recipe.description,
    ingredients: recipe.ingredients.map(({ name, quantity }) => ({ name, quantity })),
    steps: recipe.steps.map(({ instruction }) => ({ instruction })),
  } : undefined

  async function save(input: RecipeInput, images: File[]) {
    const saved = recipeId ? await recipeApi.update(recipeId, input) : await recipeApi.create(input)
    for (const image of images) {
      await recipeApi.uploadImage(saved.id, await normalizeImageForUpload(image))
    }
    navigate(`/recipes/${saved.id}`)
  }

  async function deleteImage(imageID: string) {
    if (!recipe) return
    if (!window.confirm('Remove this photo from the recipe?')) return
    await recipeApi.deleteImage(recipe.id, imageID)
    setRecipe({ ...recipe, images: recipe.images.filter((image) => image.id !== imageID) })
  }

  async function setCoverImage(imageID: string) {
    if (!recipe) return
    await recipeApi.setCoverImage(recipe.id, imageID)
    setRecipe({ ...recipe, images: recipe.images.map((image) => ({ ...image, isCover: image.id === imageID })) })
  }

  return (
    <section className={styles.page}>
      <header><Link to={recipeId ? `/recipes/${recipeId}` : '/recipes'}>← Cancel</Link><span>{editing ? 'Edit recipe' : 'New recipe'}</span><h1>{editing ? 'Make it even better.' : 'Add something delicious.'}</h1></header>
      <RecipeForm
        key={recipeId ?? 'new'}
        initialValue={initialValue}
        existingImages={recipe?.images}
        submitLabel={editing ? 'Save changes' : 'Add to cookbook'}
        recipeTypeLocked={editing}
        onDeleteImage={editing ? deleteImage : undefined}
        onSetCoverImage={editing ? setCoverImage : undefined}
        onSubmit={save}
      />
    </section>
  )
}
