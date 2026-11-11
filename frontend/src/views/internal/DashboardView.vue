<script setup>
import { ref, onMounted } from 'vue'
import { ordersApi } from '../../api/orders'
import { useAuthStore } from '../../stores/auth'
import { formatVND, formatDate } from '../../utils/format'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'

const auth = useAuthStore()
const data = ref(null)
const loading = ref(true)
const error = ref(null)
async function load() {
  loading.value = true; error.value = null
  try { data.value = await ordersApi.dashboard() } catch (e) { error.value = e.message } finally { loading.value = false }
}
onMounted(load)
</script>

<template>
  <div>
    <div class="page-head">
      <div>
        <h1>Xin chào, {{ auth.user?.full_name || auth.user?.username }}</h1>
        <p v-if="auth.isAdmin">Số liệu toàn công ty và theo từng nhân viên.</p>
        <p v-else>Số liệu khách hàng và đơn hàng do bạn nhập.</p>
      </div>
      <RouterLink to="/noi-bo/don-hang/moi" class="btn btn-primary"><AppIcon name="plus" /> Nhập đơn hàng</RouterLink>
    </div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="error" kind="error" title="Chưa tải được số liệu" :text="error" @retry="load" />
    <template v-else>
      <p v-if="data.scope === 'own'" class="scope"><AppIcon name="lock" :size="16" /> Bạn chỉ xem được dữ liệu do chính mình nhập.</p>
      <div class="stats">
        <div class="panel stat"><span class="ico"><AppIcon name="tag" /></span><div><p class="muted small">Doanh số tháng này</p><p class="v num money">{{ formatVND(data.revenue_this_month) }}</p></div></div>
        <div class="panel stat"><span class="ico"><AppIcon name="receipt" /></span><div><p class="muted small">Đơn tháng này</p><p class="v num">{{ data.orders_this_month }}</p></div></div>
        <div class="panel stat"><span class="ico"><AppIcon name="box" /></span><div><p class="muted small">Tổng số đơn</p><p class="v num">{{ data.orders_total }}</p></div></div>
        <div class="panel stat"><span class="ico"><AppIcon name="users" /></span><div><p class="muted small">Khách hàng</p><p class="v num">{{ data.customers_total }}</p></div></div>
      </div>
      <section v-if="data.staff_stats?.length" class="panel staff">
        <div class="panel-head"><h2>Theo nhân viên</h2><span class="muted small">Đơn và doanh số tính theo ngày mua trong tháng này</span></div>
        <div class="table-wrap">
          <table class="table responsive">
            <thead><tr><th>Tài khoản</th><th class="right">Khách đã tạo</th><th class="right">Đơn tháng này</th><th class="right">Doanh số tháng này</th><th class="right">Tổng số đơn</th></tr></thead>
            <tbody>
              <tr v-for="s in data.staff_stats" :key="s.user_id" :class="{ off: !s.is_active }">
                <td data-label="Tài khoản">
                  <strong>{{ s.full_name || s.username }}</strong>
                  <span class="muted small"> @{{ s.username }}</span>
                  <span v-if="s.role === 'admin'" class="badge badge-blue">Admin</span>
                  <span v-if="!s.is_active" class="badge badge-muted">Đã khóa</span>
                </td>
                <td data-label="Khách đã tạo" class="right num">{{ s.customers_total }}</td>
                <td data-label="Đơn tháng này" class="right num">{{ s.orders_this_month }}</td>
                <td data-label="Doanh số tháng này" class="right num">{{ formatVND(s.revenue_this_month) }}</td>
                <td data-label="Tổng số đơn" class="right num">{{ s.orders_total }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
      <section class="panel">
        <div class="panel-head"><h2>{{ data.scope === 'own' ? 'Đơn bạn nhập gần đây' : 'Đơn nhập gần đây' }}</h2><RouterLink to="/noi-bo/don-hang" class="small">Tất cả đơn hàng</RouterLink></div>
        <StateBlock v-if="!data.recent_orders.length" title="Chưa có đơn hàng" text="Đơn bạn nhập sẽ hiện ở đây." />
        <div v-else class="table-wrap">
          <table class="table responsive">
            <thead><tr><th>Mã đơn</th><th>Khách hàng</th><th>Ngày mua</th><th class="right">Tổng tiền</th><th>Bảo hành</th></tr></thead>
            <tbody>
              <tr v-for="o in data.recent_orders" :key="o.id">
                <td data-label="Mã đơn"><RouterLink :to="'/noi-bo/don-hang/' + o.id" class="row-link num">{{ o.order_code }}</RouterLink></td>
                <td data-label="Khách hàng">{{ o.customer.full_name }}</td>
                <td data-label="Ngày mua" class="num">{{ formatDate(o.purchase_date) }}</td>
                <td data-label="Tổng tiền" class="right num">{{ formatVND(o.total_amount) }}</td>
                <td data-label="Bảo hành"><span class="badge" :class="o.status === 'active' ? 'badge-ok' : 'badge-muted'">{{ o.status === 'active' ? 'Đến ' + formatDate(o.warranty_end_date) : 'Hết hạn' }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.stats { display: grid; gap: 14px; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); margin-bottom: 20px; }
.stat { display: flex; align-items: center; gap: 14px; padding: 18px 20px; }
.ico { width: 44px; height: 44px; border-radius: 12px; display: grid; place-items: center; background: var(--sea-50); color: var(--sea-600); }
.v { font-size: 26px; font-weight: 800; line-height: 1.15; }
.v.money { font-size: 21px; white-space: nowrap; }
.stat > div { min-width: 0; }
.scope { display: inline-flex; align-items: center; gap: 8px; margin: 0 0 14px; padding: 8px 12px; border-radius: 10px; background: var(--sea-50); color: var(--sea-800); font-size: 14px; }
.staff { margin-bottom: 20px; }
.staff .badge { margin-left: 6px; }
tr.off td { color: var(--ink-400); }
</style>
