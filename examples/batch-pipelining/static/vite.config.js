import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  build: {
    target: 'es2022', // Support for top-level await
    outDir: 'dist',
    rollupOptions: {
      input: {
        main: 'index.html'
      }
    }
  },
  esbuild: {
    target: 'es2022' // Also set esbuild target
  },
  optimizeDeps: {
    esbuildOptions: {
      target: 'es2022' // Also set for dependency optimization
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/rpc': {
        target: 'http://127.0.0.1:8000',
        changeOrigin: true
      }
    }
  }
})
