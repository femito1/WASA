import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import Page1View from '../views/Page1View.vue'
import Page2View from '../views/Page2View.vue'
const router = createRouter({
history: createWebHashHistory(),
routes: [
{path: '/', component: HomeView},
{path: '/link1', component: Page1View},
{path: '/some/:id/link', component: Page2View},
]
});
export default router