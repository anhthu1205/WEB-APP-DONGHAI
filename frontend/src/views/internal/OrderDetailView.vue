<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ordersApi } from '../../api/orders'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'
import { formatVND, formatDate, formatDateTime } from '../../utils/format'
import { prettyPhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'
import ConfirmDialog from '../../components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const toast = useToastStore()
const o = ref(null)
const loading = ref(true)
const error = ref(null)
const deleting = ref(false)
const busy = ref(false)
async function load() {
  loading.value = true; error.value = null
  try { o.value = await ordersApi.get(route.params.id) } catch (e) { error.value = e } finally { loading.value = false }
}
watch(() => route.params.id, (id) => id && load())
onMounted(load)
async function remove() {
  busy.value = true
  try { await ordersApi.remove(o.value.id); toast.success('Đã xóa đơn ' + o.value.order_code); router.replace('/noi-bo/don-hang') }
  catch (e) { toast.error(e.message) } finally { busy.value = false; deleting.value = false }
}
</script>

<template>
  <div>
    <RouterLink to="/noi-bo/don-hang" class="back small"><AppIcon name="left" :size="16" /> Danh sách đơn hàng</RouterLink>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="error" kind="error" :title="error.status === 404 ? 'Không tìm thấy đơn hàng' : 'Chưa tải được đơn hàng'" :text="error.message" @retry="load" />
    <template v-else>
      <div class="page-head">
        <div><h1 class="num">Đơn {{ o.order_code }}</h1><p>Ngày mua {{ formatDate(o.purchase_date) }}</p></div>
        <div v-if="auth.isAdmin" class="row">
          <RouterLink :to="`/noi-bo/don-hang/${o.id}/sua`" class="btn btn-ghost"><AppIcon name="edit" :size="16" /> Sửa đơn</RouterLink>
          <button class="btn btn-ghost" style="color:var(--danger-600)" @click="deleting = true"><AppIcon name="trash" :size="16" /> Xóa đơn</button>
        </div>
      </div>
      <div class="cols">
        <section class="panel">
          <div class="panel-head"><h2>Sản phẩm</h2><span class="muted small">{{ o.items_count }} sản phẩm</span></div>
          <div class="table-wrap">
            <table class="table responsive">
              <thead><tr><th>Sản phẩm</th><th>Serial</th><th class="right">SL</th><th class="right">Đơn giá</th><th class="right">Thành tiền</th></tr></thead>
              <tbody>
                <tr v-for="it in o.items" :key="it.id">
                  <td data-label="Sản phẩm"><strong>{{ it.product_name }}</strong></td>
                  <td data-label="Serial" class="num">{{ it.serial_number || '—' }}</td>
                  <td data-label="Số lượng" class="right num">{{ it.quantity }}</td>
                  <td data-label="Đơn giá" class="right num">{{ formatVND(it.unit_price) }}</td>
                  <td data-label="Thành tiền" class="right num">{{ formatVND(it.line_total) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="total"><span>Tổng tiền</span><strong class="price num">{{ formatVND(o.total_amount) }}</strong></div>
        </section>
        <aside class="stack">
          <section class="panel panel-pad">
            <h2>Bảo hành</h2>
            <span class="badge" :class="o.status === 'active' ? 'badge-ok' : 'badge-muted'"><span class="dot" />{{ o.status === 'active' ? (o.days_left ? `Còn ${o.days_left} ngày` : 'Hết hạn hôm nay') : 'Hết bảo hành' }}</span>
            <p class="big num">{{ formatDate(o.warranty_end_date) }}</p>
            <p class="muted small">2 năm kể từ ngày mua</p>
          </section>
          <section class="panel panel-pad">
            <h2>Khách hàng</h2>
            <RouterLink :to="'/noi-bo/khach-hang/' + o.customer.id" class="row-link"><strong>{{ o.customer.full_name }}</strong></RouterLink>
            <p class="num">{{ prettyPhone(o.customer.phone) }}</p>
            <p v-if="o.customer.address" class="muted small">{{ o.customer.address }}</p>
          </section>
          <section class="panel panel-pad">
            <h2>Thông tin khác</h2>
            <p class="small"><span class="muted">Người nhập:</span> {{ o.created_by_username || '—' }}</p>
            <p class="small"><span class="muted">Lúc nhập:</span> {{ formatDateTime(o.created_at) }}</p>
            <p v-if="o.note" class="small pre-line" style="margin-top:8px">{{ o.note }}</p>
          </section>
        </aside>
      </div>
    </template>
    <ConfirmDialog v-if="deleting" title="Xóa đơn hàng?" :message="`Đơn ${o.order_code} sẽ bị xóa và không còn tra cứu bảo hành được.`" confirm-text="Xóa đơn" danger :busy="busy" @confirm="remove" @cancel="deleting = false" />
  </div>
</template>

<style scoped>
.back { display: inline-flex; align-items: center; gap: 4px; margin-bottom: 12px; color: var(--ink-500); }
.cols { display: grid; gap: 16px; }
@media (min-width: 1024px) { .cols { grid-template-columns: 1fr 300px; align-items: start; } }
h2 { font-size: 16px; margin-bottom: 10px; }
.panel-head h2 { margin: 0; }
.total { display: flex; justify-content: space-between; align-items: baseline; padding: 14px 16px; border-top: 1px solid var(--line); }
.total strong { font-size: 22px; }
.big { font-size: 24px; font-weight: 800; margin-top: 10px; }
</style>
