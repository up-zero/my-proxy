

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from "pinia-plugin-persistedstate"
import Antd from 'ant-design-vue';

import App from './App.vue'
import router from './router'
import 'ant-design-vue/dist/reset.css';
// 全局 a-table 包装组件：统一修复固定表头时表头与内容错位的问题
import FitTable from './components/FitTable.vue'

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(pinia)
app.use(router)
app.use(Antd)
// 覆盖 antd 全局注册的 ATable，使所有 <a-table> 生效
app.component('ATable', FitTable)

app.mount('#app')
