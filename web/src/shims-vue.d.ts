/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhao_zhen_fang <zhaozhenfang@kylinos.cn>
 * Date: Wed Nov 15 17:28:46 2023 +0800
 */
/* eslint-disable */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}
interface Window {
    remount: any;
    unmount: any;
  readonly '__MICRO_APP_BASE_ROUTE__': string;
  __MICRO_APP_ENVIRONMENT__:any
}
declare module '*.png' {
  const value: string;
  export default value;
}

interface ImportMeta {
  env: {
    BASE_URL: string;
    // Add other environment variables here
  };
}


// fairy自定义
declare module '*.scss';
declare module '*.md';
declare module 'vue3-infinite-list' ;
declare module 'vue-search-highlight';
declare module 'vue-grid-layout' {
  import VueGridLayout from 'vue-grid-layout'
  export default VueGridLayout
}
declare module 'marked';
declare module '@kangc/v-md-editor/lib/preview'
declare module '@kangc/v-md-editor/lib/theme/vuepress.js'
declare module 'element-plus/dist/locale/zh-cn.mjs'

