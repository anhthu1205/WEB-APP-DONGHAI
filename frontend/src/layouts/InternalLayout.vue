<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useCompanyStore } from '../stores/company'
import AppIcon from '../components/AppIcon.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const drawer = ref(false)
watch(() => route.fullPath, () => { drawer.value = false })

// Menu sinh từ cấu hình, lọc theo quyền
const groups = computed(() => {
  const g = [{
    label: null,
    items: [
      { to: '/noi-bo', label: 'Tổng quan', icon: 'dashboard', exact: true },
      { to: '/noi-bo/san-pham', label: 'Sản phẩm', icon: 'box' },
      { to: '/noi-bo/bang-gia', label: 'Khuyến mãi & bảng giá', icon: 'tag' },
      { to: '/noi-bo/khach-hang', label: 'Khách hàng', icon: 'users' },
      { to: '/noi-bo/don-hang', label: 'Đơn hàng', icon: 'receipt' },
      { to: '/noi-bo/doi-mat-khau', label: 'Đổi mật khẩu', icon: 'key' },
    ],
  }]
  if (auth.isAdmin) {
    g.push({
      label: 'Quản trị',
      items: [
        { to: '/noi-bo/danh-muc', label: 'Danh mục sản phẩm', icon: 'folder' },
        { to: '/noi-bo/khuyen-mai', label: 'Quản lý khuyến mãi', icon: 'star' },
        { to: '/noi-bo/cong-ty', label: 'Thông tin công ty', icon: 'building' },
        { to: '/noi-bo/tai-khoan', label: 'Tài khoản', icon: 'lock' },
      ],
    })
  }
  return g
})
function isActive(it) {
  return it.exact ? route.path === it.to : route.path === it.to || route.path.startsWith(it.to + '/')
}
function logout() {
  auth.logout()
  router.replace('/dang-nhap')
}
const company = useCompanyStore()
company.load()
</script>

<template>
  <div class="shell" :class="{ open: drawer }">
    <aside class="side" aria-label="Menu nội bộ">
      <RouterLink to="/noi-bo" class="brand"><img :src="company.logoUrl" alt="" height="32" class="logo" /><span>Đông Hải<small>Nội bộ</small></span></RouterLink>
      <nav>
        <div v-for="(g, gi) in groups" :key="gi" class="group">
          <p v-if="g.label" class="g-label">{{ g.label }}</p>
          <RouterLink v-for="it in g.items" :key="it.to" :to="it.to" class="item" :class="{ active: isActive(it) }">
            <AppIcon :name="it.icon" :size="19" /><span>{{ it.label }}</span>
          </RouterLink>
        </div>
      </nav>
      <a href="/" target="_blank" rel="noopener" class="item site"><AppIcon name="external" :size="19" /><span>Xem trang công khai</span></a>
    </aside>
    <div class="scrim" @click="drawer = false" />
    <div class="main">
      <header class="top">
        <button class="btn btn-ghost btn-icon burger" aria-label="Mở menu" @click="drawer = true"><AppIcon name="menu" /></button>
        <div class="spacer" />
        <div class="me">
          <span class="avatar">{{ (auth.user?.full_name || auth.user?.username || '?').trim().charAt(0).toUpperCase() }}</span>
          <span class="who"><strong>{{ auth.user?.full_name || auth.user?.username }}</strong><small>{{ auth.isAdmin ? 'Quản trị viên' : 'Nhân viên' }}</small></span>
        </div>
        <button class="btn btn-ghost btn-sm" @click="logout"><AppIcon name="logout" :size="16" /> Đăng xuất</button>
      </header>
      <main class="content"><RouterView /></main>
    </div>
  </div>
</template>

<style scoped>
.shell { min-height: 100vh; display: grid; grid-template-columns: 248px 1fr; background: linear-gradient(to right, var(--sea-900) 248px, var(--mist) 248px); }
.side { position: sticky; top: 0; height: 100vh; overflow-y: auto; background: var(--sea-900); color: rgba(255,255,255,.75); padding: 18px 12px; display: flex; flex-direction: column; gap: 18px; }
.brand { display: flex; align-items: center; gap: 10px; padding: 4px 8px; color: #fff; text-decoration: none !important; font-weight: 800; font-size: 17px; }
.brand span { display: flex; flex-direction: column; line-height: 1.15; }
.brand small { font-weight: 500; font-size: 12px; color: rgba(255,255,255,.55); }
nav { display: flex; flex-direction: column; gap: 18px; flex: 1; }
.group { display: flex; flex-direction: column; gap: 2px; }
.g-label { font-size: 12px; font-weight: 600; color: rgba(255,255,255,.45); padding: 0 12px 6px; }
.item { display: flex; align-items: center; gap: 12px; min-height: 42px; padding: 0 12px; border-radius: 10px; color: rgba(255,255,255,.78); font-size: 14px; font-weight: 500; text-decoration: none !important; }
.item:hover { background: rgba(255,255,255,.07); color: #fff; }
.item.active { background: var(--sea-600); color: #fff; }
.item.site { color: rgba(255,255,255,.55); }
.main { min-width: 0; display: flex; flex-direction: column; }
.top { position: sticky; top: 0; z-index: 20; height: 60px; display: flex; align-items: center; gap: 12px; padding: 0 24px; background: rgba(255,255,255,.95); backdrop-filter: blur(8px); border-bottom: 1px solid var(--line); }
.burger { display: none; }
.me { display: flex; align-items: center; gap: 10px; }
.avatar { width: 34px; height: 34px; border-radius: 50%; background: var(--sea-50); color: var(--sea-800); display: grid; place-items: center; font-weight: 700; }
.who { display: flex; flex-direction: column; line-height: 1.2; font-size: 14px; }
.who small { color: var(--ink-500); font-size: 12px; }
.content { padding: 24px; max-width: 1240px; width: 100%; }
.scrim { display: none; }
@media (max-width: 1023px) {
  .shell { grid-template-columns: 1fr; background: var(--mist); }
  .side { position: fixed; z-index: 60; left: 0; top: 0; width: 272px; transform: translateX(-100%); transition: transform .2s ease; }
  .shell.open .side { transform: none; }
  .shell.open .scrim { display: block; position: fixed; inset: 0; z-index: 55; background: rgba(7,26,52,.5); }
  .burger { display: inline-flex; }
  .top { padding: 0 16px; }
  .content { padding: 16px; }
}
@media (max-width: 520px) { .who { display: none; } }
.logo { height: 32px; width: auto; max-width: 120px; object-fit: contain; flex-shrink: 0; border-radius: 8px; }
</style>
