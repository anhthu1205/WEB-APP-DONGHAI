<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { productsApi } from '../../api/products'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'
import { formatVND } from '../../utils/format'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'
import ProductPlaceholder from '../../components/ProductPlaceholder.vue'

const auth = useAuthStore()
const router = useRouter()
const toast = useToastStore()
const list = ref([])
const loading = ref(true)
const error = ref(null)
const q = ref('')
const busyId = ref(null)

const filtered = computed(() => {
  const k = q.value.trim().toLowerCase()
  return k ? list.value.filter((p) => (p.name + ' ' + (p.sku || '')).toLowerCase().includes(k)) : list.value
})
async function load() {
  loading.value = true; error.value = null
  try { list.value = await productsApi.list() } catch (e) { error.value = e.message } finally { loading.value = false }
}
async function toggle(p) {
  busyId.value = p.id
  try {
    await productsApi.setVisible(p.id, !p.is_visible)
    p.is_visible = !p.is_visible
    toast.success(p.is_visible ? 'Đã hiện sản phẩm trên trang chủ' : 'Đã ẩn sản phẩm khỏi trang chủ')
  } catch (e) { toast.error(e.message) } finally { busyId.value = null }
}
// Sao chép SP (kèm ảnh) rồi mở trang sửa bản sao
async function duplicate(p) {
  busyId.value = p.id
  try {
    const c = await productsApi.duplicate(p.id)
    toast.success('Đã tạo bản sao (đang ẩn). Sửa thông tin rồi bật Hiển thị để khách thấy.')
    router.push('/noi-bo/san-pham/' + c.id)
  } catch (e) { toast.error(e.message) } finally { busyId.value = null }
}
onMounted(load)
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Sản phẩm</h1><p>Giá niêm yết hiển thị công khai; giá nội bộ chỉ nhân viên thấy.</p></div>
      <RouterLink v-if="auth.isAdmin" to="/noi-bo/san-pham/moi" class="btn btn-primary"><AppIcon name="plus" /> Thêm sản phẩm</RouterLink>
    </div>
    <div class="panel">
      <div class="panel-head toolbar"><input v-model="q" class="input" type="search" placeholder="Tìm theo tên hoặc mã SKU" aria-label="Tìm sản phẩm" /></div>
      <StateBlock v-if="loading" kind="loading" />
      <StateBlock v-else-if="error" kind="error" title="Chưa tải được sản phẩm" :text="error" @retry="load" />
      <StateBlock v-else-if="!filtered.length" :title="list.length ? 'Không có sản phẩm khớp tìm kiếm' : 'Chưa có sản phẩm'" />
      <div v-else class="table-wrap">
        <table class="table responsive">
          <thead><tr><th>Sản phẩm</th><th class="right">Giá niêm yết</th><th class="right">Giá nội bộ</th><th>Trạng thái</th><th v-if="auth.isAdmin" class="right">Thao tác</th></tr></thead>
          <tbody>
            <tr v-for="p in filtered" :key="p.id">
              <td data-label="Sản phẩm">
                <div class="prod">
                  <span class="thumb"><img v-if="p.main_image_url" :src="p.main_image_url" alt="" loading="lazy" /><ProductPlaceholder v-else :seed="p.id" /></span>
                  <span><strong>{{ p.name }}</strong><br /><span class="muted small">{{ p.sku || p.slug }}</span></span>
                </div>
              </td>
              <td data-label="Giá niêm yết" class="right num">{{ formatVND(p.price_public) }}</td>
              <td data-label="Giá nội bộ" class="right num"><strong>{{ formatVND(p.price_internal) }}</strong></td>
              <td data-label="Trạng thái"><span class="badge" :class="p.is_visible ? 'badge-ok' : 'badge-muted'"><span class="dot" />{{ p.is_visible ? 'Đang hiển thị' : 'Đang ẩn' }}</span></td>
              <td v-if="auth.isAdmin" class="right actions">
                <div class="row" style="justify-content:flex-end">
                  <button class="btn btn-ghost btn-sm" :disabled="busyId === p.id" @click="toggle(p)">{{ p.is_visible ? 'Ẩn' : 'Hiện' }}</button>
                  <button class="btn btn-ghost btn-sm icon-only" :disabled="busyId === p.id" title="Sao chép sản phẩm" :aria-label="'Sao chép ' + p.name" @click="duplicate(p)"><AppIcon name="copy" :size="16" /></button>
                  <RouterLink :to="'/noi-bo/san-pham/' + p.id" class="btn btn-ghost btn-sm"><AppIcon name="edit" :size="16" /> Sửa</RouterLink>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.prod { display: flex; align-items: center; gap: 12px; text-align: left; }
.thumb { width: 48px; height: 48px; border-radius: 10px; overflow: hidden; flex: none; background: var(--mist); }
.thumb img { width: 100%; height: 100%; object-fit: cover; }
.icon-only { width: 36px; padding: 0; justify-content: center; }
.actions .row { flex-wrap: nowrap; gap: 6px; }
td.num { white-space: nowrap; }
</style>
