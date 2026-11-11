<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { publicApi } from '../api/public'
import { useCompanyStore } from '../stores/company'
import { formatVND, telHref } from '../utils/format'
import AppIcon from './AppIcon.vue'
import ImageGallery from './ImageGallery.vue'

const props = defineProps({ slug: { type: String, required: true }, preview: { type: Object, default: null } })
const emit = defineEmits(['close'])
const company = useCompanyStore()
const product = ref(null)
const loading = ref(true)
const error = ref(null)
const dialog = ref(null)
const closeBtn = ref(null)
let lastFocus = null

const view = computed(() => product.value || props.preview)
const zalo = computed(() => company.data?.socials?.zalo || '')
const phone = computed(() => company.data?.phone || '')

async function load() {
  loading.value = true
  error.value = null
  try {
    product.value = await publicApi.product(props.slug)
  } catch (e) {
    error.value = e.status === 404 ? 'Sản phẩm này không còn được bán hoặc đường dẫn không đúng.' : e.message
  } finally {
    loading.value = false
  }
}
watch(() => props.slug, load)

function onKey(e) {
  if (e.key === 'Escape') emit('close')
  if (e.key === 'Tab' && dialog.value) {
    const f = dialog.value.querySelectorAll('a[href],button:not([disabled]),input,[tabindex]:not([tabindex="-1"])')
    if (!f.length) return
    const first = f[0], last = f[f.length - 1]
    if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus() }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus() }
  }
}
onMounted(async () => {
  lastFocus = document.activeElement
  document.body.style.overflow = 'hidden'
  document.addEventListener('keydown', onKey)
  load()
  await nextTick()
  closeBtn.value?.focus()
})
onBeforeUnmount(() => {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', onKey)
  lastFocus?.focus?.()
})
</script>

<template>
  <div class="overlay" @click.self="emit('close')">
    <div ref="dialog" class="modal" role="dialog" aria-modal="true" :aria-label="view?.name || 'Chi tiết sản phẩm'">
      <button ref="closeBtn" class="close" aria-label="Đóng" @click="emit('close')"><AppIcon name="close" :size="22" /></button>

      <div v-if="error" class="state">
        <AppIcon name="alert" />
        <h3>Không mở được sản phẩm</h3>
        <p>{{ error }}</p>
        <button class="btn btn-ghost btn-sm" @click="emit('close')">Quay lại danh sách</button>
      </div>

      <div v-else class="grid">
        <div class="left">
          <div v-if="loading && !product" class="skeleton" style="aspect-ratio:1/1" />
          <ImageGallery v-else :images="product?.images || []" :alt="view?.name" :seed="view?.id || 0" />
        </div>
        <div class="right">
          <h2 class="title">{{ view?.name }}</h2>
          <p v-if="view?.short_description" class="lead">{{ view.short_description }}</p>
          <div class="price-row">
            <span class="price">{{ formatVND(view?.price) }}</span>
            <span class="muted">/ {{ view?.unit }}</span>
          </div>
          <p class="warranty"><AppIcon name="shield" :size="18" /> Bảo hành chính hãng 2 năm kể từ ngày mua</p>
          <div class="cta">
            <a v-if="phone" :href="telHref(phone)" class="btn btn-accent"><AppIcon name="phone" /> Liên hệ mua: {{ phone }}</a>
            <a v-if="zalo" :href="zalo" target="_blank" rel="noopener noreferrer" class="btn btn-ghost"><AppIcon name="chat" /> Chat Zalo</a>
          </div>

          <template v-if="loading && !product">
            <div class="skeleton" style="height:14px;width:90%;margin-top:22px" />
            <div class="skeleton" style="height:14px;width:75%;margin-top:10px" />
            <div class="skeleton" style="height:14px;width:82%;margin-top:10px" />
          </template>
          <template v-else-if="product">
            <section v-if="product.description" class="sec">
              <h3>Mô tả sản phẩm</h3>
              <p class="pre-line desc">{{ product.description }}</p>
            </section>
            <section v-if="product.specs?.length" class="sec">
              <h3>Thông số kỹ thuật</h3>
              <table class="specs">
                <tbody>
                  <tr v-for="(s, i) in product.specs" :key="i"><th scope="row">{{ s.label }}</th><td>{{ s.value }}</td></tr>
                </tbody>
              </table>
            </section>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay { position: fixed; inset: 0; z-index: 100; background: rgba(7, 26, 52, .6); display: flex; align-items: center; justify-content: center; padding: 24px; animation: fade .15s ease; }
.modal { position: relative; width: 100%; max-width: 1040px; max-height: calc(100vh - 48px); overflow-y: auto; background: #fff; border-radius: var(--r-xl); box-shadow: var(--shadow-3); animation: rise .2s ease; }
@keyframes fade { from { opacity: 0; } }
@keyframes rise { from { opacity: 0; transform: translateY(12px); } }
.close { position: absolute; top: 12px; right: 12px; z-index: 2; width: 44px; height: 44px; border-radius: 50%; border: 0; background: var(--mist); color: var(--ink-700); display: grid; place-items: center; cursor: pointer; }
.close:hover { background: var(--line); }
.grid { display: grid; grid-template-columns: 1fr; gap: 24px; padding: 24px; }
@media (min-width: 860px) { .grid { grid-template-columns: 1fr 1fr; gap: 36px; padding: 32px; } .left { position: sticky; top: 0; align-self: start; } }
.title { font-size: 24px; font-weight: 700; padding-right: 48px; }
.lead { margin-top: 8px; color: var(--ink-700); }
.price-row { margin-top: 16px; display: flex; align-items: baseline; gap: 6px; }
.price-row .price { font-size: 30px; font-weight: 800; letter-spacing: -.01em; }
.warranty { margin-top: 10px; display: flex; align-items: center; gap: 8px; color: var(--ok-700); font-weight: 600; font-size: 14px; }
.cta { margin-top: 18px; display: flex; flex-wrap: wrap; gap: 10px; }
.sec { margin-top: 26px; }
.sec h3 { font-size: 16px; margin-bottom: 10px; }
.desc { color: var(--ink-700); line-height: 1.75; }
.specs { width: 100%; border-collapse: collapse; font-size: 14px; border: 1px solid var(--line); border-radius: 10px; overflow: hidden; }
.specs th, .specs td { padding: 10px 12px; border-bottom: 1px solid var(--line); text-align: left; vertical-align: top; }
.specs tr:last-child th, .specs tr:last-child td { border-bottom: 0; }
.specs th { width: 42%; background: #f8fafc; color: var(--ink-500); font-weight: 500; }
@media (max-width: 640px) {
  .overlay { padding: 0; }
  .modal { max-height: 100vh; height: 100%; border-radius: 0; }
  .grid { padding: 16px; padding-top: 64px; }
  .cta .btn { flex: 1 1 100%; }
}
</style>
