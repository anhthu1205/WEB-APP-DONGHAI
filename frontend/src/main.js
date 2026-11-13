import { createApp } from 'vue'
import { createPinia } from 'pinia'
import '@fontsource/be-vietnam-pro/400.css'
import '@fontsource/be-vietnam-pro/500.css'
import '@fontsource/be-vietnam-pro/600.css'
import '@fontsource/be-vietnam-pro/700.css'
import '@fontsource/be-vietnam-pro/800.css'
import './assets/main.css'
import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler } from './api/http'
import { useAuthStore } from './stores/auth'
import { reveal } from './directives/reveal'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.directive('reveal', reveal)

// 401 bất kỳ lúc nào (token hết hạn / bị khóa) → xóa token, về trang đăng nhập
setUnauthorizedHandler(() => {
  const auth = useAuthStore()
  auth.logout()
  const cur = router.currentRoute.value
  if (cur.matched.some((r) => r.meta.requiresAuth)) {
    router.replace({ name: 'login', query: { redirect: cur.fullPath } })
  }
})

app.mount('#app')
