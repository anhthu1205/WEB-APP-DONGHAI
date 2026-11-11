<script setup>
import { reactive, ref } from 'vue'
import { authApi } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'

const auth = useAuthStore()
const toast = useToastStore()
const f = reactive({ current: '', next: '', confirm: '' })
const errs = ref({})
const busy = ref(false)

async function save() {
  errs.value = {}
  if (f.next !== f.confirm) { errs.value = { confirm: 'Mật khẩu nhập lại không khớp' }; return }
  busy.value = true
  try {
    const res = await authApi.changePassword(f.current, f.next)
    auth.replaceToken(res.token)
    f.current = f.next = f.confirm = ''
    toast.success('Đã đổi mật khẩu. Các phiên đăng nhập khác đã bị đăng xuất.')
  } catch (e) {
    errs.value = e.fields && Object.keys(e.fields).length ? e.fields : { _: e.message }
  } finally { busy.value = false }
}
</script>

<template>
  <div>
    <div class="page-head"><div><h1>Đổi mật khẩu</h1><p>Mật khẩu từ 8 đến 72 ký tự.</p></div></div>
    <form class="panel panel-pad form-grid" style="max-width:480px" novalidate @submit.prevent="save">
      <div v-if="errs._" class="alert alert-error" role="alert">{{ errs._ }}</div>
      <div class="field"><label for="cp1">Mật khẩu hiện tại</label><input id="cp1" v-model="f.current" class="input" type="password" autocomplete="current-password" :class="{ invalid: errs.current_password }" /><span v-if="errs.current_password" class="err">{{ errs.current_password }}</span></div>
      <div class="field"><label for="cp2">Mật khẩu mới</label><input id="cp2" v-model="f.next" class="input" type="password" autocomplete="new-password" :class="{ invalid: errs.new_password }" /><span v-if="errs.new_password" class="err">{{ errs.new_password }}</span></div>
      <div class="field"><label for="cp3">Nhập lại mật khẩu mới</label><input id="cp3" v-model="f.confirm" class="input" type="password" autocomplete="new-password" :class="{ invalid: errs.confirm }" /><span v-if="errs.confirm" class="err">{{ errs.confirm }}</span></div>
      <div><button class="btn btn-primary" type="submit" :disabled="busy || !f.current || !f.next"><span v-if="busy" class="spinner" />Đổi mật khẩu</button></div>
    </form>
  </div>
</template>
