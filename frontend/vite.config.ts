import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  // Set the third parameter to '' to load all env regardless of the `VITE_` prefix.
  const env = loadEnv(mode, process.cwd(), '')
  
  // Service URLs
  const PASSPORT_SERVICE_URL = env.VITE_PASSPORT_SERVICE_URL || 'http://127.0.0.1:8089'
  const BACKEND_SERVICE_URL = env.VITE_BACKEND_SERVICE_URL || 'http://127.0.0.1:8080'

  return {
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    proxy: {
      '/api/passport': {
        target: PASSPORT_SERVICE_URL,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/passport/, '/api/v1'),
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq) => {
            proxyReq.removeHeader('Origin')
          })
        }
      },
      '/api/sev': {
        target: BACKEND_SERVICE_URL,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/sev/, '/api/v1/sev'),
      },
    },
  },
}
})
