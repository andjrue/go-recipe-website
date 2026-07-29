import { useEffect, useRef, useState, type FormEvent } from 'react'
import type { RecipeImage, RecipeInput } from '../types'
import { ImageEditor, type PendingImage } from './ImageEditor'
import { IngredientEditor, type IngredientRow } from './IngredientEditor'
import { StepEditor, type StepRow } from './StepEditor'
import styles from './RecipeForm.module.css'

type Props = {
  initialValue?: RecipeInput
  existingImages?: RecipeImage[]
  submitLabel: string
  onSubmit: (recipe: RecipeInput, images: File[]) => Promise<void>
  onDeleteImage?: (imageID: string) => Promise<void>
  onSetCoverImage?: (imageID: string) => Promise<void>
  recipeTypeLocked?: boolean
}

const blankRecipe: RecipeInput = {
  name: '', recipeType: 'structured', timeToCook: '', description: '',
  ingredients: [{ name: '', quantity: '' }], steps: [{ instruction: '' }],
}

export function RecipeForm({ initialValue = blankRecipe, existingImages = [], submitLabel, onSubmit, onDeleteImage, onSetCoverImage, recipeTypeLocked = false }: Props) {
  const [name, setName] = useState(initialValue.name)
  const [recipeType, setRecipeType] = useState<RecipeInput['recipeType']>(initialValue.recipeType)
  const [timeToCook, setTimeToCook] = useState(initialValue.timeToCook)
  const [description, setDescription] = useState(initialValue.description)
  const [ingredients, setIngredients] = useState<IngredientRow[]>(() => initialValue.ingredients.map((item) => ({ ...item, key: crypto.randomUUID() })))
  const [steps, setSteps] = useState<StepRow[]>(() => initialValue.steps.map((item) => ({ ...item, key: crypto.randomUUID() })))
  const [pendingImages, setPendingImages] = useState<PendingImage[]>([])
  const pendingImagesRef = useRef(pendingImages)
  pendingImagesRef.current = pendingImages
  const [submitting, setSubmitting] = useState(false)
  const [failed, setFailed] = useState(false)

  useEffect(() => () => {
    pendingImagesRef.current.forEach((image) => URL.revokeObjectURL(image.previewURL))
  }, [])

  function selectRecipeType(nextType: RecipeInput['recipeType']) {
    if (nextType === 'structured') {
      pendingImages.forEach((image) => URL.revokeObjectURL(image.previewURL))
      setPendingImages([])
    }
    setRecipeType(nextType)
    setFailed(false)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (recipeType === 'image' && existingImages.length === 0 && pendingImages.length === 0) {
      setFailed(true)
      return
    }
    setSubmitting(true)
    setFailed(false)
    try {
      await onSubmit({
        name, recipeType, timeToCook, description,
        ingredients: recipeType === 'structured' ? ingredients.map(({ name: itemName, quantity }) => ({ name: itemName, quantity })) : [],
        steps: recipeType === 'structured' ? steps.map(({ instruction }) => ({ instruction })) : [],
      }, pendingImages.map(({ file }) => file))
    } catch {
      setFailed(true)
      setSubmitting(false)
    }
  }

  return (
    <form className={styles.form} onSubmit={(event) => void handleSubmit(event)}>
      <section className={styles.basics}>
        <label><span>Recipe name</span><input required maxLength={200} value={name} placeholder="The dish everyone asks for" onChange={(event) => setName(event.target.value)} /></label>
        <div>
          <span className={styles.fieldLabel}>Recipe type</span>
          <div className={styles.typePicker} aria-label="Recipe type">
            <button type="button" disabled={recipeTypeLocked} aria-pressed={recipeType === 'structured'} onClick={() => selectRecipeType('structured')}>Type it out</button>
            <button type="button" disabled={recipeTypeLocked} aria-pressed={recipeType === 'image'} onClick={() => selectRecipeType('image')}>Use photos</button>
          </div>
        </div>
        <label><span>Cook time</span><input maxLength={100} value={timeToCook} placeholder="45 minutes" onChange={(event) => setTimeToCook(event.target.value)} /></label>
        <label><span>Description</span><textarea rows={4} maxLength={10000} value={description} placeholder="What makes this one worth keeping?" onChange={(event) => setDescription(event.target.value)} /></label>
      </section>
      {recipeType === 'structured' ? <>
        <IngredientEditor rows={ingredients} onChange={setIngredients} />
        <StepEditor rows={steps} onChange={setSteps} />
      </> : <ImageEditor existingImages={existingImages} pendingImages={pendingImages} onPendingChange={setPendingImages} onDeleteExisting={onDeleteImage} onSetCover={onSetCoverImage} />}
      <footer className={styles.footer}>
        {failed && <p role="alert">We couldn’t save this recipe. Check the fields and photos, then try again.</p>}
        <button type="submit" disabled={submitting}>{submitting ? 'Saving…' : submitLabel}</button>
      </footer>
    </form>
  )
}
