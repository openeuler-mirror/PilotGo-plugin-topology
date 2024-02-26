import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
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
  ]
})

export default router
