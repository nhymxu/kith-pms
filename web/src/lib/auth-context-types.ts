// Shared auth state type — imported by both auth-context.tsx and router.tsx
// to avoid circular imports.
import type { User } from "../schemas/auth"

export interface AuthState {
	user: User | null
	isLoading: boolean
	login: (password: string) => Promise<void>
	logout: () => Promise<void>
	logoutAll: () => Promise<void>
	refresh: () => Promise<void>
}
