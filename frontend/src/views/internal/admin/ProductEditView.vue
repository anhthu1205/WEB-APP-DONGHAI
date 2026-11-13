<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { productsApi } from '../../../api/products'
import { categoriesApi } from '../../../api/categories'
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
const f = reactive({ name: '', slug: '', sku: '', short_description: '', description: '', specs: [], unit: 'cái', price_public: null, price_internal: null, is_visible: true, sort_order: 0, category_id: null, badge: '' })
const badgeIdeas = ['Mua 5 tặng 1', 'Mới', 'Bán chạy', 'Giảm giá', 'Chính hãng']
const categories = ref([])
const images = ref([])
const loading = ref(!!id.value)
const loadErr = ref(null)
const errs = ref({})
const msg = ref('')
const saving = ref(false)
const deleting = ref(false)
const busyDel = ref(false)
const copying = ref(false)

// Sao chép SP đang xem (kèm ảnh) rồi chuyển sang trang sửa bản sao
async function duplicate() {
  copying.value = true
  try {
    const c = await productsApi.duplicate(id.value)
    toast.success('Đã tạo bản sao (đang ẩn). Sửa thông tin rồi bật Hiển thị để khách thấy.')
    router.push('/noi-bo/san-pham/' + c.id)
  } catch (e) { toast.error(e.message) } finally { copying.value = false }
}

async function load() {
  if (!id.value) return
  loading.value = true; loadErr.value = null
  try {
    const p = await productsApi.get(id.value)
    Object.assign(f, { name: p.name, slug: p.slug, sku: p.sku || '', short_description: p.short_description, description: p.description,
      specs: p.specs.map((s) => ({ ...s })), unit: p.unit, price_public: p.price_public, price_internal: p.price_internal,
      is_visible: p.is_visible, sort_order: p.sort_order, category_id: p.category_id ?? null, badge: p.badge || '' })
    images.value = p.images || []
  } catch (e) { loadErr.value = e } finally { loading.value = false }
}
async function loadCategories() {
  try { categories.value = await categoriesApi.list() } catch { /* không chặn form: chỉ thiếu danh sách danh mục */ }
}
const currentCat = computed(() => categories.value.find((c) => c.id === f.category_id) || null)
// Nơi SP xuất hiện với khách, để admin hiểu tác dụng của ô Danh mục
const placeText = computed(() => {
  const c = currentCat.value
  if (!c) return 'Sản phẩm chính: hiện ở trang chủ.'
  let t = `Hiện ở trang “${c.name}” (/danh-muc/${c.slug})`
  if (c.show_on_home) t += ' và trang chủ'
  t += '.'
  if (!c.is_visible) t += ' Danh mục này đang ẩn nên khách chưa thấy.'
  return t
})
async function reloadImages() {
  try { images.value = (await productsApi.get(id.value)).images || [] } catch { /* bỏ qua */ }
}
onMounted(() => { load(); loadCategories() })
// Cùng component được dùng lại khi chuyển sang SP khác (vd. sau khi sao chép) → tải lại
watch(id, (n, o) => { if (n && o && n !== o) load() })

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
    category_id: f.category_id || null,
    badge: f.badge.trim(),
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
      <div><h1>{{ id ? 'Sửa sản phẩm' : 'Thêm sản phẩm' }}</h1><p v-if="id && f.is_visible"><a :href="(currentCat ? '/danh-muc/' + currentCat.slug : '/') + '?sp=' + f.slug" target="_blank" rel="noopener">Xem như khách</a></p></div>
      <div v-if="id" class="row">
        <button class="btn btn-ghost" :disabled="copying" @click="duplicate"><span v-if="copying" class="spinner" /><AppIcon v-else name="copy" :size="16" /> Sao chép</button>
        <button class="btn btn-ghost" style="color:var(--danger-600)" @click="deleting = true"><AppIcon name="trash" :size="16" /> Xóa sản phẩm</button>
      </div>
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
        <div class="field span-2">
          <label for="pc">Danh mục</label>
          <select id="pc" v-model="f.category_id" class="select" :class="{ invalid: errs.category_id }">
            <option :value="null">Sản phẩm chính (trang chủ)</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}{{ c.is_visible ? '' : ' (đang ẩn)' }}</option>
          </select>
          <span v-if="errs.category_id" class="err">{{ errs.category_id }}</span>
          <span v-else class="hint">{{ placeText }} <RouterLink to="/noi-bo/danh-muc">Quản lý danh mục</RouterLink></span>
        </div>
        <div class="field span-2">
          <label for="pb">Nhãn góc thẻ</label>
          <div class="badge-row">
            <input id="pb" v-model="f.badge" class="input" maxlength="40" placeholder="Vd. Mua 5 tặng 1 — để trống nếu không cần" :class="{ invalid: errs.badge }" />
            <span v-if="f.badge.trim()" class="badge-preview" aria-hidden="true">{{ f.badge.trim() }}</span>
          </div>
          <span v-if="errs.badge" class="err">{{ errs.badge }}</span>
          <div class="ideas">
            <span class="hint">Gợi ý:</span>
            <button v-for="b in badgeIdeas" :key="b" type="button" class="chip" @click="f.badge = b">{{ b }}</button>
            <button v-if="f.badge" type="button" class="chip ghost" @click="f.badge = ''">Bỏ nhãn</button>
          </div>
        </div>
        <div class="field"><label for="pu">Đơn vị</label><input id="pu" v-model="f.unit" class="input" maxlength="20" /></div>
        <div class="field"><label for="po">Thứ tự hiển thị</label><input id="po" v-model.number="f.sort_order" class="input num" type="number" /><span class="hint">Số nhỏ hiển thị trước.</span></div>
        <div class="field span-2"><label for="psd">Mô tả ngắn</label><textarea id="psd" v-model="f.short_description" class="textarea" style="min-height:70px" maxlength="500" /></div>
        <div class="field span-2"><label for="pd">Mô tả chi tiết</label><textarea id="pd" v-model="f.description" class="textarea" style="min-height:160px" maxlength="20000" /><span class="hint">Xuống dòng được giữ nguyên khi hiển thị.</span></div>
        <div class="field span-2"><span class="label">Thông số kỹ thuật</span><SpecsEditor v-model="f.specs" /></div>
        <label class="check span-2"><input v-model="f.is_visible" type="checkbox" /> Hiển thị cho khách</label>
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
.badge-row { display: flex; gap: 10px; align-items: center; }
.badge-row .input { flex: 1; }
.badge-preview { flex-shrink: 0; max-width: 45%; padding: 6px 12px; border-radius: 999px; font-size: 13px; font-weight: 700; color: #fff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; background: linear-gradient(135deg, #ef4444, #f97316); }
.ideas { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin-top: 6px; }
.chip { border: 1px solid var(--line-strong); background: #fff; border-radius: 999px; padding: 4px 10px; font-size: 13px; cursor: pointer; color: var(--ink-700); }
.chip:hover { border-color: var(--sea-600); color: var(--sea-800); }
.chip.ghost { border-style: dashed; color: var(--ink-500); }
</style>
