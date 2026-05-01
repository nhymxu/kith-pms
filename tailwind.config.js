/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/web/templates/**/*.templ",
    "./internal/web/templates/**/*.html",
  ],
  theme: {
    extend: {
      // PostHog-inspired palette from DESIGN.md §2
      colors: {
        // Primary text
        "olive-ink":     "#4d4f46",
        "deep-olive":    "#23251d",

        // Brand accent (hover only)
        "posthog-orange": "#F54E00",

        // Secondary accents
        "amber-gold":    "#F7A501",
        "gold-border":   "#b17816",

        // Surfaces
        "parchment":     "#fdfdf8",
        "sage-cream":    "#eeefe9",
        "light-sage":    "#e5e7e0",
        "warm-tan":      "#d4c9b8",
        "hover-white":   "#f4f4f4",

        // Neutrals
        "muted-olive":   "#65675e",
        "sage-placeholder": "#9ea096",
        "sage-border":   "#bfc1b7",
        "light-border":  "#b6b7af",

        // Dark CTA / high-contrast
        "near-black":    "#1e1f23",
        "dark-text":     "#111827",
      },
      fontFamily: {
        sans: [
          "IBM Plex Sans Variable",
          "IBM Plex Sans",
          "-apple-system",
          "system-ui",
          "Avenir Next",
          "Avenir",
          "Segoe UI",
          "Helvetica Neue",
          "Helvetica",
          "Ubuntu",
          "Roboto",
          "Noto",
          "Arial",
          "sans-serif",
        ],
        mono: [
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "Monaco",
          "Consolas",
          "Liberation Mono",
          "Courier New",
          "monospace",
        ],
      },
      borderRadius: {
        // PostHog uses tight radii: 2px, 4px, 6px, 9999px
        sm:   "2px",
        DEFAULT: "4px",
        md:   "6px",
        full: "9999px",
      },
    },
  },
  plugins: [],
};
