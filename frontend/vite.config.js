import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Dev: Vite :5173 proxy /api → Go :8080 (cùng origin, không cần CORS).
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: false },
    },
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    chunkSizeWarningLimit: 600,
  },
})
