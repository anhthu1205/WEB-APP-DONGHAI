<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { customersApi } from '../../api/customers'
import { useToastStore } from '../../stores/toast'
import { formatDate } from '../../utils/format'
import { prettyPhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'
import PaginationBar from '../../components/PaginationBar.vue'
import CustomerFormModal from './CustomerFormModal.vue'

const router = useRouter()
const toast = useToastStore()
const q = ref('')
const page = ref(1)
const data = ref(null)
const loading = ref(true)
const error = ref(null)
const adding = ref(false)
let timer

async function load() {
  loading.value = true; error.value = null
  try { data.value = await customersApi.list({ q: q.value.trim(), page: page.value, page_size: 20 }) }
  catch (e) { error.value = e.message } finally { loading.value = false }
}
watch(q, () => { clearTimeout(timer); timer = setTimeout(() => { page.value = 1; load() }, 300) })
watch(page, load)
onMounted(load)
function saved(c) { adding.value = false; toast.success('Đã thêm khách hàng'); router.push('/noi-bo/khach-hang/' + c.id) }
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Khách hàng</h1><p>Tìm theo tên hoặc số điện thoại.</p></div>
      <button class="btn btn-primary" @click="adding = true"><AppIcon name="plus" /> Thêm khách hàng</button>
    </div>
    <div class="panel">
      <div class="panel-head toolbar"><input v-model="q" class="input" type="search" placeholder="Tên khách hoặc SĐT, VD: 0912 345" aria-label="Tìm khách hàng" /></div>
      <StateBlock v-if="loading && !data" kind="loading" />
      <StateBlock v-else-if="error" kind="error" title="Chưa tải được danh sách" :text="error" @retry="load" />
      <StateBlock v-else-if="!data.items.length" :title="q ? 'Không có khách hàng khớp tìm kiếm' : 'Chưa có khách hàng'" :text="q ? '' : 'Thêm khách hàng mới hoặc nhập đơn hàng kèm khách.'" />
      <template v-else>
        <div class="table-wrap" :style="{ opacity: loading ? .6 : 1 }">
          <table class="table responsive">
            <thead><tr><th>Họ tên</th><th>Số điện thoại</th><th>Địa chỉ</th><th class="right">Số đơn</th><th>Ngày tạo</th></tr></thead>
            <tbody>
              <tr v-for="c in data.items" :key="c.id">
                <td data-label="Họ tên"><RouterLink :to="'/noi-bo/khach-hang/' + c.id" class="row-link">{{ c.full_name }}</RouterLink></td>
                <td data-label="Số điện thoại" class="num">{{ prettyPhone(c.phone) }}</td>
                <td data-label="Địa chỉ" class="muted">{{ c.address || '—' }}</td>
                <td data-label="Số đơn" class="right num">{{ c.orders_count }}</td>
                <td data-label="Ngày tạo" class="num">{{ formatDate(c.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <PaginationBar :page="data.page" :page-size="data.page_size" :total="data.total" @change="page = $event" />
      </template>
    </div>
    <CustomerFormModal v-if="adding" @close="adding = false" @saved="saved" />
  </div>
</template>
