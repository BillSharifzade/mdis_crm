import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
// base можно переопределить для деплоя под суб-путём (например /mdis_crm/)
// через переменную окружения VITE_BASE_PATH во время сборки.
export default defineConfig({
  base: process.env.VITE_BASE_PATH || '/',
  plugins: [react()],
})
