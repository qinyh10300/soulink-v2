import path from 'node:path'

import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const devApiProxyTarget =
    env.VITE_DEV_API_PROXY_TARGET?.trim() || 'http://127.0.0.1:18080'

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      proxy: {
        '/api/v1/ws': {
          target: devApiProxyTarget,
          changeOrigin: true,
          ws: true,
        },
        '/api': {
          target: devApiProxyTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
