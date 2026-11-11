<script setup>
import { ref, computed, onMounted } from 'vue'
import { publicApi } from '../../api/public'
import { useCompanyStore } from '../../stores/company'
import { formatDate, telHref } from '../../utils/format'
import { normalizePhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import PhoneInput from '../../components/PhoneInput.vue'

const company = useCompanyStore()
const phone = ref('')
const fieldErr = ref('')
const loading = ref(false)
const error = ref(null)
const result = ref(null)
const companyPhone = computed(() => company.data?.phone || '')

async function lookup() {
  error.value = null
  if (!normalizePhone(phone.value)) {
    fieldErr.value = 'Số điện thoại chưa đúng. Ví dụ: 0912 345 678 hoặc +84 912 345 678'
    return
  }
  fieldErr.value = ''
  loading.value = true
  try {
    result.value = await publicApi.warrantyLookup(phone.value)
  } catch (e) {
    result.value = null
    if (e.status === 400 && e.fields?.phone) fieldErr.value = e.fields.phone
    else error.value = e
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const p = window.history.state?.phone
  if (p) { phone.value = p; lookup() }
})

function statusText(o) {
  if (o.status !== 'active') return 'Hết bảo hành'
  if (o.days_left === 0) return 'Còn bảo hành (hết hạn hôm nay)'
  return `Còn bảo hành · còn ${o.days_left} ngày`
}
</script>

<template>
  <div class="page">
    <div class="container narrow">
      <header class="intro">
        <span class="seal"><AppIcon name="shield" :size="26" /></span>
        <div>
          <h1>Tra cứu bảo hành</h1>
          <p>Mọi sản phẩm mua tại Đông Hải được bảo hành 2 năm kể từ ngày mua. Nhập số điện thoại đã dùng khi mua hàng để xem hạn bảo hành.</p>
        </div>
      </header>

      <form class="panel panel-pad lookup" novalidate @submit.prevent="lookup">
        <label for="w-phone" class="label">Số điện thoại mua hàng</label>
        <div class="line">
          <PhoneInput id="w-phone" v-model="phone" :invalid="!!fieldErr" autofocus />
          <button class="btn btn-primary" type="submit" :disabled="loading">
            <span v-if="loading" class="spinner" /><AppIcon v-else name="search" /> Tra cứu
          </button>
        </div>
        <p v-if="fieldErr" class="err" role="alert">{{ fieldErr }}</p>
      </form>

      <div v-if="error" class="alert" :class="error.status === 429 ? 'alert-warn' : 'alert-error'" role="alert">
        <AppIcon name="alert" :size="18" />
        <span>{{ error.status === 429 ? 'Bạn thao tác quá nhanh, vui lòng thử lại sau ít phút.' : error.message }}</span>
      </div>

      <section v-if="result" class="results" aria-live="polite">
        <div v-if="!result.orders.length" class="panel state">
          <AppIcon name="search" />
          <h3>Không tìm thấy đơn hàng</h3>
          <p>Không tìm thấy đơn hàng với số điện thoại {{ result.phone_masked }}. Vui lòng liên hệ
            <a v-if="companyPhone" :href="telHref(companyPhone)">{{ companyPhone }}</a><span v-else>cửa hàng</span> để được hỗ trợ.</p>
        </div>
        <template v-else>
          <div class="who">
            <div>
              <p class="muted small">Khách hàng</p>
              <p class="name">{{ result.customer_name_masked }} <span class="muted num">· {{ result.phone_masked }}</span></p>
            </div>
            <p class="muted small">{{ result.orders.length }} đơn hàng, tính đến {{ formatDate(result.today) }}</p>
          </div>
          <article v-for="o in result.orders" :key="o.order_code" class="order panel" :class="o.status">
            <div class="o-head">
              <div>
                <h2 class="num">Đơn {{ o.order_code }}</h2>
                <p class="muted small"><AppIcon name="calendar" :size="14" /> Ngày mua {{ formatDate(o.purchase_date) }}</p>
              </div>
              <span class="badge" :class="o.status === 'active' ? 'badge-ok' : 'badge-muted'"><span class="dot" />{{ statusText(o) }}</span>
            </div>
            <ul class="items">
              <li v-for="(it, i) in o.items" :key="i">
                <span class="pname">{{ it.product_name }}<span v-if="it.quantity > 1" class="muted"> × {{ it.quantity }}</span></span>
                <span v-if="it.serial_number" class="serial num">Serial {{ it.serial_number }}</span>
              </li>
            </ul>
            <div class="o-foot">
              <span>Hạn bảo hành</span>
              <strong class="num">{{ formatDate(o.warranty_end_date) }}</strong>
            </div>
          </article>
        </template>
      </section>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 40px 0 0; }
.narrow { max-width: 760px; }
.intro { display: flex; gap: 16px; align-items: flex-start; margin-bottom: 24px; }
.seal { flex: none; width: 52px; height: 52px; border-radius: 14px; background: var(--sea-600); color: #fff; display: grid; place-items: center; }
h1 { font-size: 28px; font-weight: 800; letter-spacing: -.02em; }
.intro p { margin-top: 6px; color: var(--ink-700); max-width: 60ch; }
.lookup { display: flex; flex-direction: column; gap: 8px; }
.line { display: flex; gap: 10px; }
.line .input { flex: 1; font-size: 17px; min-height: 50px; }
.line .btn { min-height: 50px; }
.err { color: var(--danger-600); font-size: 13px; }
.alert { margin-top: 16px; }
.results { margin-top: 24px; display: flex; flex-direction: column; gap: 14px; }
.who { display: flex; justify-content: space-between; align-items: flex-end; gap: 12px; flex-wrap: wrap; }
.who .name { font-weight: 700; font-size: 17px; }
.order { padding: 18px 20px; border-left: 4px solid var(--ok-700); }
.order.expired { border-left-color: var(--line-strong); }
.o-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; flex-wrap: wrap; }
.o-head h2 { font-size: 17px; }
.o-head p { display: flex; align-items: center; gap: 6px; margin-top: 2px; }
.items { list-style: none; padding: 0; margin: 14px 0 0; display: flex; flex-direction: column; gap: 8px; }
.items li { display: flex; justify-content: space-between; gap: 12px; flex-wrap: wrap; padding: 10px 12px; background: var(--mist); border-radius: 10px; }
.pname { font-weight: 500; }
.serial { font-size: 13px; color: var(--ink-500); }
.o-foot { margin-top: 14px; display: flex; justify-content: space-between; align-items: baseline; padding-top: 12px; border-top: 1px dashed var(--line-strong); }
.o-foot span { color: var(--ink-500); font-size: 14px; }
.o-foot strong { font-size: 18px; color: var(--ink-900); }
.order.active .o-foot strong { color: var(--ok-700); }
@media (max-width: 480px) { .line { flex-direction: column; } .line .btn { width: 100%; } h1 { font-size: 24px; } }
</style>
