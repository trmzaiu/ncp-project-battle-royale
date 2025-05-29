import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import process from 'process'

export default defineConfig({
  plugins: [react()],
  base: process.env.NODE_ENV === 'production' ? '/royaka-2025-fe/' : '/',
  server: {
    host: true,
    port: 3000,
  },
})
