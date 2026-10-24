import { createRouter, createWebHistory } from 'vue-router'

import { Permission } from '@scaffold/shared'

import DefaultLayout from '../layouts/DefaultLayout.vue'
import { useAuthStore } from '../stores/auth'
import { usePluginStore } from '../stores/plugin'
import { featureRoutes } from './features.gen'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true, titleKey: 'login.title' },
    },
    {
      path: '/oauth/callback',
      name: 'oauth-callback',
      component: () => import('../views/OAuthCallbackView.vue'),
      meta: { public: true, titleKey: 'oauth.callbackTitle' },
    },
    {
      path: '/',
      name: 'root',
      component: DefaultLayout,
      redirect: '/dashboard',
      children: [
        ...featureRoutes,
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('../views/DashboardView.vue'),
          meta: { titleKey: 'nav.dashboard' },
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('../views/UsersView.vue'),
          meta: { titleKey: 'nav.users', permission: Permission.UserList },
        },
        {
          path: 'roles',
          name: 'roles',
          component: () => import('../views/RolesView.vue'),
          meta: { titleKey: 'nav.roles', permission: Permission.RoleList },
        },
        {
          path: 'permissions',
          name: 'permissions',
          component: () => import('../views/PermissionsView.vue'),
          meta: { titleKey: 'nav.permissions', permission: Permission.RolePermissionList },
        },
        {
          path: 'plugins',
          name: 'plugins',
          component: () => import('../views/PluginListView.vue'),
          meta: { titleKey: 'nav.plugins', permission: Permission.PluginList },
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('../views/ProfileView.vue'),
          meta: { titleKey: 'nav.profile' },
        },
        {
          path: '403',
          name: 'forbidden',
          component: () => import('../views/ForbiddenView.vue'),
        },
        {
          path: ':pathMatch(.*)*',
          name: 'not-found',
          component: () => import('../views/NotFoundView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true

  await auth.restore()
  if (!auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (auth.user?.mustChangePassword && to.name !== 'profile') {
    return { name: 'profile', query: { changePassword: '1' } }
  }

  const plugins = usePluginStore()
  if (!plugins.loaded) {
    try {
      await plugins.loadAll(router)
    } catch (error) {
      console.error('failed to load plugins', error)
    }
    if (to.name === 'not-found' && router.resolve(to.fullPath).name !== 'not-found') {
      return to.fullPath
    }
  }

  if (to.meta.permission && !auth.hasPermission(to.meta.permission)) {
    return { name: 'forbidden' }
  }
  return true
})

export default router
