<script setup>
import { reactive, ref, watch, onMounted } from 'vue'
import { ordersApi } from '../../api/orders'
import { formatVND, formatDate } from '../../utils/format'
import { prettyPhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'
import PaginationBar from '../../components/PaginationBar.vue'

const f = reactive({ q: '', from: '', to: '' })
const page = ref(1)
const data = ref(null)
const loading = ref(true)
const error = ref(null)
let timer
async function load() {
  loading.value = true; error.value = null
  try { data.value = await ordersApi.list({ q: f.q.trim(), from: f.from, to: f.to, page: page.value, page_size: 20 }) }
  catch (e) { error.value = e.message } finally { loading.value = false }
}
watch(() => ({ ...f }), () => { clearTimeout(timer); timer = setTimeout(() => { page.value = 1; load() }, 300) })
watch(page, load)
onMounted(load)
function reset() { f.q = ''; f.from = ''; f.to = '' }
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Đơn hàng</h1><p>Mỗi đơn được bảo hành 2 năm kể từ ngày mua.</p></div>
      <RouterLink to="/noi-bo/don-hang/moi" class="btn btn-primary"><AppIcon name="plus" /> Nhập đơn hàng</RouterLink>
    </div>
    <div class="panel">
      <div class="panel-head toolbar">
        <input v-model="f.q" class="input" type="search" placeholder="Tên hoặc SĐT khách" aria-label="Tìm theo khách" />
        <label class="date"><span class="small muted">Từ</span><input v-model="f.from" class="input" type="date" aria-label="Từ ngày" /></label>
        <label class="date"><span class="small muted">Đến</span><input v-model="f.to" class="input" type="date" aria-label="Đến ngày" /></label>
        <button v-if="f.q || f.from || f.to" class="btn btn-text" @click="reset">Xóa lọc</button>
      </div>
      <StateBlock v-if="loading && !data" kind="loading" />
      <StateBlock v-else-if="error" kind="error" title="Chưa tải được đơn hàng" :text="error" @retry="load" />
      <StateBlock v-else-if="!data.items.length" title="Không có đơn hàng phù hợp" />
      <template v-else>
        <div class="table-wrap" :style="{ opacity: loading ? .6 : 1 }">
          <table class="table responsive">
            <thead><tr><th>Mã đơn</th><th>Ngày mua</th><th>Khách hàng</th><th class="right">SL</th><th class="right">Tổng tiền</th><th>Bảo hành</th><th>Người nhập</th></tr></thead>
            <tbody>
              <tr v-for="o in data.items" :key="o.id">
                <td data-label="Mã đơn"><RouterLink :to="'/noi-bo/don-hang/' + o.id" class="row-link num">{{ o.order_code }}</RouterLink></td>
                <td data-label="Ngày mua" class="num">{{ formatDate(o.purchase_date) }}</td>
                <td data-label="Khách hàng"><RouterLink :to="'/noi-bo/khach-hang/' + o.customer.id">{{ o.customer.full_name }}</RouterLink><br /><span class="muted small num">{{ prettyPhone(o.customer.phone) }}</span></td>
                <td data-label="Số lượng" class="right num">{{ o.items_count }}</td>
                <td data-label="Tổng tiền" class="right num">{{ formatVND(o.total_amount) }}</td>
                <td data-label="Bảo hành"><span class="badge" :class="o.status === 'active' ? 'badge-ok' : 'badge-muted'">{{ o.status === 'active' ? 'Đến ' + formatDate(o.warranty_end_date) : 'Hết hạn' }}</span></td>
                <td data-label="Người nhập" class="muted">{{ o.created_by_username || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <PaginationBar :page="data.page" :page-size="data.page_size" :total="data.total" @change="page = $event" />
      </template>
    </div>
  </div>
</template>

<style scoped>
.date { display: flex; align-items: center; gap: 6px; flex: 0 1 200px; }
.date .input { flex: 1; }
</style>
