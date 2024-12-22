import {createApp, reactive} from 'vue'
import router from './router'
import App from './App.vue'
import Component1 from './components/Component1.vue'
const app = createApp(App)
app.component("Component1", Component1);
app.use(router)
app.mount('#app')