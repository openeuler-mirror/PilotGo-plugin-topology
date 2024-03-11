import { createRouter, createWebHistory } from 'vue-router'
let baseRoute = '';
if (window.__MICRO_APP_ENVIRONMENT__) {
  console.log('在微前端环境中')
  baseRoute = window.__MICRO_APP_BASE_ROUTE__ || '/';
} else {
  baseRoute = import.meta.env.BASE_URL;
}

const router = createRouter({
  history: createWebHistory(baseRoute),
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
