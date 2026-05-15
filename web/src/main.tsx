import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from './lib/auth-context'
import { getRouter } from './router'
import { getContext } from './integrations/tanstack-query/root-provider'

const { queryClient } = getContext()
// Router created once — auth injected per-render via RouterProvider context prop.
const router = getRouter(queryClient)

// Inner component has access to auth via useAuth; passes live state to router.
function InnerApp() {
  const auth = useAuth()
  return <RouterProvider router={router} context={{ auth }} />
}

const rootEl = document.getElementById('root')
if (!rootEl) throw new Error('Root element not found')

createRoot(rootEl).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider onClearCache={() => queryClient.clear()}>
        <InnerApp />
      </AuthProvider>
    </QueryClientProvider>
  </StrictMode>,
)
