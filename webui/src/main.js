// File: webui/src/main.js
import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'
import axios from './services/axios'
import './assets/main.css'
import './assets/dashboard.css'

axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000'

const app = createApp(App)
app.config.globalProperties.$axios = axios
app.use(router)
app.mount('#app')
