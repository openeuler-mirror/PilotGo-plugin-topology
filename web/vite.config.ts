import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  base: "/plugin/topology",
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server:{
    proxy:{
      "/plugin/topology/api": {
        // target: 'http://192.168.241.129:9991',
        target: 'http://10.44.55.72:9991',
        changeOrigin:true,
        // rewrite: (path)=> path.replace("/^\/api/", ""),
      },
      "/plugin/prometheus/api/v1": {
        target: 'http://10.44.55.72:8090',
        changeOrigin:true,
        // rewrite: (path)=> path.replace("/^\/api/", ""),
      }
    }
  }
})
