<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import AppIcon from '../../components/AppIcon.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const username = ref('')
const password = ref('')
const show = ref(false)
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  if (!username.value.trim() || !password.value) { error.value = 'Vui lòng nhập tên đăng nhập và mật khẩu.'; return }
  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    const r = typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/noi-bo') ? route.query.redirect : '/noi-bo'
    router.replace(r)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="wrap">
    <RouterLink to="/" class="back"><AppIcon name="left" :size="18" /> Về trang chủ</RouterLink>
    <form class="card" novalidate @submit.prevent="submit">
      <img src="/favicon.svg" alt="" width="48" height="48" />
      <h1>Đăng nhập nội bộ</h1>
      <p class="muted">Dành cho nhân viên Đông Hải.</p>
      <div v-if="error" class="alert alert-error" role="alert"><AppIcon name="alert" :size="18" />{{ error }}</div>
      <div class="field">
        <label for="u">Tên đăng nhập</label>
        <input id="u" v-model="username" class="input" autocomplete="username" autocapitalize="none" spellcheck="false" autofocus />
      </div>
      <div class="field">
        <label for="p">Mật khẩu</label>
        <div class="pw">
          <input id="p" v-model="password" class="input" :type="show ? 'text' : 'password'" autocomplete="current-password" />
          <button type="button" class="eye" :aria-label="show ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'" @click="show = !show"><AppIcon :name="show ? 'eyeOff' : 'eye'" :size="18" /></button>
        </div>
      </div>
      <button class="btn btn-primary btn-block" type="submit" :disabled="loading"><span v-if="loading" class="spinner" />Đăng nhập</button>
    </form>
  </div>
</template>

<style scoped>
.wrap { min-height: 100vh; display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 24px 16px; background: linear-gradient(160deg, var(--sea-800), var(--sea-900)); }
.back { color: rgba(255,255,255,.8); display: flex; align-items: center; gap: 6px; margin-bottom: 18px; font-size: 14px; }
.back:hover { color: #fff; }
.card { width: 100%; max-width: 400px; background: #fff; border-radius: var(--r-xl); padding: 32px 28px; display: flex; flex-direction: column; gap: 16px; box-shadow: var(--shadow-3); }
h1 { font-size: 22px; margin-top: 4px; }
.card > p { margin-top: -10px; }
.pw { position: relative; }
.pw .input { padding-right: 48px; }
.eye { position: absolute; right: 4px; top: 50%; transform: translateY(-50%); width: 40px; height: 40px; border: 0; background: none; color: var(--ink-500); cursor: pointer; border-radius: 8px; display: grid; place-items: center; }
</style>
