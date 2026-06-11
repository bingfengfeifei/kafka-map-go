import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

function normalizeBasePath(basePath) {
  if (!basePath) {
    return './'
  }

  if (basePath === '/') {
    return '/'
  }

  const prefixed = basePath.startsWith('/') || basePath.startsWith('./') ? basePath : `/${basePath}`
  return prefixed.endsWith('/') ? prefixed : `${prefixed}/`
}

// https://vitejs.dev/config/
export default defineConfig({
  base: normalizeBasePath(process.env.VITE_BASE_PATH),
  plugins: [react()]
})
