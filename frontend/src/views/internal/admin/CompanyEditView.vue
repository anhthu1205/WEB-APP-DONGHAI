<script setup>
import { reactive, ref, onMounted } from 'vue'
import { companyApi } from '../../../api/company'
import { useCompanyStore } from '../../../stores/company'
import { useToastStore } from '../../../stores/toast'
import StateBlock from '../../../components/StateBlock.vue'

const store = useCompanyStore()
const toast = useToastStore()
const f = reactive({ name: '', slogan: '', about: '', address: '', phone: '', email: '', working_hours: '', facebook_url: '', tiktok_url: '', zalo_url: '', shopee_url: '' })
const loading = ref(true)
const loadErr = ref(null)
const errs = ref({})
const msg = ref('')
const saving = ref(false)
const socials = [
  ['facebook_url', 'Facebook', 'https://facebook.com/ten-trang'],
  ['tiktok_url', 'TikTok', 'https://www.tiktok.com/@ten-kenh'],
  ['zalo_url', 'Zalo', 'https://zalo.me/0901234567'],
  ['shopee_url', 'Shopee', 'https://shopee.vn/ten-shop'],
]
async function load() {
  loading.value = true; loadErr.value = null
  const c = await store.load(true)
  if (!c) { loadErr.value = store.error?.message || 'Không tải được'; loading.value = false; return }
  Object.assign(f, { name: c.name, slogan: c.slogan, about: c.about, address: c.address, phone: c.phone, email: c.email, working_hours: c.working_hours,
    facebook_url: c.socials.facebook, tiktok_url: c.socials.tiktok, zalo_url: c.socials.zalo, shopee_url: c.socials.shopee })
  loading.value = false
}
onMounted(load)
async function save() {
  errs.value = {}; msg.value = ''; saving.value = true
  try {
    store.data = await companyApi.update({ ...f })
    toast.success('Đã lưu thông tin công ty')
  } catch (e) {
    errs.value = e.fields || {}
    msg.value = Object.keys(errs.value).length ? 'Vui lòng kiểm tra các ô được đánh dấu.' : e.message
  } finally { saving.value = false }
}
</script>

<template>
  <div>
    <div class="page-head"><div><h1>Thông tin công ty</h1><p>Hiển thị ở đầu trang chủ và chân trang.</p></div></div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="loadErr" kind="error" title="Chưa tải được dữ liệu" :text="loadErr" @retry="load" />
    <form v-else class="stack" novalidate @submit.prevent="save">
      <section class="panel panel-pad form-grid cols-2">
        <div class="field span-2"><label for="cn">Tên công ty</label><input id="cn" v-model="f.name" class="input" :class="{ invalid: errs.name }" /><span v-if="errs.name" class="err">{{ errs.name }}</span></div>
        <div class="field span-2"><label for="cs">Slogan (tiêu đề lớn trang chủ)</label><input id="cs" v-model="f.slogan" class="input" /></div>
        <div class="field span-2"><label for="ca">Giới thiệu</label><textarea id="ca" v-model="f.about" class="textarea" /></div>
        <div class="field span-2"><label for="cad">Địa chỉ</label><input id="cad" v-model="f.address" class="input" /></div>
        <div class="field"><label for="cp">Số điện thoại</label><input id="cp" v-model="f.phone" class="input" inputmode="tel" /></div>
        <div class="field"><label for="ce">Email</label><input id="ce" v-model="f.email" class="input" type="email" :class="{ invalid: errs.email }" /><span v-if="errs.email" class="err">{{ errs.email }}</span></div>
        <div class="field span-2"><label for="cw">Giờ làm việc</label><input id="cw" v-model="f.working_hours" class="input" /></div>
      </section>
      <section class="panel panel-pad form-grid cols-2">
        <div class="span-2"><h2 style="font-size:17px">Mạng xã hội</h2><p class="muted small">Đường dẫn phải bắt đầu bằng https://. Để trống thì nút tương ứng sẽ ẩn ở chân trang.</p></div>
        <div v-for="[k, label, ph] in socials" :key="k" class="field">
          <label :for="k">{{ label }}</label><input :id="k" v-model="f[k]" class="input" type="url" :placeholder="ph" :class="{ invalid: errs[k] }" /><span v-if="errs[k]" class="err">{{ errs[k] }}</span>
        </div>
      </section>
      <div v-if="msg" class="alert alert-error" role="alert">{{ msg }}</div>
      <div><button class="btn btn-primary" type="submit" :disabled="saving"><span v-if="saving" class="spinner" />Lưu thông tin</button></div>
    </form>
  </div>
</template>

<style scoped>
.err { color: var(--danger-600); font-size: 13px; }
</style>
