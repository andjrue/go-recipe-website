const maxImageEdge = 2400

export async function normalizeImageForUpload(file: File): Promise<File> {
  const source = await loadImage(file)
  const scale = Math.min(1, maxImageEdge / Math.max(source.width, source.height))
  const width = Math.max(1, Math.round(source.width * scale))
  const height = Math.max(1, Math.round(source.height * scale))
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const context = canvas.getContext('2d')
  if (!context) throw new Error('image_processing_unavailable')
  context.fillStyle = '#fffdf8'
  context.fillRect(0, 0, width, height)
  context.drawImage(source.image, 0, 0, width, height)
  source.close()

  const blob = await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((result) => result ? resolve(result) : reject(new Error('image_conversion_failed')), 'image/jpeg', 0.88)
  })
  const baseName = file.name.replace(/\.[^.]+$/, '') || 'recipe-photo'
  return new File([blob], `${baseName}.jpg`, { type: 'image/jpeg', lastModified: file.lastModified })
}

async function loadImage(file: File): Promise<{ image: CanvasImageSource; width: number; height: number; close: () => void }> {
  if ('createImageBitmap' in window) {
    const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' })
    return { image: bitmap, width: bitmap.width, height: bitmap.height, close: () => bitmap.close() }
  }

  const url = URL.createObjectURL(file)
  const image = new Image()
  image.src = url
  await image.decode()
  return {
    image,
    width: image.naturalWidth,
    height: image.naturalHeight,
    close: () => URL.revokeObjectURL(url),
  }
}
