import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: "../build/frontend",
    emptyOutDir: true,
  },
  server: {
    port: 34115,
    strictPort: true,
  },
})
