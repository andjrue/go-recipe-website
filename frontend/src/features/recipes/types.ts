export type Ingredient = {
  id: string
  name: string
  quantity: string
  position: number
}

export type RecipeStep = {
  id: string
  stepNumber: number
  instruction: string
}

export type RecipeSummary = {
  id: string
  name: string
  recipeType: 'structured' | 'image'
  timeToCook: string
  description: string
  userID: string
  datePosted: string
  lastEditedAt: string
}

export type Recipe = RecipeSummary & {
  ingredients: Ingredient[]
  steps: RecipeStep[]
  images: RecipeImage[]
}

export type RecipeImage = {
  id: string
  fileName: string
  contentType: string
  fileSize: number
  position: number
  isCover: boolean
  uploadedAt: string
}

export type RecipeInput = {
  name: string
  recipeType: 'structured'
  timeToCook: string
  description: string
  ingredients: { name: string; quantity: string }[]
  steps: { instruction: string }[]
}
