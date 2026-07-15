import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'
import { RecipeForm } from './RecipeForm'

it('builds an ordered structured recipe payload', async () => {
  const user = userEvent.setup()
  const onSubmit = vi.fn().mockResolvedValue(undefined)
  render(<RecipeForm submitLabel="Add to cookbook" onSubmit={onSubmit} />)

  await user.type(screen.getByLabelText('Recipe name'), 'Tomato soup')
  await user.type(screen.getByLabelText('Cook time'), '30 minutes')
  await user.type(screen.getByLabelText('Description'), 'A weeknight favorite')
  await user.type(screen.getByLabelText('Ingredient'), 'Tomatoes')
  await user.type(screen.getByLabelText('Quantity'), '4')
  await user.click(screen.getByRole('button', { name: '+ Add ingredient' }))
  await user.type(screen.getAllByLabelText('Ingredient')[1], 'Salt')
  await user.type(screen.getAllByLabelText('Quantity')[1], 'to taste')
  await user.type(screen.getByLabelText('Instruction 1'), 'Simmer everything.')
  await user.click(screen.getByRole('button', { name: 'Add to cookbook' }))

  expect(onSubmit).toHaveBeenCalledWith({
    name: 'Tomato soup',
    recipeType: 'structured',
    timeToCook: '30 minutes',
    description: 'A weeknight favorite',
    ingredients: [
      { name: 'Tomatoes', quantity: '4' },
      { name: 'Salt', quantity: 'to taste' },
    ],
    steps: [{ instruction: 'Simmer everything.' }],
  })
})
