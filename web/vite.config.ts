import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5178,
    strictPort: true,
    proxy: {
      "/ui": "http://127.0.0.1:8770",
      "/healthz": "http://127.0.0.1:8770",
    },
  },
  build: {
    outDir: "../internal/terminal/webdist",
    emptyOutDir: true,
    sourcemap: false,
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
  },
});
