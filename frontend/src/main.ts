import { createApp } from 'vue'
import { createPinia } from 'pinia'

import '@fontsource-variable/inter'
import '@fontsource/noto-sans-myanmar/400.css'
import '@fontsource/noto-sans-myanmar/600.css'
import '@fontsource/noto-sans-myanmar/700.css'
import './styles/global.css'

import App from './App.vue'
import router from './router'
import { i18n } from './i18n'

const app = createApp(App)
app.use(createPinia())
app.use(i18n)
app.use(router)
app.mount('#app')
