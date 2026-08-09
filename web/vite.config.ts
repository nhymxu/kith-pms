import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { defineConfig, loadEnv, mergeConfig } from "vite";
import { defineConfig as defineVitestConfig } from "vitest/config";

// Per-environment dev settings read from web/.env (see .env.example).
// Precedence: --port / --host CLI flags > VITE_* vars > built-in defaults.
const env = loadEnv(
	/** vite injects `mode`; vitest doesn't, so fall back to NODE_ENV. */
	import.meta.env?.MODE ?? process.env.NODE_ENV ?? "development",
	process.cwd(),
	"",
);
const devPort = Number(process.env.VITE_DEV_PORT ?? env.VITE_DEV_PORT ?? 3100);
const apiTarget =
	process.env.VITE_API_PROXY_TARGET ??
	env.VITE_API_PROXY_TARGET ??
	"http://localhost:8100";

const viteConfig = defineConfig({
	plugins: [
		tanstackRouter({ target: "react", autoCodeSplitting: true }),
		viteReact(),
		tailwindcss(),
	],
	server: {
		port: devPort,
		proxy: {
			"/v1": {
				target: apiTarget,
				changeOrigin: true,
			},
		},
	},
});

export default mergeConfig(
	viteConfig,
	defineVitestConfig({
		test: {
			environment: "jsdom",
			globals: true,
		},
	}),
);
