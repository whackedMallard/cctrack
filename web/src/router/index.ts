import { createRouter, createWebHashHistory } from 'vue-router'

const Overview = () => import('../views/Overview.vue')
const Sessions = () => import('../views/Sessions.vue')
const Projects = () => import('../views/Projects.vue')
const SettingsView = () => import('../views/Settings.vue')
const RateCard = () => import('../views/RateCard.vue')

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: Overview, name: 'overview' },
    { path: '/sessions', component: Sessions, name: 'sessions' },
    { path: '/projects', component: Projects, name: 'projects' },
    { path: '/settings', component: SettingsView, name: 'settings' },
    { path: '/rates', component: RateCard, name: 'rates' },
  ],
})

export default router
