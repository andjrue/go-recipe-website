import { useState, type ChangeEvent } from 'react'
import type { RecipeImage } from '../types'
import styles from './ImageEditor.module.css'

export type PendingImage = { key: string; file: File; previewURL: string }

type Props = {
  existingImages: RecipeImage[]
  pendingImages: PendingImage[]
  onPendingChange: (images: PendingImage[]) => void
  onDeleteExisting?: (imageID: string) => Promise<void>
  onSetCover?: (imageID: string) => Promise<void>
}

export function ImageEditor({ existingImages, pendingImages, onPendingChange, onDeleteExisting, onSetCover }: Props) {
  const [busyImageID, setBusyImageID] = useState('')
  const [actionFailed, setActionFailed] = useState(false)

  function addFiles(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []).filter((file) => file.type.startsWith('image/') || /\.(heic|heif)$/i.test(file.name))
    const additions = files.map((file) => ({ key: crypto.randomUUID(), file, previewURL: URL.createObjectURL(file) }))
    onPendingChange([...pendingImages, ...additions])
    event.target.value = ''
  }

  function removePending(key: string) {
    const image = pendingImages.find((item) => item.key === key)
    if (image) URL.revokeObjectURL(image.previewURL)
    onPendingChange(pendingImages.filter((item) => item.key !== key))
  }

  function movePending(index: number, direction: -1 | 1) {
    const nextIndex = index + direction
    if (nextIndex < 0 || nextIndex >= pendingImages.length) return
    const reordered = [...pendingImages]
    const current = reordered[index]
    reordered[index] = reordered[nextIndex]
    reordered[nextIndex] = current
    onPendingChange(reordered)
  }

  async function runExistingAction(imageID: string, action: (id: string) => Promise<void>) {
    setBusyImageID(imageID)
    setActionFailed(false)
    try { await action(imageID) } catch { setActionFailed(true) } finally { setBusyImageID('') }
  }

  return (
    <fieldset className={styles.fieldset}>
      <legend><span>01</span> Recipe photos</legend>
      <p>Photograph a recipe card or choose photos from your library. The first new photo becomes the cover.</p>
      <div className={styles.pickers}>
        <label className={styles.picker}>Take a photo<input type="file" accept="image/*,.heic,.heif" capture="environment" onChange={addFiles} /></label>
        <label className={styles.pickerSecondary}>Choose photos<input type="file" accept="image/*,.heic,.heif" multiple onChange={addFiles} /></label>
      </div>
      {actionFailed && <p role="alert">We couldn’t update that photo. Please try again.</p>}

      {(existingImages.length > 0 || pendingImages.length > 0) && (
        <div className={styles.grid} aria-label="Recipe photos">
          {existingImages.map((image) => (
            <article className={styles.photo} key={image.id}>
              <img src={image.url} alt="Uploaded recipe" />
              <div className={styles.photoActions}>
                {image.isCover ? <strong>Cover</strong> : onSetCover && <button type="button" disabled={busyImageID !== ''} onClick={() => void runExistingAction(image.id, onSetCover)}>Make cover</button>}
                {onDeleteExisting && <button className={styles.remove} type="button" disabled={busyImageID !== ''} onClick={() => void runExistingAction(image.id, onDeleteExisting)}>Remove</button>}
              </div>
            </article>
          ))}
          {pendingImages.map((image, index) => (
            <article className={styles.photo} key={image.key}>
              <img src={image.previewURL} alt={`New recipe upload ${index + 1}`} />
              <div className={styles.photoActions}>
                {existingImages.length === 0 && index === 0 && <strong>Cover</strong>}
                <span className={styles.order}>
                  <button type="button" aria-label={`Move photo ${index + 1} earlier`} disabled={index === 0} onClick={() => movePending(index, -1)}>←</button>
                  <button type="button" aria-label={`Move photo ${index + 1} later`} disabled={index === pendingImages.length - 1} onClick={() => movePending(index, 1)}>→</button>
                </span>
                <button className={styles.remove} type="button" onClick={() => removePending(image.key)}>Remove</button>
              </div>
            </article>
          ))}
        </div>
      )}
    </fieldset>
  )
}
