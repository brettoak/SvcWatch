import { createApp } from 'vue'
import { createPinia } from 'pinia'
import '@/assets/index.css'

import App from './App.vue'
import router from './router'
import { useServerStore } from '@/stores/server'

const app = createApp(App)

app.use(createPinia())
app.use(router)

const serverStore = useServerStore()
serverStore.loadConfig().finally(() => {
  app.mount('#app')
})

