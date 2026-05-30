import { createApp } from 'vue'
import { router } from './router'
import App from './App.vue'

// 引入飞书底层 UI 框架样式
import './styles/base.css'

const app = createApp(App)
app.use(router)
app.mount('#app')
