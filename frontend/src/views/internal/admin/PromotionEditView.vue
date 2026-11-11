<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { promotionsApi } from '../../../api/promotions'
import { productsApi } from '../../../api/products'
import { useToastStore } from '../../../stores/toast'
import { formatVND } from '../../../utils/format'
import AppIcon from '../../../components/AppIcon.vue'
import StateBlock from '../../../components/StateBlock.vue'
import MoneyInput from '../../../components/MoneyInput.vue'

const route = useRoute()
const router = useRouter()
const toast = useToastStore()
const id = computed(() => (route.params.id ? Number(route.params.id) : null))
const f = reactive({ title: '', content: '', start_date: '', end_date: '', is_active: true, items: [] })
const products = ref([])
const loading = ref(true)
const loadErr = ref(null)
const errs = ref({})
const msg = ref('')
const saving = ref(false)
const pmap = computed(() => Object.fromEntries(products.value.map((p) => [p.id, p])))

onMounted(async () => {
  try {
    products.value = await productsApi.list()
    if (id.value) {
      const p = await promotionsApi.get(id.value)
      Object.assign(f, { title: p.title, content: p.content, start_date: p.start_date, end_date: p.end_date, is_active: p.is_active,
        items: p.items.map((i) => ({ product_id: i.product_id, promo_price: i.promo_price, note: i.note })) })
    }
  } catch (e) { loadErr.value = e } finally { loading.value = false }
})
function addItem() { f.items.push({ product_id: '', promo_price: null, note: '' }) }
async function save() {
  errs.value = {}; msg.value = ''; saving.value = true
  const body = { title: f.title, content: f.content, start_date: f.start_date, end_date: f.end_date, is_active: f.is_active,
    items: f.items.map((i) => ({ product_id: Number(i.product_id) || 0, promo_price: i.promo_price ?? 0, note: i.note })) }
  try {
    if (id.value) await promotionsApi.update(id.value, body)
    else await promotionsApi.create(body)
    toast.success('Đã lưu chương trình khuyến mãi')
    router.push('/noi-bo/khuyen-mai')
  } catch (e) {
    errs.value = e.fields || {}
    msg.value = Object.keys(errs.value).length ? 'Vui lòng kiểm tra các ô được đánh dấu.' : e.message
  } finally { saving.value = false }
}
</script>

<template>
  <div>
    <RouterLink to="/noi-bo/khuyen-mai" class="back small"><AppIcon name="left" :size="16" /> Danh sách khuyến mãi</RouterLink>
    <div class="page-head"><div><h1>{{ id ? 'Sửa chương trình khuyến mãi' : 'Thêm chương trình khuyến mãi' }}</h1></div></div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="loadErr" kind="error" title="Chưa tải được dữ liệu" :text="loadErr.message" />
    <form v-else class="panel panel-pad form-grid cols-2" novalidate @submit.prevent="save">
      <div class="field span-2"><label for="t">Tên chương trình</label><input id="t" v-model="f.title" class="input" maxlength="200" :class="{ invalid: errs.title }" /><span v-if="errs.title" class="err">{{ errs.title }}</span></div>
      <div class="field"><label for="sd">Từ ngày</label><input id="sd" v-model="f.start_date" class="input" type="date" :class="{ invalid: errs.start_date }" /><span v-if="errs.start_date" class="err">{{ errs.start_date }}</span></div>
      <div class="field"><label for="ed">Đến ngày</label><input id="ed" v-model="f.end_date" class="input" type="date" :class="{ invalid: errs.end_date }" /><span v-if="errs.end_date" class="err">{{ errs.end_date }}</span></div>
      <div class="field span-2"><label for="ct">Nội dung</label><textarea id="ct" v-model="f.content" class="textarea" maxlength="10000" /></div>
      <label class="check span-2"><input v-model="f.is_active" type="checkbox" /> Bật chương trình</label>
      <div class="span-2">
        <p class="label">Sản phẩm áp dụng</p>
        <div v-if="!f.items.length" class="muted small" style="margin:6px 0 10px">Chưa có sản phẩm nào.</div>
        <div v-for="(it, i) in f.items" :key="i" class="item">
          <select v-model="it.product_id" class="select" :aria-label="'Sản phẩm ' + (i + 1)" :class="{ invalid: errs[`items.${i}.product_id`] }">
            <option value="" disabled>Chọn sản phẩm</option>
            <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <div><MoneyInput v-model="it.promo_price" :invalid="!!errs[`items.${i}.promo_price`]" /><span v-if="pmap[it.product_id]" class="muted small">Niêm yết {{ formatVND(pmap[it.product_id].price_public) }}</span></div>
          <input v-model="it.note" class="input" placeholder="Ghi chú (quà tặng…)" maxlength="200" :aria-label="'Ghi chú ' + (i + 1)" />
          <button type="button" class="btn btn-ghost btn-icon" aria-label="Bỏ sản phẩm" @click="f.items.splice(i, 1)"><AppIcon name="trash" :size="18" /></button>
          <span v-if="errs[`items.${i}.product_id`]" class="err full">{{ errs[`items.${i}.product_id`] }}</span>
        </div>
        <button type="button" class="btn btn-ghost btn-sm" @click="addItem"><AppIcon name="plus" :size="16" /> Thêm sản phẩm</button>
      </div>
      <div v-if="msg" class="alert alert-error span-2" role="alert">{{ msg }}</div>
      <div class="span-2"><button class="btn btn-primary" type="submit" :disabled="saving"><span v-if="saving" class="spinner" />Lưu chương trình</button></div>
    </form>
  </div>
</template>

<style scoped>
.back { display: inline-flex; align-items: center; gap: 4px; margin-bottom: 12px; color: var(--ink-500); }
.err { color: var(--danger-600); font-size: 13px; }
.item { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1.4fr) 44px; gap: 8px; margin: 8px 0; align-items: start; }
.item .full { grid-column: 1 / -1; }
@media (max-width: 760px) { .item { grid-template-columns: 1fr 44px; padding: 10px; border: 1px solid var(--line); border-radius: 12px; } .item > :not(.btn) { grid-column: 1; } .item .btn { grid-column: 2; grid-row: 1; } }
</style>
