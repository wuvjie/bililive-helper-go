import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'fs'
import path from 'path'

const templatesDir = path.resolve(__dirname, '../templates')

export default defineConfig({
  plugins: [
    vue(),
    {
      name: 'clean-templates-assets',
      buildStart() {
        const assetsDir = path.join(templatesDir, 'assets')
        if (fs.existsSync(assetsDir)) {
          for (const f of fs.readdirSync(assetsDir)) {
            fs.unlinkSync(path.join(assetsDir, f))
          }
        }
      }
    }
  ],
  server: {
    proxy: {
      '/api': { target: 'http://localhost:5000', changeOrigin: true },
      '/login': { target: 'http://localhost:5000', changeOrigin: true },
      '/logout': { target: 'http://localhost:5000', changeOrigin: true }
    }
  },
  build: {
    outDir: '../templates',
    emptyOutDir: false
  }
})
