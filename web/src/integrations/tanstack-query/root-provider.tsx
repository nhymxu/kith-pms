import { createQueryClient } from '../../lib/query-client'

// Singleton — createQueryClient is called once per app lifetime.
let _queryClient: ReturnType<typeof createQueryClient> | null = null

export function getContext() {
  if (!_queryClient) {
    _queryClient = createQueryClient()
  }
  return { queryClient: _queryClient }
}
