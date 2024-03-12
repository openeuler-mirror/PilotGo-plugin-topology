import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(window.__MICRO_APP_BASE_ROUTE__ || '/'),
  routes: [
    {
      path: '/',
      redirect: '/topoList'
    },
    {
      path: '/topoList',
      name: 'topoList',
      component: () => import('../views/topoList.vue')
    },
    {
      path: '/topoDisplay',
      name: 'topoDisplay',
      component: () => import('../views/TopoDisplay.vue')
    }, {
      path: '/customTopo',
      name: 'customTopo',
      component: () => import('../views/customTopo.vue')
    }
  ]
})

export default router
