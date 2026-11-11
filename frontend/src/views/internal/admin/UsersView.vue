<script setup>
import { ref, reactive, onMounted } from 'vue'
import { usersApi } from '../../../api/users'
import { useAuthStore } from '../../../stores/auth'
import { useToastStore } from '../../../stores/toast'
import { formatDateTime } from '../../../utils/format'
import AppIcon from '../../../components/AppIcon.vue'
import StateBlock from '../../../components/StateBlock.vue'
import ModalShell from '../../../components/ModalShell.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'

const auth = useAuthStore()
const toast = useToastStore()
const list = ref([])
const loading = ref(true)
const error = ref(null)
const creating = ref(false)
const nu = reactive({ username: '', full_name: '', password: '' })
const resetting = ref(null)
const newPw = ref('')
const toggling = ref(null)
const errs = ref({})
const msg = ref('')
const busy = ref(false)

async function load() {
  loading.value = true; error.value = null
  try { list.value = await usersApi.list() } catch (e) { error.value = e.message } finally { loading.value = false }
}
onMounted(load)
function openCreate() { Object.assign(nu, { username: '', full_name: '', password: '' }); errs.value = {}; msg.value = ''; creating.value = true }
async function create() {
  busy.value = true; errs.value = {}; msg.value = ''
  try { await usersApi.create({ ...nu }); toast.success('Đã tạo tài khoản ' + nu.username.toLowerCase()); creating.value = false; load() }
  catch (e) { errs.value = e.fields || {}; msg.value = Object.keys(errs.value).length ? '' : e.message } finally { busy.value = false }
}
function openReset(u) { resetting.value = u; newPw.value = ''; errs.value = {}; msg.value = '' }
async function reset() {
  busy.value = true; errs.value = {}; msg.value = ''
  try { await usersApi.resetPassword(resetting.value.id, newPw.value); toast.success('Đã đặt lại mật khẩu cho ' + resetting.value.username); resetting.value = null }
  catch (e) { errs.value = e.fields || {}; msg.value = Object.keys(errs.value).length ? '' : e.message } finally { busy.value = false }
}
async function toggle() {
  const u = toggling.value
  busy.value = true
  try {
    await usersApi.update(u.id, { is_active: !u.is_active })
    toast.success(u.is_active ? 'Đã khóa tài khoản ' + u.username : 'Đã mở khóa tài khoản ' + u.username)
    toggling.value = null
    load()
  } catch (e) { toast.error(e.message) } finally { busy.value = false }
}
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Tài khoản</h1><p>Tài khoản nhân viên được xem và nhập khách hàng, đơn hàng.</p></div>
      <button class="btn btn-primary" @click="openCreate"><AppIcon name="plus" /> Tạo tài khoản</button>
    </div>
    <div class="panel">
      <StateBlock v-if="loading" kind="loading" />
      <StateBlock v-else-if="error" kind="error" title="Chưa tải được danh sách" :text="error" @retry="load" />
      <div v-else class="table-wrap">
        <table class="table responsive">
          <thead><tr><th>Tên đăng nhập</th><th>Họ tên</th><th>Vai trò</th><th>Trạng thái</th><th>Đăng nhập gần nhất</th><th class="right">Thao tác</th></tr></thead>
          <tbody>
            <tr v-for="u in list" :key="u.id">
              <td data-label="Tên đăng nhập"><strong>{{ u.username }}</strong></td>
              <td data-label="Họ tên">{{ u.full_name || '—' }}</td>
              <td data-label="Vai trò"><span class="badge" :class="u.role === 'admin' ? 'badge-blue' : 'badge-muted'">{{ u.role === 'admin' ? 'Quản trị' : 'Nhân viên' }}</span></td>
              <td data-label="Trạng thái"><span class="badge" :class="u.is_active ? 'badge-ok' : 'badge-red'"><span class="dot" />{{ u.is_active ? 'Hoạt động' : 'Đã khóa' }}</span></td>
              <td data-label="Đăng nhập gần nhất" class="num muted">{{ u.last_login_at ? formatDateTime(u.last_login_at) : 'Chưa đăng nhập' }}</td>
              <td class="right actions">
                <div v-if="u.role !== 'admin' && u.id !== auth.user?.id" class="row" style="justify-content:flex-end">
                  <button class="btn btn-ghost btn-sm" @click="openReset(u)"><AppIcon name="key" :size="16" /> Đặt lại MK</button>
                  <button class="btn btn-ghost btn-sm" @click="toggling = u"><AppIcon :name="u.is_active ? 'lock' : 'unlock'" :size="16" /> {{ u.is_active ? 'Khóa' : 'Mở khóa' }}</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <ModalShell v-if="creating" title="Tạo tài khoản nhân viên" :busy="busy" @close="creating = false">
      <form id="nu-form" class="form-grid" novalidate @submit.prevent="create">
        <div v-if="msg" class="alert alert-error">{{ msg }}</div>
        <div class="field"><label for="nu-u">Tên đăng nhập</label><input id="nu-u" v-model="nu.username" class="input" autocapitalize="none" :class="{ invalid: errs.username }" /><span class="hint">3–32 ký tự: chữ thường không dấu, số, dấu _ hoặc .</span><span v-if="errs.username" class="err">{{ errs.username }}</span></div>
        <div class="field"><label for="nu-n">Họ tên</label><input id="nu-n" v-model="nu.full_name" class="input" maxlength="100" /></div>
        <div class="field"><label for="nu-p">Mật khẩu ban đầu</label><input id="nu-p" v-model="nu.password" class="input" type="text" autocomplete="off" :class="{ invalid: errs.password }" /><span v-if="errs.password" class="err">{{ errs.password }}</span></div>
      </form>
      <template #footer><button class="btn btn-ghost" @click="creating = false">Hủy</button><button class="btn btn-primary" form="nu-form" type="submit" :disabled="busy"><span v-if="busy" class="spinner" />Tạo tài khoản</button></template>
    </ModalShell>

    <ModalShell v-if="resetting" :title="'Đặt lại mật khẩu: ' + resetting.username" :busy="busy" @close="resetting = null">
      <form id="rs-form" class="form-grid" novalidate @submit.prevent="reset">
        <div v-if="msg" class="alert alert-error">{{ msg }}</div>
        <div class="field"><label for="rs-p">Mật khẩu mới</label><input id="rs-p" v-model="newPw" class="input" type="text" autocomplete="off" :class="{ invalid: errs.new_password }" /><span v-if="errs.new_password" class="err">{{ errs.new_password }}</span><span class="hint">Nhân viên sẽ bị đăng xuất khỏi mọi thiết bị.</span></div>
      </form>
      <template #footer><button class="btn btn-ghost" @click="resetting = null">Hủy</button><button class="btn btn-primary" form="rs-form" type="submit" :disabled="busy"><span v-if="busy" class="spinner" />Đặt lại mật khẩu</button></template>
    </ModalShell>

    <ConfirmDialog v-if="toggling" :title="toggling.is_active ? 'Khóa tài khoản?' : 'Mở khóa tài khoản?'" :message="toggling.is_active ? `${toggling.username} sẽ bị đăng xuất ngay và không đăng nhập được nữa.` : `${toggling.username} sẽ đăng nhập lại được.`" :confirm-text="toggling.is_active ? 'Khóa tài khoản' : 'Mở khóa'" :danger="toggling.is_active" :busy="busy" @confirm="toggle" @cancel="toggling = null" />
  </div>
</template>

<style scoped>
.err { color: var(--danger-600); font-size: 13px; }
</style>
