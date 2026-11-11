<script setup>
import { ref, watch, onMounted } from 'vue'
import { promotionsApi } from '../../../api/promotions'
import { useToastStore } from '../../../stores/toast'
import { formatDate } from '../../../utils/format'
import AppIcon from '../../../components/AppIcon.vue'
import StateBlock from '../../../components/StateBlock.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'

const toast = useToastStore()
const status = ref('')
const page = ref(1)
const data = ref(null)
const loading = ref(true)
const error = ref(null)
const del = ref(null)
const busy = ref(false)
const labels = { active: ['Đang diễn ra', 'badge-ok'], upcoming: ['Sắp diễn ra', 'badge-amber'], ended: ['Đã kết thúc', 'badge-muted'] }
async function load() {
  loading.value = true; error.value = null
  try { data.value = await promotionsApi.list({ status: status.value, page: page.value, page_size: 20 }) }
  catch (e) { error.value = e.message } finally { loading.value = false }
}
watch(status, () => { page.value = 1; load() })
watch(page, load)
onMounted(load)
async function remove() {
  busy.value = true
  try { await promotionsApi.remove(del.value.id); toast.success('Đã xóa chương trình'); del.value = null; load() }
  catch (e) { toast.error(e.message) } finally { busy.value = false }
}
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Quản lý khuyến mãi</h1><p>Chương trình đang bật sẽ hiện ở trang “Khuyến mãi & bảng giá” cho nhân viên.</p></div>
      <RouterLink to="/noi-bo/khuyen-mai/moi" class="btn btn-primary"><AppIcon name="plus" /> Thêm chương trình</RouterLink>
    </div>
    <div class="panel">
      <div class="panel-head toolbar">
        <select v-model="status" class="select" aria-label="Lọc theo trạng thái" style="flex:0 1 240px">
          <option value="">Tất cả chương trình</option><option value="active">Đang diễn ra</option><option value="upcoming">Sắp diễn ra</option><option value="ended">Đã kết thúc</option>
        </select>
      </div>
      <StateBlock v-if="loading && !data" kind="loading" />
      <StateBlock v-else-if="error" kind="error" title="Chưa tải được danh sách" :text="error" @retry="load" />
      <StateBlock v-else-if="!data.items.length" title="Chưa có chương trình nào" />
      <template v-else>
        <div class="table-wrap">
          <table class="table responsive">
            <thead><tr><th>Chương trình</th><th>Thời gian</th><th class="right">Số SP</th><th>Trạng thái</th><th class="right">Thao tác</th></tr></thead>
            <tbody>
              <tr v-for="p in data.items" :key="p.id">
                <td data-label="Chương trình"><RouterLink :to="'/noi-bo/khuyen-mai/' + p.id" class="row-link">{{ p.title }}</RouterLink></td>
                <td data-label="Thời gian" class="num">{{ formatDate(p.start_date) }} – {{ formatDate(p.end_date) }}</td>
                <td data-label="Số SP" class="right num">{{ p.items_count }}</td>
                <td data-label="Trạng thái"><span v-if="!p.is_active" class="badge badge-muted">Đã tắt</span><span v-else class="badge" :class="labels[p.status][1]">{{ labels[p.status][0] }}</span></td>
                <td class="right actions"><div class="row" style="justify-content:flex-end">
                  <RouterLink :to="'/noi-bo/khuyen-mai/' + p.id" class="btn btn-ghost btn-sm"><AppIcon name="edit" :size="16" /> Sửa</RouterLink>
                  <button class="btn btn-ghost btn-sm" style="color:var(--danger-600)" aria-label="Xóa" @click="del = p"><AppIcon name="trash" :size="16" /></button>
                </div></td>
              </tr>
            </tbody>
          </table>
        </div>
        <PaginationBar :page="data.page" :page-size="data.page_size" :total="data.total" @change="page = $event" />
      </template>
    </div>
    <ConfirmDialog v-if="del" title="Xóa chương trình khuyến mãi?" :message="del.title" confirm-text="Xóa chương trình" danger :busy="busy" @confirm="remove" @cancel="del = null" />
  </div>
</template>
