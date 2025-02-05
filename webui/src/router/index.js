// File: webui/src/router/index.js
import { createRouter, createWebHashHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ChatListView from '../views/ChatListView.vue'
import ChatView from '../views/ChatView.vue'
import ContactsView from '../views/ContactsView.vue' // Optional

const routes = [
  { path: '/', name: 'Login', component: LoginView, meta: { public: true } },
  { path: '/chats', name: 'ChatList', component: ChatListView },
  { path: '/chat/:convId', name: 'Chat', component: ChatView, props: true },
  { path: '/contacts', name: 'Contacts', component: ContactsView } // Optional
];

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes
});

// Global navigation guard: if route is not public, ensure token exists.
router.beforeEach((to, from, next) => {
  const isPublic = to.meta.public || false;
  const token = localStorage.getItem('authToken');
  if (!isPublic && !token) {
    return next({ name: 'Login' });
  }
  next();
});

export default router;
