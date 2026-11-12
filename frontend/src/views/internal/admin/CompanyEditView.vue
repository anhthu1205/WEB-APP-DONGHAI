<script setup>
import { reactive, ref, onMounted } from 'vue'
import { companyApi } from '../../../api/company'
import { useCompanyStore } from '../../../stores/company'
import { useToastStore } from '../../../stores/toast'
import StateBlock from '../../../components/StateBlock.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'
import AppIcon from '../../../components/AppIcon.vue'

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
    store.set(await companyApi.update({ ...f }))
    toast.success('Đã lưu thông tin công ty')
  } catch (e) {
    errs.value = e.fields || {}
    msg.value = Object.keys(errs.value).length ? 'Vui lòng kiểm tra các ô được đánh dấu.' : e.message
  } finally { saving.value = false }
}

// ---- Logo ----
const LOGO_MAX_PX = 512
const MAX_BYTES = 2 * 1024 * 1024
const logoInput = ref(null)
const logoBusy = ref(false)
const confirmReset = ref(false)

// Thu nhỏ về cạnh dài ≤ 512px, GIỮ nền trong suốt (không tô nền trắng); xuất WebP nếu trình duyệt hỗ trợ, ngược lại PNG.
function shrinkLogo(file) {
  return new Promise((resolve, reject) => {
    // Đọc bằng data URL (CSP chỉ cho img-src 'self' data:, không cho blob:)
    const reader = new FileReader()
    const img = new Image()
    reader.onerror = () => reject(new Error('Không đọc được tệp'))
    reader.onload = () => { img.src = reader.result }
    img.onerror = () => reject(new Error('Tệp không phải ảnh hợp lệ'))
    img.onload = () => {
      const scale = Math.min(1, LOGO_MAX_PX / Math.max(img.width, img.height))
      if (scale === 1 && file.size <= MAX_BYTES) return resolve({ blob: file, name: file.name })
      const canvas = document.createElement('canvas')
      canvas.width = Math.max(1, Math.round(img.width * scale))
      canvas.height = Math.max(1, Math.round(img.height * scale))
      canvas.getContext('2d').drawImage(img, 0, 0, canvas.width, canvas.height)
      canvas.toBlob((b) => {
        if (b && b.type === 'image/webp') return resolve({ blob: b, name: 'logo.webp' })
        canvas.toBlob((p) => (p ? resolve({ blob: p, name: 'logo.png' }) : reject(new Error('Không xử lý được ảnh'))), 'image/png')
      }, 'image/webp', 0.92)
    }
    reader.readAsDataURL(file)
  })
}

async function onLogo(e) {
  const file = e.target.files?.[0]
  e.target.value = ''
  if (!file) return
  if (!/^image\/(jpeg|png|webp)$/.test(file.type)) { toast.error('Logo chỉ nhận ảnh JPEG, PNG hoặc WebP'); return }
  logoBusy.value = true
  try {
    const { blob, name } = await shrinkLogo(file)
    if (blob.size > MAX_BYTES) throw new Error('Logo vượt quá 2 MB')
    store.set(await companyApi.uploadLogo(blob, name))
    toast.success('Đã đổi logo')
  } catch (err) {
    toast.error(err.message || 'Không tải được logo')
  } finally { logoBusy.value = false }
}

async function resetLogo() {
  confirmReset.value = false
  logoBusy.value = true
  try {
    store.set(await companyApi.deleteLogo())
    toast.success('Đã chuyển về logo mặc định')
  } catch (err) {
    toast.error(err.message || 'Không xóa được logo')
  } finally { logoBusy.value = false }
}
</script>

<template>
  <div>
    <div class="page-head"><div><h1>Thông tin công ty</h1><p>Hiển thị ở đầu trang chủ và chân trang.</p></div></div>
    <section v-if="!loading && !loadErr" class="panel panel-pad logo-panel">
      <div class="logo-preview" :class="{ busy: logoBusy }">
        <img :src="store.logoUrl" alt="Logo hiện tại" />
      </div>
      <div class="logo-info">
        <h2>Logo</h2>
        <p class="muted small">Hiện ở đầu trang, chân trang, khu nội bộ và biểu tượng trên tab trình duyệt. Nên dùng ảnh PNG nền trong suốt, hình vuông hoặc ngang, tối thiểu 256px. Nhận JPEG, PNG, WebP; ảnh lớn được tự thu nhỏ.</p>
        <p class="small"><span class="badge" :class="store.data?.logo_url ? 'badge-ok' : 'badge-muted'">{{ store.data?.logo_url ? 'Đang dùng logo riêng' : 'Đang dùng logo mặc định' }}</span></p>
        <div class="row">
          <input ref="logoInput" type="file" accept="image/png,image/jpeg,image/webp" class="sr-only" @change="onLogo" />
          <button type="button" class="btn btn-primary" :disabled="logoBusy" @click="logoInput.click()">
            <span v-if="logoBusy" class="spinner" /><AppIcon v-else name="upload" /> {{ store.data?.logo_url ? 'Đổi logo' : 'Tải logo lên' }}
          </button>
          <button v-if="store.data?.logo_url" type="button" class="btn btn-ghost" :disabled="logoBusy" @click="confirmReset = true">Dùng logo mặc định</button>
        </div>
      </div>
    </section>
    <ConfirmDialog v-if="confirmReset" title="Dùng lại logo mặc định?" message="Logo đã tải lên sẽ bị xóa khỏi hệ thống." confirm-text="Dùng logo mặc định" danger @confirm="resetLogo" @cancel="confirmReset = false" />
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
.logo-panel { display: flex; gap: 20px; align-items: center; margin-bottom: 16px; flex-wrap: wrap; }
.logo-preview { width: 132px; height: 132px; border-radius: 14px; display: grid; place-items: center; padding: 12px;
  background: repeating-conic-gradient(#eef1f5 0 25%, #fff 0 50%) 0 0 / 16px 16px; border: 1px solid var(--line); flex-shrink: 0; }
.logo-preview img { max-width: 100%; max-height: 100%; object-fit: contain; }
.logo-preview.busy { opacity: .5; }
.logo-info { flex: 1; min-width: 240px; }
.logo-info h2 { font-size: 17px; margin-bottom: 4px; }
.logo-info .row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 10px; }
</style>
