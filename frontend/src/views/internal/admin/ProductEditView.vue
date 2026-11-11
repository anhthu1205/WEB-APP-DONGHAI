<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { productsApi } from '../../../api/products'
import { useToastStore } from '../../../stores/toast'
import AppIcon from '../../../components/AppIcon.vue'
import StateBlock from '../../../components/StateBlock.vue'
import MoneyInput from '../../../components/MoneyInput.vue'
import SpecsEditor from '../../../components/SpecsEditor.vue'
import ImageUploader from '../../../components/ImageUploader.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const toast = useToastStore()
const id = computed(() => (route.params.id ? Number(route.params.id) : null))
const f = reactive({ name: '', slug: '', sku: '', short_description: '', description: '', specs: [], unit: 'cái', price_public: null, price_internal: null, is_visible: true, sort_order: 0 })
const images = ref([])
const loading = ref(!!id.value)
const loadErr = ref(null)
const errs = ref({})
const msg = ref('')
const saving = ref(false)
const deleting = ref(false)
const busyDel = ref(false)

async function load() {
  if (!id.value) return
  loading.value = true; loadErr.value = null
  try {
    const p = await productsApi.get(id.value)
    Object.assign(f, { name: p.name, slug: p.slug, sku: p.sku || '', short_description: p.short_description, description: p.description,
      specs: p.specs.map((s) => ({ ...s })), unit: p.unit, price_public: p.price_public, price_internal: p.price_internal,
      is_visible: p.is_visible, sort_order: p.sort_order })
    images.value = p.images || []
  } catch (e) { loadErr.value = e } finally { loading.value = false }
}
async function reloadImages() {
  try { images.value = (await productsApi.get(id.value)).images || [] } catch { /* bỏ qua */ }
}
onMounted(load)

async function save() {
  errs.value = {}; msg.value = ''
  saving.value = true
  // Gửi đúng tập trường của request (K7)
  const body = {
    name: f.name, slug: f.slug.trim() || null, sku: f.sku.trim() || null,
    short_description: f.short_description, description: f.description,
    specs: f.specs.filter((s) => s.label.trim() || s.value.trim()).map((s) => ({ label: s.label, value: s.value })),
    unit: f.unit, price_public: f.price_public ?? 0, price_internal: f.price_internal,
    is_visible: f.is_visible, sort_order: Number(f.sort_order) || 0,
  }
  try {
    if (id.value) {
      const p = await productsApi.update(id.value, body)
      f.slug = p.slug
      toast.success('Đã lưu sản phẩm')
    } else {
      const p = await productsApi.create(body)
      toast.success('Đã thêm sản phẩm. Bạn có thể tải ảnh lên ngay.')
      router.replace('/noi-bo/san-pham/' + p.id)
      setTimeout(load, 0)
    }
  } catch (e) {
    errs.value = e.fields || {}
    msg.value = Object.keys(errs.value).length ? 'Vui lòng kiểm tra các ô được đánh dấu.' : e.message
  } finally { saving.value = false }
}
async function remove() {
  busyDel.value = true
  try { await productsApi.remove(id.value); toast.success('Đã xóa sản phẩm'); router.replace('/noi-bo/san-pham') }
  catch (e) { toast.error(e.message) } finally { busyDel.value = false; deleting.value = false }
}
</script>

<template>
  <div>
    <RouterLink to="/noi-bo/san-pham" class="back small"><AppIcon name="left" :size="16" /> Danh sách sản phẩm</RouterLink>
    <div class="page-head">
      <div><h1>{{ id ? 'Sửa sản phẩm' : 'Thêm sản phẩm' }}</h1><p v-if="id && f.is_visible"><a :href="'/?sp=' + f.slug" target="_blank" rel="noopener">Xem trên trang chủ</a></p></div>
      <button v-if="id" class="btn btn-ghost" style="color:var(--danger-600)" @click="deleting = true"><AppIcon name="trash" :size="16" /> Xóa sản phẩm</button>
    </div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="loadErr" kind="error" :title="loadErr.status === 404 ? 'Không tìm thấy sản phẩm' : 'Chưa tải được sản phẩm'" :text="loadErr.message" @retry="load" />
    <div v-else class="cols">
      <form class="panel panel-pad form-grid cols-2" novalidate @submit.prevent="save">
        <div class="field span-2"><label for="pn">Tên sản phẩm</label><input id="pn" v-model="f.name" class="input" maxlength="200" :class="{ invalid: errs.name }" /><span v-if="errs.name" class="err">{{ errs.name }}</span></div>
        <div class="field"><label for="ps">Đường dẫn (slug)</label><input id="ps" v-model="f.slug" class="input" placeholder="Để trống để tự tạo từ tên" :class="{ invalid: errs.slug }" /><span v-if="errs.slug" class="err">{{ errs.slug }}</span></div>
        <div class="field"><label for="pk">Mã SKU</label><input id="pk" v-model="f.sku" class="input" maxlength="50" :class="{ invalid: errs.sku }" /><span v-if="errs.sku" class="err">{{ errs.sku }}</span></div>
        <div class="field"><label for="pp">Giá niêm yết (công khai)</label><MoneyInput id="pp" v-model="f.price_public" :invalid="!!errs.price_public" /><span v-if="errs.price_public" class="err">{{ errs.price_public }}</span></div>
        <div class="field"><label for="pi">Giá nội bộ</label><MoneyInput id="pi" v-model="f.price_internal" placeholder="Chưa đặt" :invalid="!!errs.price_internal" /><span class="hint">Chỉ nhân viên đã đăng nhập thấy giá này.</span></div>
        <div class="field"><label for="pu">Đơn vị</label><input id="pu" v-model="f.unit" class="input" maxlength="20" /></div>
        <div class="field"><label for="po">Thứ tự hiển thị</label><input id="po" v-model.number="f.sort_order" class="input num" type="number" /><span class="hint">Số nhỏ hiển thị trước.</span></div>
        <div class="field span-2"><label for="psd">Mô tả ngắn</label><textarea id="psd" v-model="f.short_description" class="textarea" style="min-height:70px" maxlength="500" /></div>
        <div class="field span-2"><label for="pd">Mô tả chi tiết</label><textarea id="pd" v-model="f.description" class="textarea" style="min-height:160px" maxlength="20000" /><span class="hint">Xuống dòng được giữ nguyên khi hiển thị.</span></div>
        <div class="field span-2"><span class="label">Thông số kỹ thuật</span><SpecsEditor v-model="f.specs" /></div>
        <label class="check span-2"><input v-model="f.is_visible" type="checkbox" /> Hiển thị trên trang chủ</label>
        <div v-if="msg" class="alert alert-error span-2" role="alert">{{ msg }}</div>
        <div class="span-2 row"><button class="btn btn-primary" type="submit" :disabled="saving"><span v-if="saving" class="spinner" />{{ id ? 'Lưu thay đổi' : 'Thêm sản phẩm' }}</button></div>
      </form>
      <section class="panel panel-pad">
        <h2 class="h">Ảnh sản phẩm</h2>
        <ImageUploader v-if="id" :product-id="id" :images="images" @changed="reloadImages" />
        <p v-else class="muted">Lưu sản phẩm trước, sau đó tải ảnh lên.</p>
      </section>
    </div>
    <ConfirmDialog v-if="deleting" title="Xóa sản phẩm?" message="Chỉ xóa được sản phẩm chưa có đơn hàng. Sản phẩm đã bán nên chọn Ẩn." confirm-text="Xóa sản phẩm" danger :busy="busyDel" @confirm="remove" @cancel="deleting = false" />
  </div>
</template>

<style scoped>
.back { display: inline-flex; align-items: center; gap: 4px; margin-bottom: 12px; color: var(--ink-500); }
.cols { display: grid; gap: 16px; }
@media (min-width: 1200px) { .cols { grid-template-columns: 1.4fr 1fr; align-items: start; } }
.h { font-size: 17px; margin-bottom: 14px; }
.err { color: var(--danger-600); font-size: 13px; }
</style>
