import type { Config } from "tailwindcss";
import forms from "@tailwindcss/forms";
import typography from "@tailwindcss/typography";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#eff6ff",
          100: "#dbeafe",
          200: "#bfdbfe",
          300: "#93c5fd",
          400: "#60a5fa",
          500: "#3b82f6",
          600: "#2563eb",
          700: "#1d4ed8",
          800: "#1e40af",
          900: "#1e3a8a",
        },
        muted: "#f5f7fb",
        success: "#16a34a",
        warn: "#f59e0b",
        error: "#dc2626",
      },
      borderRadius: {
        lg: "12px",
      },
    },
  },
  plugins: [forms, typography],
} satisfies Config;
