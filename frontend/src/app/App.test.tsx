import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { App } from './App'

function jsonResponse(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  }))
}

describe('App', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    window.history.replaceState({}, '', '/')
  })

  it('sends an unauthenticated visitor to the login page', async () => {
    vi.stubGlobal('fetch', vi.fn(() => jsonResponse({ error: 'unauthorized' }, 401)))

    render(<App />)

    expect(await screen.findByRole('link', { name: 'Continue with Google' })).toHaveAttribute('href', '/api/auth/google/login')
  })

  it('shows the recipe collection for an authenticated user', async () => {
    vi.stubGlobal('fetch', vi.fn((input: string | URL | Request) => {
      const path = input.toString()
      if (path === '/api/me') {
        return jsonResponse({ id: 'user-1', email: 'cook@example.com', alias: 'Cook', role: 'user', dateJoined: '2026-07-13T00:00:00Z' })
      }
      if (path === '/api/recipes') return jsonResponse([])
      return jsonResponse({ error: 'not_found' }, 404)
    }))

    render(<App />)

    expect(await screen.findByRole('heading', { name: 'The cookbook is ready.' })).toBeInTheDocument()
    expect(screen.getByText('Cook')).toBeInTheDocument()
  })
})
