import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ command }) => ({
  base: command === 'serve' ? '/' : '/dist/',
  plugins: [vue()],
  publicDir: false,
  build: {
    manifest: true,
    outDir: path.resolve(__dirname, 'public/dist'),
    rollupOptions: {
      input: 'resources/js/app.js',
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname,'/resources/js'),
    },
  },
}));
