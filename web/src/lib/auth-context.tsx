// AuthProvider, useAuth, RequireAuth — session-cookie auth, no localStorage.

import { useNavigate } from "@tanstack/react-router";
import {
	createContext,
	type ReactNode,
	useCallback,
	useContext,
	useEffect,
	useRef,
	useState,
} from "react";
import {
	login as apiLogin,
	logout as apiLogout,
	logoutAll as apiLogoutAll,
	me,
} from "../endpoints/auth";
import type { User } from "../schemas/auth";
import { ApiError, onSessionLost } from "./api-client";
import type { AuthState } from "./auth-context-types";

export type { AuthState };

const AuthContext = createContext<AuthState | null>(null);
// Separate context so RequireAuth can read hasResolved without coupling to AuthState shape.
const HasResolvedContext = createContext<boolean>(false);

// ---- provider --------------------------------------------------------------

interface AuthProviderProps {
	children: ReactNode;
	onClearCache?: () => void;
	// AuthProvider mounts above RouterProvider (router context needs live auth
	// state), so useNavigate() isn't available here — the redirect is injected
	// from main.tsx via router.navigate instead.
	onSessionCleared?: () => void;
}

export function AuthProvider({
	children,
	onClearCache,
	onSessionCleared,
}: AuthProviderProps) {
	const [user, setUser] = useState<User | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	// hasResolved flips true after the first /v1/auth/me settles (success or 401).
	// RequireAuth gates navigation on this — never redirects before the fetch completes.
	const [hasResolved, setHasResolved] = useState(false);

	// Stable reference so the session-lost subscription doesn't re-subscribe on every render.
	const stableOnClearCache = useCallback(() => {
		onClearCache?.();
	}, [onClearCache]);

	const clearSession = useCallback(() => {
		setUser(null);
		stableOnClearCache();
		onSessionCleared?.();
	}, [stableOnClearCache, onSessionCleared]);

	// Fetch current user on mount
	const fetchMe = useCallback(async () => {
		setIsLoading(true);
		try {
			const u = await me();
			setUser(u);
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				setUser(null);
			}
		} finally {
			setIsLoading(false);
			setHasResolved(true);
		}
	}, []);

	// biome-ignore lint/correctness/useExhaustiveDependencies: fetchMe is stable; run on mount only
	useEffect(() => {
		void fetchMe();
	}, []);

	// Subscribe to session-lost bus (fired by apiFetch on any 401).
	// stableOnClearCache won't change unless the parent passes a new onClearCache identity.
	const cleanupRef = useRef<(() => void) | null>(null);
	useEffect(() => {
		cleanupRef.current = onSessionLost(clearSession);
		return () => {
			cleanupRef.current?.();
		};
	}, [clearSession]);

	const login = async (password: string) => {
		await apiLogin(password);
		await fetchMe();
	};

	const logout = async () => {
		await apiLogout();
		clearSession();
	};

	const logoutAll = async () => {
		await apiLogoutAll();
		clearSession();
	};

	const refresh = async () => {
		await fetchMe();
	};

	return (
		<AuthContext.Provider
			value={{ user, isLoading, login, logout, logoutAll, refresh }}
		>
			{/* hasResolved exposed via context below for RequireAuth */}
			<HasResolvedContext.Provider value={hasResolved}>
				{children}
			</HasResolvedContext.Provider>
		</AuthContext.Provider>
	);
}

// ---- hook ------------------------------------------------------------------

export function useAuth(): AuthState {
	const ctx = useContext(AuthContext);
	if (!ctx) throw new Error("useAuth must be used within AuthProvider");
	return ctx;
}

// ---- route guard -----------------------------------------------------------

interface RequireAuthProps {
	children: ReactNode;
}

export function RequireAuth({ children }: RequireAuthProps) {
	const { user, isLoading } = useAuth();
	// hasResolved distinguishes "fetch finished with no user" from "fetch not yet started".
	const hasResolved = useContext(HasResolvedContext);
	const navigate = useNavigate();

	useEffect(() => {
		// Only redirect once the initial /v1/auth/me has actually settled.
		if (hasResolved && !isLoading && user === null) {
			navigate({ to: "/login" });
		}
	}, [hasResolved, isLoading, user, navigate]);

	if (!hasResolved || isLoading) {
		// Minimal placeholder — Phase 4 will replace with a proper skeleton
		return (
			<div
				style={{
					display: "flex",
					alignItems: "center",
					justifyContent: "center",
					height: "100vh",
				}}
			>
				Loading…
			</div>
		);
	}

	if (user === null) {
		return null;
	}

	return <>{children}</>;
}
