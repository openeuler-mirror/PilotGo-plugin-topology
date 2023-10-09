import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect:"/cluster",
    },
    {
      path: '/node',
      name: 'node',
      component: () => import('../views/NodeView.vue')
    },
    {
      path: '/cluster',
      name: 'cluster',
      component: () => import('../views/ClusterView.vue')
    }
  ]
})

export default router
