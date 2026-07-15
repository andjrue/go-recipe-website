import { apiRequest } from '../../shared/api/client'
import type { Recipe, RecipeInput, RecipeSummary } from './types'

export const recipeApi = {
  list: () => apiRequest<RecipeSummary[]>('/api/recipes'),
  get: (id: string) => apiRequest<Recipe>(`/api/recipes/${id}`),
  create: (recipe: RecipeInput) =>
    apiRequest<Recipe>('/api/recipes', { method: 'POST', body: JSON.stringify(recipe) }),
  update: (id: string, recipe: RecipeInput) =>
    apiRequest<Recipe>(`/api/recipes/${id}`, { method: 'PUT', body: JSON.stringify(recipe) }),
  remove: (id: string) => apiRequest<void>(`/api/recipes/${id}`, { method: 'DELETE' }),
}
