

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from "pinia-plugin-persistedstate"
import Antd from 'ant-design-vue';

import App from './App.vue'
import router from './router'
import 'ant-design-vue/dist/reset.css';

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(pinia)
app.use(router)
app.use(Antd)

app.mount('#app')
