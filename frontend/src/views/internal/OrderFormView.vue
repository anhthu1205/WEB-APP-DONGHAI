<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { productsApi } from '../../api/products'
import { customersApi } from '../../api/customers'
import { ordersApi } from '../../api/orders'
import { useToastStore } from '../../stores/toast'
import { formatVND, formatDate, todayVN, addWarrantyYears } from '../../utils/format'
import { normalizePhone, prettyPhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'
import PhoneInput from '../../components/PhoneInput.vue'
import MoneyInput from '../../components/MoneyInput.vue'

const route = useRoute()
const router = useRouter()
const toast = useToastStore()
const editId = computed(() => (route.name === 'order-edit' ? Number(route.params.id) : null))

const products = ref([])
const loading = ref(true)
const loadErr = ref(null)

// Bước 1: khách hàng
const phone = ref('')
const phoneErr = ref('')
const searching = ref(false)
const customer = ref(null) // khách có sẵn
const isNew = ref(false) // tạo khách mới
const newCust = reactive({ full_name: '', address: '' })

// Bước 2: đơn
const today = todayVN()
const form = reactive({ purchase_date: today, note: '', items: [] })
const errs = ref({})
const saving = ref(false)
const formMsg = ref('')

const step2 = computed(() => !!customer.value || isNew.value)
const total = computed(() => form.items.reduce((s, it) => s + (Number(it.quantity) || 0) * (it.unit_price ?? 0), 0))
const warrantyEnd = computed(() => addWarrantyYears(form.purchase_date))
const productMap = computed(() => Object.fromEntries(products.value.map((p) => [p.id, p])))
const sortedProducts = computed(() => [...products.value].sort((a, b) => Number(b.is_visible) - Number(a.is_visible)))

function newLine() { return { product_id: '', quantity: 1, unit_price: null, serial_number: '' } }
function onProduct(it) {
  const p = productMap.value[it.product_id]
  if (p) it.unit_price = p.price_public
}

async function findCustomer() {
  phoneErr.value = ''
  if (!normalizePhone(phone.value)) { phoneErr.value = 'Số điện thoại chưa đúng'; return }
  searching.value = true
  customer.value = null; isNew.value = false
  try {
    customer.value = await customersApi.byPhone(phone.value)
  } catch (e) {
    if (e.status === 404) isNew.value = true
    else phoneErr.value = e.message
  } finally { searching.value = false }
  if (step2.value && !form.items.length) form.items.push(newLine())
}
function changeCustomer() { customer.value = null; isNew.value = false }

onMounted(async () => {
  try {
    products.value = await productsApi.list()
    if (editId.value) {
      const o = await ordersApi.get(editId.value)
      customer.value = { id: o.customer.id, full_name: o.customer.full_name, phone: o.customer.phone, address: o.customer.address }
      form.purchase_date = o.purchase_date
      form.note = o.note
      form.items = o.items.map((i) => ({ product_id: i.product_id, quantity: i.quantity, unit_price: i.unit_price, serial_number: i.serial_number || '' }))
    } else if (route.query.khach) {
      const c = await customersApi.get(route.query.khach)
      customer.value = c
      phone.value = c.phone
      form.items.push(newLine())
    }
  } catch (e) { loadErr.value = e.message } finally { loading.value = false }
})

function itemsBody() {
  return form.items.map((it) => ({
    product_id: Number(it.product_id) || 0,
    quantity: Number(it.quantity) || 0,
    unit_price: it.unit_price,
    serial_number: it.serial_number.trim() || null,
  }))
}

async function save() {
  errs.value = {}; formMsg.value = ''
  saving.value = true
  try {
    let o
    if (editId.value) {
      o = await ordersApi.update(editId.value, { customer_id: customer.value.id, purchase_date: form.purchase_date, note: form.note, items: itemsBody() })
      toast.success('Đã lưu đơn hàng')
    } else {
      const body = { purchase_date: form.purchase_date, note: form.note, items: itemsBody() }
      if (customer.value) body.customer_id = customer.value.id
      else body.customer = { full_name: newCust.full_name, phone: phone.value, address: newCust.address }
      o = await ordersApi.create(body)
      toast.success(o.customer_reused && !customer.value ? 'Đã lưu đơn (SĐT đã có, dùng khách sẵn có)' : 'Đã lưu đơn hàng ' + o.order_code)
    }
    router.replace('/noi-bo/don-hang/' + o.id)
  } catch (e) {
    errs.value = e.fields || {}
    formMsg.value = Object.keys(errs.value).length ? 'Vui lòng kiểm tra các ô được đánh dấu.' : e.message
  } finally { saving.value = false }
}
const err = (k) => errs.value[k]
</script>

<template>
  <div>
    <RouterLink :to="editId ? '/noi-bo/don-hang/' + editId : '/noi-bo/don-hang'" class="back small"><AppIcon name="left" :size="16" /> {{ editId ? 'Chi tiết đơn' : 'Danh sách đơn hàng' }}</RouterLink>
    <div class="page-head"><div><h1>{{ editId ? 'Sửa đơn hàng' : 'Nhập đơn hàng' }}</h1><p>Đơn đã lưu tra cứu bảo hành được ngay bằng SĐT khách.</p></div></div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="loadErr" kind="error" title="Chưa tải được dữ liệu" :text="loadErr" />
    <form v-else class="stack" novalidate @submit.prevent="save">
      <!-- Bước 1 -->
      <section class="panel panel-pad">
        <h2 class="step"><span class="n">1</span> Khách hàng</h2>
        <div v-if="customer" class="cust">
          <span class="avatar">{{ customer.full_name.charAt(0) }}</span>
          <div><strong>{{ customer.full_name }}</strong><p class="muted small num">{{ prettyPhone(customer.phone) }}<template v-if="customer.address"> — {{ customer.address }}</template></p></div>
          <span class="badge badge-blue">Khách đã có</span>
          <button v-if="!editId" type="button" class="btn btn-text" @click="changeCustomer">Đổi khách</button>
        </div>
        <template v-else>
          <div class="find">
            <div class="field grow">
              <label for="o-phone">Số điện thoại khách</label>
              <PhoneInput id="o-phone" v-model="phone" :invalid="!!phoneErr || !!err('customer.phone')" @keydown.enter.prevent="findCustomer" />
            </div>
            <button type="button" class="btn btn-ghost" :disabled="searching" @click="findCustomer"><span v-if="searching" class="spinner" /><AppIcon v-else name="search" /> Tìm khách</button>
          </div>
          <p v-if="phoneErr || err('customer.phone')" class="err">{{ phoneErr || err('customer.phone') }}</p>
          <div v-if="isNew" class="new">
            <p class="alert alert-info"><AppIcon name="plus" :size="18" />Chưa có khách với SĐT này — nhập thông tin để tạo khách mới cùng đơn.</p>
            <div class="form-grid cols-2">
              <div class="field"><label for="o-name">Họ tên khách</label><input id="o-name" v-model="newCust.full_name" class="input" :class="{ invalid: err('customer.full_name') }" maxlength="100" /><span v-if="err('customer.full_name')" class="err">{{ err('customer.full_name') }}</span></div>
              <div class="field"><label for="o-addr">Địa chỉ</label><input id="o-addr" v-model="newCust.address" class="input" maxlength="300" /></div>
            </div>
          </div>
        </template>
        <p v-if="err('customer') || err('customer_id')" class="err">{{ err('customer') || err('customer_id') }}</p>
      </section>

      <!-- Bước 2 -->
      <section v-if="step2" class="panel panel-pad">
        <h2 class="step"><span class="n">2</span> Sản phẩm & ngày mua</h2>
        <div class="form-grid cols-2">
          <div class="field"><label for="o-date">Ngày mua</label><input id="o-date" v-model="form.purchase_date" class="input" type="date" :max="today" min="2000-01-01" :class="{ invalid: err('purchase_date') }" /><span v-if="err('purchase_date')" class="err">{{ err('purchase_date') }}</span></div>
          <div class="field"><span class="label">Hạn bảo hành dự kiến</span><p class="warr"><AppIcon name="shield" :size="18" />{{ warrantyEnd ? formatDate(warrantyEnd) : '—' }}</p></div>
        </div>

        <div class="lines">
          <div class="line head"><span>Sản phẩm</span><span>SL</span><span>Đơn giá</span><span>Số serial</span><span /></div>
          <div v-for="(it, i) in form.items" :key="i" class="line">
            <div class="field">
              <label class="m-label" :for="'p' + i">Sản phẩm</label>
              <select :id="'p' + i" v-model="it.product_id" class="select" :class="{ invalid: err(`items.${i}.product_id`) }" @change="onProduct(it)">
                <option value="" disabled>Chọn sản phẩm</option>
                <option v-for="p in sortedProducts" :key="p.id" :value="p.id">{{ p.name }}{{ p.is_visible ? '' : ' (ngừng KD)' }}</option>
              </select>
              <span v-if="err(`items.${i}.product_id`)" class="err">{{ err(`items.${i}.product_id`) }}</span>
            </div>
            <div class="field"><label class="m-label" :for="'q' + i">Số lượng</label><input :id="'q' + i" v-model.number="it.quantity" class="input num" type="number" min="1" max="10000" :class="{ invalid: err(`items.${i}.quantity`) }" /></div>
            <div class="field"><label class="m-label" :for="'u' + i">Đơn giá</label><MoneyInput :id="'u' + i" v-model="it.unit_price" :invalid="!!err(`items.${i}.unit_price`)" /></div>
            <div class="field"><label class="m-label" :for="'s' + i">Số serial</label><input :id="'s' + i" v-model="it.serial_number" class="input" maxlength="100" placeholder="Không bắt buộc" /></div>
            <button type="button" class="btn btn-ghost btn-icon rm" :disabled="form.items.length <= 1" aria-label="Xóa dòng" @click="form.items.splice(i, 1)"><AppIcon name="trash" :size="18" /></button>
          </div>
          <p v-if="err('items')" class="err">{{ err('items') }}</p>
          <button type="button" class="btn btn-ghost btn-sm" :disabled="form.items.length >= 50" @click="form.items.push(newLine())"><AppIcon name="plus" :size="16" /> Thêm dòng sản phẩm</button>
        </div>

        <div class="field" style="margin-top:16px"><label for="o-note">Ghi chú</label><textarea id="o-note" v-model="form.note" class="textarea" maxlength="1000" placeholder="VD: lắp đặt tại nhà, giao chiều thứ 7" /></div>

        <div class="sum">
          <div><span class="muted">Tổng tiền</span><strong class="price num">{{ formatVND(total) }}</strong></div>
          <div v-if="formMsg" class="alert alert-error" role="alert"><AppIcon name="alert" :size="18" />{{ formMsg }}</div>
          <button class="btn btn-primary" type="submit" :disabled="saving"><span v-if="saving" class="spinner" />{{ editId ? 'Lưu thay đổi' : 'Lưu đơn hàng' }}</button>
        </div>
      </section>
    </form>
  </div>
</template>

<style scoped>
.back { display: inline-flex; align-items: center; gap: 4px; margin-bottom: 12px; color: var(--ink-500); }
.step { font-size: 17px; display: flex; align-items: center; gap: 10px; margin-bottom: 16px; }
.n { width: 28px; height: 28px; border-radius: 50%; background: var(--sea-600); color: #fff; display: grid; place-items: center; font-size: 14px; }
.find { display: flex; gap: 10px; align-items: flex-end; flex-wrap: wrap; }
.grow { flex: 1 1 260px; }
.err { color: var(--danger-600); font-size: 13px; margin-top: 6px; }
.new { margin-top: 16px; display: flex; flex-direction: column; gap: 14px; }
.cust { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; padding: 12px; border: 1px solid var(--sea-100); background: var(--sea-50); border-radius: 12px; }
.cust p { margin-top: 2px; }
.avatar { width: 40px; height: 40px; border-radius: 50%; background: var(--sea-600); color: #fff; display: grid; place-items: center; font-weight: 700; }
.warr { min-height: 44px; display: flex; align-items: center; gap: 8px; font-weight: 700; color: var(--ok-700); }
.lines { margin-top: 20px; display: flex; flex-direction: column; gap: 10px; align-items: stretch; }
.lines > .btn { align-self: flex-start; }
.line { display: grid; grid-template-columns: minmax(0, 2.4fr) 90px minmax(0, 1.3fr) minmax(0, 1.3fr) 44px; gap: 8px; align-items: start; }
.line.head { font-size: 13px; color: var(--ink-500); font-weight: 600; }
.m-label { display: none; }
@media (max-width: 860px) {
  .line { grid-template-columns: 1fr 1fr; padding: 12px; border: 1px solid var(--line); border-radius: 12px; }
  .line.head { display: none; }
  .line .field:first-child { grid-column: span 2; }
  .m-label { display: block; font-size: 13px; font-weight: 600; color: var(--ink-700); }
  .rm { grid-column: span 2; justify-self: end; width: auto; padding: 0 12px; }
}
.sum { margin-top: 20px; padding-top: 16px; border-top: 1px solid var(--line); display: flex; align-items: center; gap: 16px; flex-wrap: wrap; justify-content: flex-end; }
.sum > div:first-child { display: flex; align-items: baseline; gap: 10px; margin-right: auto; }
.sum strong { font-size: 24px; }
</style>
