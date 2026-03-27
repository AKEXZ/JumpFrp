import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 前台路由
    {
      path: '/',
      component: () => import('../views/user/HomeView.vue'),
    },
    {
      path: '/login',
      component: () => import('../views/user/LoginView.vue'),
    },
    {
      path: '/register',
      component: () => import('../views/user/RegisterView.vue'),
    },
    // 用户控制台（带导航布局）
    {
      path: '/dashboard',
      component: () => import('../views/user/UserLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', component: () => import('../views/user/DashboardView.vue') },
      ],
    },
    {
      path: '/tunnels',
      component: () => import('../views/user/UserLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', component: () => import('../views/user/TunnelsView.vue') },
      ],
    },
    {
      path: '/subdomains',
      component: () => import('../views/user/UserLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', component: () => import('../views/user/SubdomainsView.vue') },
      ],
    },
    {
      path: '/vip',
      component: () => import('../views/user/UserLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', component: () => import('../views/user/VIPView.vue') },
      ],
    },
    // 管理员路由
    {
      path: '/admin',
      component: () => import('../views/admin/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', component: () => import('../views/admin/DashboardView.vue') },
        { path: 'users', component: () => import('../views/admin/UsersView.vue') },
        { path: 'nodes', component: () => import('../views/admin/NodesView.vue') },
        { path: 'tunnels', component: () => import('../views/admin/TunnelsView.vue') },
        { path: 'subdomains', component: () => import('../views/admin/SubdomainsView.vue') },
        { path: 'orders', component: () => import('../views/admin/OrdersView.vue') },
        { path: 'settings', component: () => import('../views/admin/SettingsView.vue') },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.token) {
    next('/login')
  } else if (to.meta.requiresAdmin && !auth.isAdmin) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
