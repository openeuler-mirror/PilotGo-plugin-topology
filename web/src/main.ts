/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Oct 9 11:19:00 2023 +0800
 */
import './styles/main.scss'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import VueGridLayout from 'vue-grid-layout'
import pinia from '@/stores/persist'
import { useLogStore } from './stores/log'

const app = createApp(App)

app.use(createPinia())
app.use(pinia);
app.use(router)
app.use(ElementPlus)
app.use(VueGridLayout)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

router.isReady().then(() => app.mount('#app'))

// 监听应用卸载生命周期
window.addEventListener('unmount',() => {
  useLogStore().clientId = 0;
})
