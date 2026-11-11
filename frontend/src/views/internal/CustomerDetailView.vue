<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { customersApi } from '../../api/customers'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'
import { formatVND, formatDate } from '../../utils/format'
import { prettyPhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'
import ConfirmDialog from '../../components/ConfirmDialog.vue'
import CustomerFormModal from './CustomerFormModal.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const toast = useToastStore()
const c = ref(null)
const loading = ref(true)
const error = ref(null)
const editing = ref(false)
const deleting = ref(false)
const busy = ref(false)

async function load() {
  loading.value = true; error.value = null
  try { c.value = await customersApi.get(route.params.id) } catch (e) { error.value = e } finally { loading.value = false }
}
watch(() => route.params.id, (id) => id && load())
onMounted(load)
function saved() { editing.value = false; toast.success('Đã lưu thông tin khách hàng'); load() }
async function remove() {
  busy.value = true
  try { await customersApi.remove(c.value.id); toast.success('Đã xóa khách hàng'); router.replace('/noi-bo/khach-hang') }
  catch (e) { toast.error(e.message); deleting.value = false } finally { busy.value = false }
}
</script>

<template>
  <div>
    <RouterLink to="/noi-bo/khach-hang" class="back small"><AppIcon name="left" :size="16" /> Danh sách khách hàng</RouterLink>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="error" kind="error" :title="error.status === 404 ? 'Không tìm thấy khách hàng' : 'Chưa tải được dữ liệu'" :text="error.message" @retry="load" />
    <template v-else>
      <div class="page-head">
        <div><h1>{{ c.full_name }}</h1><p class="num">{{ prettyPhone(c.phone) }}</p></div>
        <div class="row">
          <RouterLink :to="{ path: '/noi-bo/don-hang/moi', query: { khach: c.id } }" class="btn btn-primary"><AppIcon name="plus" /> Nhập đơn cho khách</RouterLink>
          <template v-if="auth.isAdmin">
            <button class="btn btn-ghost" @click="editing = true"><AppIcon name="edit" :size="16" /> Sửa</button>
            <button class="btn btn-ghost" style="color:var(--danger-600)" @click="deleting = true"><AppIcon name="trash" :size="16" /> Xóa</button>
          </template>
        </div>
      </div>
      <div class="cols">
        <section class="panel panel-pad info">
          <h2>Thông tin</h2>
          <dl>
            <dt>Địa chỉ</dt><dd>{{ c.address || '—' }}</dd>
            <dt>Ghi chú</dt><dd class="pre-line">{{ c.note || '—' }}</dd>
            <dt>Ngày tạo</dt><dd>{{ formatDate(c.created_at) }}</dd>
            <dt>Số đơn</dt><dd>{{ c.orders_count }}</dd>
          </dl>
        </section>
        <section class="panel">
          <div class="panel-head"><h2>Lịch sử mua hàng</h2></div>
          <StateBlock v-if="!c.orders.length" title="Khách chưa có đơn hàng" />
          <div v-else class="table-wrap">
            <table class="table responsive">
              <thead><tr><th>Mã đơn</th><th>Ngày mua</th><th class="right">Tổng tiền</th><th>Bảo hành</th></tr></thead>
              <tbody>
                <tr v-for="o in c.orders" :key="o.id">
                  <td data-label="Mã đơn"><RouterLink :to="'/noi-bo/don-hang/' + o.id" class="row-link num">{{ o.order_code }}</RouterLink><br /><span class="muted small">{{ o.items.map((i) => i.product_name).join(', ') }}</span></td>
                  <td data-label="Ngày mua" class="num">{{ formatDate(o.purchase_date) }}</td>
                  <td data-label="Tổng tiền" class="right num">{{ formatVND(o.total_amount) }}</td>
                  <td data-label="Bảo hành"><span class="badge" :class="o.status === 'active' ? 'badge-ok' : 'badge-muted'">{{ o.status === 'active' ? 'Còn đến ' + formatDate(o.warranty_end_date) : 'Hết hạn ' + formatDate(o.warranty_end_date) }}</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </template>
    <CustomerFormModal v-if="editing" :customer="c" @close="editing = false" @saved="saved" />
    <ConfirmDialog v-if="deleting" title="Xóa khách hàng?" :message="`Xóa ${c.full_name}. Chỉ xóa được khách chưa có đơn hàng.`" confirm-text="Xóa khách hàng" danger :busy="busy" @confirm="remove" @cancel="deleting = false" />
  </div>
</template>

<style scoped>
.back { display: inline-flex; align-items: center; gap: 4px; margin-bottom: 12px; color: var(--ink-500); }
.cols { display: grid; gap: 16px; }
@media (min-width: 1024px) { .cols { grid-template-columns: 320px 1fr; align-items: start; } }
.info h2 { font-size: 17px; margin-bottom: 12px; }
dl { margin: 0; display: grid; grid-template-columns: 90px 1fr; gap: 10px 12px; font-size: 14px; }
dt { color: var(--ink-500); }
dd { margin: 0; }
</style>
