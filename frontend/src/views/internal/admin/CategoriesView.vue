<script setup>
import { ref, reactive, onMounted } from 'vue'
import { categoriesApi } from '../../../api/categories'
import { useCategoriesStore } from '../../../stores/categories'
import { useToastStore } from '../../../stores/toast'
import AppIcon from '../../../components/AppIcon.vue'
import StateBlock from '../../../components/StateBlock.vue'
import ModalShell from '../../../components/ModalShell.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'

// Quản lý danh mục (ngách) sản phẩm — mỗi danh mục có trang riêng /danh-muc/{slug}.
const toast = useToastStore()
const publicCats = useCategoriesStore()
const list = ref([])
const loading = ref(true)
const error = ref(null)
const editing = ref(null) // null | 'new' | category
const f = reactive({ name: '', slug: '', description: '', is_visible: true, show_in_menu: true, show_on_home: false, sort_order: 0 })
const errs = ref({})
const msg = ref('')
const saving = ref(false)
const coverBusy = ref(false)
const coverInput = ref(null)
const delTarget = ref(null)
const deleting = ref(false)

async function load() {
  loading.value = true; error.value = null
  try { list.value = await categoriesApi.list() } catch (e) { error.value = e.message } finally { loading.value = false }
}
onMounted(load)
// Menu công khai đọc store riêng → làm mới sau mỗi thay đổi
function refreshPublic() { publicCats.load(true) }

function openNew() {
  Object.assign(f, { name: '', slug: '', description: '', is_visible: true, show_in_menu: true, show_on_home: false, sort_order: (list.value.at(-1)?.sort_order ?? 0) + 1 })
  errs.value = {}; msg.value = ''; editing.value = 'new'
}
function openEdit(c) {
  Object.assign(f, { name: c.name, slug: c.slug, description: c.description, is_visible: c.is_visible, show_in_menu: c.show_in_menu, show_on_home: c.show_on_home, sort_order: c.sort_order })
  errs.value = {}; msg.value = ''; editing.value = c
}
function replaceInList(c) {
  const i = list.value.findIndex((x) => x.id === c.id)
  if (i >= 0) list.value[i] = c; else list.value.push(c)
  list.value.sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
}

async function save() {
  errs.value = {}; msg.value = ''; saving.value = true
  const body = { name: f.name, slug: f.slug.trim() || null, description: f.description, is_visible: f.is_visible,
    show_in_menu: f.show_in_menu, show_on_home: f.show_on_home, sort_order: Number(f.sort_order) || 0 }
  try {
    const isNew = editing.value === 'new'
    const c = isNew ? await categoriesApi.create(body) : await categoriesApi.update(editing.value.id, body)
    replaceInList(c)
    refreshPublic()
    toast.success(isNew ? 'Đã thêm danh mục — có thể tải ảnh bìa ngay' : 'Đã lưu danh mục')
    // Sau khi tạo mới, giữ cửa sổ mở ở chế độ sửa để tải ảnh bìa
    if (isNew) openEdit(c); else editing.value = null
  } catch (e) {
    errs.value = e.fields || {}
    msg.value = Object.keys(errs.value).length ? 'Vui lòng kiểm tra các ô được đánh dấu.' : e.message
  } finally { saving.value = false }
}

// Ảnh bìa: thu nhỏ cạnh dài ≤ 1920px (đọc bằng data URL vì CSP không cho blob:)
function shrink(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    const img = new Image()
    reader.onerror = () => reject(new Error('Không đọc được tệp'))
    reader.onload = () => { img.src = reader.result }
    img.onerror = () => reject(new Error('Tệp không phải ảnh hợp lệ'))
    img.onload = () => {
      const scale = Math.min(1, 1920 / Math.max(img.width, img.height))
      if (scale === 1 && file.size <= 2 * 1024 * 1024) return resolve({ blob: file, name: file.name })
      const cv = document.createElement('canvas')
      cv.width = Math.round(img.width * scale); cv.height = Math.round(img.height * scale)
      cv.getContext('2d').drawImage(img, 0, 0, cv.width, cv.height)
      cv.toBlob((b) => {
        if (b && b.type === 'image/webp') return resolve({ blob: b, name: 'bia.webp' })
        cv.toBlob((j) => (j ? resolve({ blob: j, name: 'bia.jpg' }) : reject(new Error('Không xử lý được ảnh'))), 'image/jpeg', 0.85)
      }, 'image/webp', 0.85)
    }
    reader.readAsDataURL(file)
  })
}
async function onCover(e) {
  const file = e.target.files?.[0]
  e.target.value = ''
  if (!file || editing.value === 'new' || !editing.value) return
  if (!/^image\/(jpeg|png|webp)$/.test(file.type)) { toast.error('Ảnh bìa chỉ nhận JPEG, PNG hoặc WebP'); return }
  coverBusy.value = true
  try {
    const { blob, name } = await shrink(file)
    if (blob.size > 2 * 1024 * 1024) throw new Error('Ảnh vượt quá 2 MB')
    const c = await categoriesApi.uploadCover(editing.value.id, blob, name)
    replaceInList(c); editing.value = c; refreshPublic()
    toast.success('Đã cập nhật ảnh bìa')
  } catch (err) { toast.error(err.message) } finally { coverBusy.value = false }
}
async function removeCover() {
  coverBusy.value = true
  try {
    const c = await categoriesApi.removeCover(editing.value.id)
    replaceInList(c); editing.value = c; refreshPublic()
    toast.success('Đã xóa ảnh bìa')
  } catch (err) { toast.error(err.message) } finally { coverBusy.value = false }
}

async function remove() {
  deleting.value = true
  try {
    await categoriesApi.remove(delTarget.value.id)
    list.value = list.value.filter((x) => x.id !== delTarget.value.id)
    refreshPublic()
    toast.success('Đã xóa danh mục')
    delTarget.value = null
  } catch (e) { toast.error(e.message) } finally { deleting.value = false }
}
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Danh mục sản phẩm</h1><p>Mỗi danh mục có trang riêng cho khách (vd. Đèn dự án). Sản phẩm không thuộc danh mục nào hiện ở trang chủ.</p></div>
      <button class="btn btn-primary" @click="openNew"><AppIcon name="plus" /> Thêm danh mục</button>
    </div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="error" kind="error" title="Chưa tải được danh mục" :text="error" @retry="load" />
    <StateBlock v-else-if="!list.length" title="Chưa có danh mục" text="Thêm danh mục để tạo trang riêng cho một nhóm sản phẩm." />
    <div v-else class="cards">
      <article v-for="c in list" :key="c.id" class="panel card" :class="{ off: !c.is_visible }">
        <div class="cover" :style="c.cover_url ? { backgroundImage: `url(&quot;${c.cover_url}&quot;)` } : {}">
          <span v-if="!c.cover_url" class="muted small">Chưa có ảnh bìa</span>
        </div>
        <div class="info">
          <h2>{{ c.name }}</h2>
          <p class="muted small mono">/danh-muc/{{ c.slug }}</p>
          <p class="tags">
            <span class="badge" :class="c.is_visible ? 'badge-ok' : 'badge-muted'">{{ c.is_visible ? 'Đang hiển thị' : 'Đang ẩn' }}</span>
            <span v-if="c.show_in_menu" class="badge badge-blue">Trên menu</span>
            <span v-if="c.show_on_home" class="badge badge-amber">Cả trang chủ</span>
            <span class="badge badge-muted">{{ c.product_count }} sản phẩm</span>
          </p>
          <div class="row actions">
            <a v-if="c.is_visible" :href="'/danh-muc/' + c.slug" target="_blank" rel="noopener" class="btn btn-ghost btn-sm"><AppIcon name="external" :size="16" /> Xem trang</a>
            <button class="btn btn-ghost btn-sm" @click="openEdit(c)"><AppIcon name="edit" :size="16" /> Sửa</button>
            <button class="btn btn-ghost btn-sm danger" :title="c.product_count ? 'Còn sản phẩm, chưa xóa được' : 'Xóa danh mục'" :disabled="c.product_count > 0" @click="delTarget = c"><AppIcon name="trash" :size="16" /></button>
          </div>
        </div>
      </article>
    </div>
    <p class="muted small note">Để đưa sản phẩm vào danh mục: mở <RouterLink to="/noi-bo/san-pham">Sản phẩm</RouterLink> → Sửa → chọn ô “Danh mục”.</p>

    <ModalShell v-if="editing" :title="editing === 'new' ? 'Thêm danh mục' : 'Sửa danh mục'" :busy="saving || coverBusy" wide @close="editing = null">
      <form id="cat-form" class="form-grid cols-2" novalidate @submit.prevent="save">
        <div class="field"><label for="cn">Tên danh mục</label><input id="cn" v-model="f.name" class="input" maxlength="100" :class="{ invalid: errs.name }" /><span v-if="errs.name" class="err">{{ errs.name }}</span></div>
        <div class="field"><label for="cs">Đường dẫn</label><input id="cs" v-model="f.slug" class="input" placeholder="Để trống để tự tạo từ tên" :class="{ invalid: errs.slug }" /><span v-if="errs.slug" class="err">{{ errs.slug }}</span><span v-else class="hint">Trang: /danh-muc/{{ f.slug || '…' }}</span></div>
        <div class="field span-2"><label for="cd">Mô tả (hiện dưới tiêu đề trang)</label><textarea id="cd" v-model="f.description" class="textarea" maxlength="2000" style="min-height:90px" /></div>
        <div class="field"><label for="co">Thứ tự trên menu</label><input id="co" v-model.number="f.sort_order" class="input num" type="number" /><span class="hint">Số nhỏ đứng trước.</span></div>
        <div class="field checks">
          <label class="check"><input v-model="f.is_visible" type="checkbox" /> Hiển thị cho khách</label>
          <label class="check"><input v-model="f.show_in_menu" type="checkbox" /> Hiện trên menu đầu trang</label>
          <label class="check"><input v-model="f.show_on_home" type="checkbox" /> Sản phẩm cũng hiện ở trang chủ</label>
        </div>
        <div v-if="editing !== 'new'" class="span-2 cover-edit">
          <div class="cover big" :style="editing.cover_url ? { backgroundImage: `url(&quot;${editing.cover_url}&quot;)` } : {}">
            <span v-if="!editing.cover_url" class="muted small">Chưa có ảnh bìa — trang dùng nền xanh mặc định</span>
          </div>
          <div>
            <p class="label">Ảnh bìa</p>
            <p class="muted small">Ảnh ngang (khuyến nghị 1920×600), làm nền phần đầu trang danh mục. JPEG, PNG, WebP; ảnh lớn được tự thu nhỏ.</p>
            <div class="row">
              <input ref="coverInput" type="file" accept="image/png,image/jpeg,image/webp" class="sr-only" @change="onCover" />
              <button type="button" class="btn btn-ghost btn-sm" :disabled="coverBusy" @click="coverInput.click()"><span v-if="coverBusy" class="spinner" /><AppIcon v-else name="upload" :size="16" /> {{ editing.cover_url ? 'Đổi ảnh bìa' : 'Tải ảnh bìa' }}</button>
              <button v-if="editing.cover_url" type="button" class="btn btn-text" :disabled="coverBusy" @click="removeCover">Xóa ảnh bìa</button>
            </div>
          </div>
        </div>
        <p v-else class="span-2 muted small">Lưu danh mục xong là tải được ảnh bìa.</p>
        <div v-if="msg" class="alert alert-error span-2" role="alert">{{ msg }}</div>
      </form>
      <template #footer>
        <button class="btn btn-ghost" :disabled="saving" @click="editing = null">Đóng</button>
        <button class="btn btn-primary" type="submit" form="cat-form" :disabled="saving"><span v-if="saving" class="spinner" />{{ editing === 'new' ? 'Thêm danh mục' : 'Lưu thay đổi' }}</button>
      </template>
    </ModalShell>

    <ConfirmDialog v-if="delTarget" :title="'Xóa danh mục “' + delTarget.name + '”?'" message="Trang danh mục sẽ không còn truy cập được." confirm-text="Xóa danh mục" danger :busy="deleting" @confirm="remove" @cancel="delTarget = null" />
  </div>
</template>

<style scoped>
.cards { display: grid; gap: 16px; grid-template-columns: repeat(auto-fill, minmax(min(100%, 340px), 1fr)); }
.card { overflow: hidden; display: flex; flex-direction: column; }
.card.off .info h2 { color: var(--ink-500); }
.cover { aspect-ratio: 16 / 5; background: linear-gradient(135deg, var(--sea-600), var(--sea-900)) center / cover no-repeat; display: grid; place-items: center; }
.cover .muted { color: rgba(255,255,255,.75); }
.cover.big { aspect-ratio: 16 / 5; border-radius: 10px; }
.info { padding: 14px 16px 16px; display: flex; flex-direction: column; gap: 6px; flex: 1; }
.info h2 { font-size: 18px; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; overflow-wrap: anywhere; }
.tags { display: flex; flex-wrap: wrap; gap: 6px; }
.actions { margin-top: auto; padding-top: 6px; gap: 8px; flex-wrap: wrap; }
.danger { color: var(--danger-600); }
.note { margin-top: 16px; }
.checks { justify-content: center; gap: 0; }
.cover-edit { display: grid; gap: 14px; grid-template-columns: 1fr; }
@media (min-width: 640px) { .cover-edit { grid-template-columns: 1.2fr 1fr; align-items: center; } }
.label { font-weight: 600; margin-bottom: 4px; }
.err { color: var(--danger-600); font-size: 13px; }
</style>
