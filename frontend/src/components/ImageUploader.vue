<script setup>
import { ref } from 'vue'
import { productsApi } from '../api/products'
import { useToastStore } from '../stores/toast'
import AppIcon from './AppIcon.vue'
import ConfirmDialog from './ConfirmDialog.vue'

const props = defineProps({ productId: { type: Number, required: true }, images: { type: Array, default: () => [] } })
const emit = defineEmits(['changed'])
const toast = useToastStore()
const busy = ref(false)
const progress = ref('')
const input = ref(null)
const confirmDel = ref(null)
const MAX = 8
const MAX_BYTES = 2 * 1024 * 1024

// Thu nhỏ về cạnh dài ≤ 1600px rồi nén (webp nếu trình duyệt hỗ trợ, ngược lại jpeg 0.85)
function resize(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = () => reject(new Error('Không đọc được tệp'))
    reader.onload = () => {
      const img = new Image()
      img.onerror = () => reject(new Error('Tệp không phải ảnh hợp lệ'))
      img.onload = () => {
        const scale = Math.min(1, 1600 / Math.max(img.width, img.height))
        const w = Math.round(img.width * scale)
        const h = Math.round(img.height * scale)
        const canvas = document.createElement('canvas')
        canvas.width = w
        canvas.height = h
        const ctx = canvas.getContext('2d')
        ctx.fillStyle = '#fff'
        ctx.fillRect(0, 0, w, h)
        ctx.drawImage(img, 0, 0, w, h)
        canvas.toBlob((b) => {
          if (b && b.type === 'image/webp') return resolve({ blob: b, name: 'anh.webp' })
          canvas.toBlob((j) => (j ? resolve({ blob: j, name: 'anh.jpg' }) : reject(new Error('Không nén được ảnh'))), 'image/jpeg', 0.85)
        }, 'image/webp', 0.85)
      }
      img.src = reader.result
    }
    reader.readAsDataURL(file)
  })
}

async function onFiles(e) {
  const files = Array.from(e.target.files || [])
  e.target.value = ''
  if (!files.length) return
  const room = MAX - props.images.length
  if (room <= 0) { toast.error('Mỗi sản phẩm tối đa 8 ảnh'); return }
  busy.value = true
  let ok = 0
  try {
    for (const [i, f] of files.slice(0, room).entries()) {
      progress.value = `Đang tải ảnh ${i + 1}/${Math.min(files.length, room)}…`
      if (!/^image\/(jpeg|png|webp)$/.test(f.type)) { toast.error(`${f.name}: chỉ nhận JPEG, PNG hoặc WebP`); continue }
      let payload = { blob: f, name: f.name }
      try { payload = await resize(f) } catch { /* giữ tệp gốc */ }
      if (payload.blob.size > MAX_BYTES) { toast.error(`${f.name}: ảnh vượt quá 2 MB`); continue }
      try {
        await productsApi.uploadImage(props.productId, payload.blob, payload.name)
        ok++
      } catch (err) {
        toast.error(`${f.name}: ${err.message}`)
      }
    }
    if (files.length > room) toast.error(`Chỉ tải thêm được ${room} ảnh (tối đa 8)`)
    if (ok) toast.success(`Đã tải lên ${ok} ảnh`)
  } finally {
    busy.value = false
    progress.value = ''
    emit('changed')
  }
}

async function setMain(im) {
  busy.value = true
  try { await productsApi.setMainImage(im.id); toast.success('Đã đặt ảnh chính'); emit('changed') }
  catch (e) { toast.error(e.message) }
  finally { busy.value = false }
}
async function doDelete() {
  const im = confirmDel.value
  busy.value = true
  try { await productsApi.removeImage(im.id); toast.success('Đã xóa ảnh'); emit('changed') }
  catch (e) { toast.error(e.message) }
  finally { busy.value = false; confirmDel.value = null }
}
</script>

<template>
  <div class="uploader">
    <div class="grid">
      <figure v-for="im in images" :key="im.id" class="img" :class="{ main: im.is_main }">
        <img :src="im.url" alt="" loading="lazy" />
        <span v-if="im.is_main" class="badge badge-amber tag">Ảnh chính</span>
        <figcaption>
          <button v-if="!im.is_main" type="button" class="btn btn-sm btn-ghost" :disabled="busy" @click="setMain(im)"><AppIcon name="star" :size="15" /> Đặt chính</button>
          <button type="button" class="btn btn-sm btn-ghost danger" :disabled="busy" aria-label="Xóa ảnh" @click="confirmDel = im"><AppIcon name="trash" :size="15" /></button>
        </figcaption>
      </figure>
      <button v-if="images.length < MAX" type="button" class="add" :disabled="busy" @click="input.click()">
        <span v-if="busy" class="spinner" /><AppIcon v-else name="upload" :size="26" />
        <span>{{ busy ? progress || 'Đang xử lý…' : 'Tải ảnh lên' }}</span>
        <small>JPEG, PNG, WebP · tối đa 2 MB</small>
      </button>
    </div>
    <p class="muted small">{{ images.length }}/{{ MAX }} ảnh. Ảnh lớn được tự thu nhỏ về 1600px trước khi tải lên.</p>
    <input ref="input" type="file" accept="image/jpeg,image/png,image/webp" multiple hidden @change="onFiles" />
    <ConfirmDialog v-if="confirmDel" title="Xóa ảnh này?" message="Ảnh sẽ bị xóa vĩnh viễn khỏi sản phẩm." confirm-text="Xóa ảnh" danger :busy="busy" @confirm="doDelete" @cancel="confirmDel = null" />
  </div>
</template>

<style scoped>
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; margin-bottom: 8px; }
.img { position: relative; margin: 0; border: 1px solid var(--line); border-radius: 12px; overflow: hidden; background: #fff; }
.img.main { border-color: var(--amber-500); box-shadow: 0 0 0 2px rgba(245,158,11,.25); }
.img img { width: 100%; aspect-ratio: 1/1; object-fit: cover; }
.tag { position: absolute; top: 8px; left: 8px; }
figcaption { display: flex; gap: 6px; padding: 8px; justify-content: space-between; }
figcaption .btn { flex: 1; padding: 0 8px; }
figcaption .danger { flex: 0 0 36px; color: var(--danger-600); }
.add { aspect-ratio: 1 / 1.24; border: 2px dashed var(--line-strong); border-radius: 12px; background: var(--mist); color: var(--ink-500); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; cursor: pointer; font-weight: 600; padding: 12px; text-align: center; }
.add:hover:not(:disabled) { border-color: var(--sea-600); color: var(--sea-600); background: var(--sea-50); }
.add small { font-weight: 400; font-size: 12px; }
</style>
