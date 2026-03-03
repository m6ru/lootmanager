import { createRouter, createWebHashHistory } from 'vue-router'
import Requirements from '../views/Search.vue'
import Hideout from '../views/Hideout.vue'
import Quests from '../views/Quests.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: Requirements },
    { path: '/hideout', component: Hideout },
    { path: '/quests', component: Quests },
  ]
})

export default router