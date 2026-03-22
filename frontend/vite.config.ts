import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// Service URLs
const PASSPORT_SERVICE_URL = 'http://127.0.0.1:8089'
const BACKEND_SERVICE_URL = 'http://127.0.0.1:8080'

// https://vite.dev/config/
export default defineConfig({
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
})
