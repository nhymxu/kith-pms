import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { defineConfig, mergeConfig } from "vite";
import { defineConfig as defineVitestConfig } from "vitest/config";

const viteConfig = defineConfig({
	plugins: [
		tanstackRouter({ target: "react", autoCodeSplitting: true }),
		viteReact(),
		tailwindcss(),
	],
	server: {
		proxy: {
			"/v1": {
				target: "http://localhost:8000",
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
