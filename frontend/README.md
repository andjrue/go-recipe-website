# Cookbook frontend

React and TypeScript frontend for the private cookbook. Vite serves the development app and
proxies `/api` to the Go service on port 8080.

## Structure

```text
src/
  app/                 application composition and routes
  components/layout/   shared application chrome
  features/auth/       session API, provider, route guard, and login page
  features/recipes/    recipe API, types, components, and route pages
  shared/              cross-feature HTTP and status components
  styles/              global tokens and element defaults
```

Pages load data and handle navigation. Feature components own reusable UI behavior. Network
calls stay in each feature's `api.ts`; shared HTTP mechanics stay in `shared/api/client.ts`.

## Commands

```bash
npm install
npm run dev
npm test
npm run lint
npm run build
```
