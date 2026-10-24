import { createRouter, createWebHistory } from 'vue-router'

import DefaultLayout from '../layouts/DefaultLayout.vue'
import { usePluginStore } from '../stores/plugin'
import { featureRoutes } from './features.gen'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'root',
      component: DefaultLayout,
      children: [
        ...featureRoutes,
        {
          path: '',
          name: 'home',
          component: () => import('../views/HomeView.vue'),
          meta: { titleKey: 'nav.home' },
        },
        {
          path: 'plugins',
          name: 'plugins',
          component: () => import('../views/PluginListView.vue'),
          meta: { titleKey: 'nav.plugins' },
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
  return true
})

export default router
