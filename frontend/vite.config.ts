import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

const uiPort = Number(process.env.UI_PORT || 5175)
const apiPort = process.env.PORT || '4321'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '127.0.0.1',
    port: uiPort,
    strictPort: true,
    open: false,
    proxy: {
      '/api': {
        target: `http://127.0.0.1:${apiPort}`,
        changeOrigin: true,
        ws: true,
        timeout: 0,
        proxyTimeout: 0,
      },
      '/health': {
        target: `http://127.0.0.1:${apiPort}`,
        changeOrigin: true,
      },
    },
  },
})
