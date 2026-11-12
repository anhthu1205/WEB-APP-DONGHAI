<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useCompanyStore } from '../stores/company'
import { useCategoriesStore } from '../stores/categories'
import { telHref } from '../utils/format'
import AppIcon from './AppIcon.vue'

const company = useCompanyStore()
const route = useRoute()
const open = ref(false)
const phone = computed(() => company.data?.phone || '')
const categories = useCategoriesStore()
categories.load()
const links = computed(() => [
  { to: { path: '/', hash: '#san-pham' }, label: 'Sản phẩm' },
  ...categories.menu.map((c) => ({ to: { path: '/danh-muc/' + c.slug }, label: c.name })),
  { to: { path: '/bao-hanh' }, label: 'Bảo hành' },
  // Liên hệ: chân trang có ở mọi trang công khai → cuộn tới #lien-he ngay trên trang hiện tại
  { to: { path: route.path, hash: '#lien-he' }, label: 'Liên hệ' },
])
</script>

<template>
  <header class="hdr">
    <div class="container bar">
      <RouterLink to="/" class="brand" aria-label="Đông Hải — trang chủ" @click="open = false">
        <img :src="company.logoUrl" alt="" height="36" class="logo" />
        <span class="name">Đông Hải</span>
      </RouterLink>
      <nav class="nav" :class="{ open }" aria-label="Điều hướng chính">
        <RouterLink v-for="l in links" :key="l.label" :to="l.to" class="nav-link" @click="open = false">{{ l.label }}</RouterLink>
      </nav>
      <div class="end">
        <a v-if="phone" :href="telHref(phone)" class="call">
          <span class="call-ico"><AppIcon name="phone" :size="18" /></span>
          <span class="call-txt"><small>Gọi đặt hàng</small><strong class="num">{{ phone }}</strong></span>
        </a>
        <button class="burger btn btn-ghost btn-icon" :aria-expanded="open" aria-label="Mở menu" @click="open = !open">
          <AppIcon :name="open ? 'close' : 'menu'" />
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.hdr { position: sticky; top: 0; z-index: 50; background: rgba(255,255,255,.94); backdrop-filter: saturate(1.4) blur(10px); border-bottom: 1px solid var(--line); }
.bar { display: flex; align-items: center; gap: 24px; height: 64px; }
.brand { display: flex; align-items: center; gap: 10px; text-decoration: none !important; }
.brand .name { font-weight: 800; font-size: 20px; color: var(--sea-800); letter-spacing: -.02em; }
.nav { display: flex; gap: 4px; }
.nav-link { padding: 8px 14px; border-radius: 8px; color: var(--ink-700); font-weight: 500; text-decoration: none !important; }
.nav-link:hover { background: var(--sea-50); color: var(--sea-800); }
.nav-link.router-link-exact-active:not([href*="#"]) { color: var(--sea-800); background: var(--sea-50); }
.end { margin-left: auto; display: flex; align-items: center; gap: 8px; }
.call { display: flex; align-items: center; gap: 10px; text-decoration: none !important; color: var(--ink-900); }
.call-ico { width: 38px; height: 38px; border-radius: 50%; display: grid; place-items: center; background: var(--amber-500); color: var(--ink-900); }
.call-txt { display: flex; flex-direction: column; line-height: 1.2; }
.call-txt small { font-size: 12px; color: var(--ink-500); }
.call-txt strong { font-size: 16px; color: var(--sea-800); }
.burger { display: none; }
@media (max-width: 767px) {
  .bar { gap: 12px; }
  .burger { display: inline-flex; }
  .call-txt { display: none; }
  .nav { display: none; position: absolute; top: 64px; left: 0; right: 0; flex-direction: column; background: #fff; border-bottom: 1px solid var(--line); padding: 8px 12px 14px; box-shadow: var(--shadow-2); }
  .nav.open { display: flex; }
  .nav-link { padding: 12px 14px; }
}
.logo { height: 36px; width: auto; max-width: 160px; object-fit: contain; flex-shrink: 0; border-radius: 8px; }
</style>
