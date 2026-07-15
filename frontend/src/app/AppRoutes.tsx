import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '../components/layout/AppShell'
import { RequireAuth } from '../features/auth/RequireAuth'
import { LoginPage } from '../features/auth/pages/LoginPage'
import { RecipeDetailPage } from '../features/recipes/pages/RecipeDetailPage'
import { RecipeEditorPage } from '../features/recipes/pages/RecipeEditorPage'
import { RecipeListPage } from '../features/recipes/pages/RecipeListPage'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<RequireAuth />}>
        <Route element={<AppShell />}>
          <Route index element={<Navigate to="/recipes" replace />} />
          <Route path="/recipes" element={<RecipeListPage />} />
          <Route path="/recipes/new" element={<RecipeEditorPage />} />
          <Route path="/recipes/:recipeId" element={<RecipeDetailPage />} />
          <Route path="/recipes/:recipeId/edit" element={<RecipeEditorPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/recipes" replace />} />
    </Routes>
  )
}
