import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../src/frontend/dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/v1": "http://localhost:8080",
    },
  },
  // Also write to src/frontend/dist during development
  experimental: {
    renderBuilderOutput: "../src/frontend/dist",
  },
  // Watch the dist directory for changes
  watch: {
    include: ["src/**", "public/**"],
  },
});
