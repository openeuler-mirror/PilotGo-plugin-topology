/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Oct 9 11:19:00 2023 +0800
 */
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import basicSsl from '@vitejs/plugin-basic-ssl'

const INVALID_CHAR_REGEX = /[\u0000-\u001F"#$&*+,:;<=>?[\]^`{|}\u007F]/g;
const DRIVE_LETTER_REGEX = /^[a-z]:/i;

// https://vitejs.dev/config/
export default defineConfig({
  base: "/plugin/topology",
  plugins: [
    vue(),basicSsl()
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server:{
    proxy:{
      "/plugin/topology/api": {
        target: 'http://10.41.107.29:9991',
        // target: 'http://10.44.55.72:9991',
        changeOrigin:true,
        // rewrite: (path)=> path.replace("/^\/api/", ""),
      },
      "/plugin/prometheus/api/v1": {
        target: 'http://10.44.55.72:8090',
        changeOrigin:true,
        // rewrite: (path)=> path.replace("/^\/api/", ""),
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        // https://github.com/rollup/rollup/blob/master/src/utils/sanitizeFileName.ts
        sanitizeFileName(name) {
          const match = DRIVE_LETTER_REGEX.exec(name);
          const driveLetter = match ? match[0] : "";
          // A `:` is only allowed as part of a windows drive letter (ex: C:\foo)
          // Otherwise, avoid them because they can refer to NTFS alternate data streams.
          return (
            driveLetter +
            name.slice(driveLetter.length).replace(INVALID_CHAR_REGEX, "")
          );
        },
      },
    },
  },
})
