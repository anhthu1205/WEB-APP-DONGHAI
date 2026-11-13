<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publicApi } from '../../api/public'
import { useCompanyStore } from '../../stores/company'
import { telHref } from '../../utils/format'
import AppIcon from '../../components/AppIcon.vue'
import ProductCard from '../../components/ProductCard.vue'
import ProductModal from '../../components/ProductModal.vue'
import HeroLights from '../../components/HeroLights.vue'

// Trang danh mục (ngách) sản phẩm: /danh-muc/:slug — vd. Đèn dự án.
const route = useRoute()
const router = useRouter()
const company = useCompanyStore()
const category = ref(null)
const products = ref([])
const loading = ref(true)
const error = ref(null)
const notFound = ref(false)
const q = ref('')

const slug = computed(() => String(route.params.slug || ''))
const phone = computed(() => company.data?.phone || '')
const openSlug = computed(() => (typeof route.query.sp === 'string' ? route.query.sp : ''))
const preview = computed(() => products.value.find((p) => p.slug === openSlug.value) || null)
const heroStyle = computed(() => (category.value?.cover_url ? { '--cover': `url("${category.value.cover_url}")` } : {}))

function strip(s) {
  return s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/đ/g, 'd')
}
const filtered = computed(() => {
  const k = strip(q.value.trim())
  if (!k) return products.value
  return products.value.filter((p) => strip(p.name + ' ' + p.short_description).includes(k))
})

async function load() {
  loading.value = true; error.value = null; notFound.value = false
  try {
    const d = await publicApi.category(slug.value)
    category.value = d.category
    products.value = d.products
    document.title = d.category.name + ' — Đông Hải'
  } catch (e) {
    if (e.status === 404) notFound.value = true
    else error.value = e.message
    category.value = null; products.value = []
  } finally { loading.value = false }
}
// Chuyển giữa các danh mục (cùng component) → tải lại
watch(slug, () => { q.value = ''; load() }, { immediate: true })

function open(p) { router.replace({ query: { ...route.query, sp: p.slug } }) }
function close() {
  const { sp, ...rest } = route.query
  router.replace({ query: rest })
}
</script>

<template>
  <section class="hero" :class="{ 'has-cover': category?.cover_url }" :style="heroStyle">
    <HeroLights />
    <div class="container hero-in">
      <nav class="crumbs small enter" aria-label="Đường dẫn">
        <RouterLink to="/">Trang chủ</RouterLink><span aria-hidden="true">/</span><span>{{ category?.name || 'Danh mục' }}</span>
      </nav>
      <template v-if="category">
        <h1 class="enter" style="--d:80ms">{{ category.name }}</h1>
        <p v-if="category.description" class="sub enter" style="--d:180ms">{{ category.description }}</p>
        <div class="ctas enter" style="--d:280ms">
          <a href="#ds-san-pham" class="btn btn-accent btn-shine">Xem {{ category.product_count }} sản phẩm</a>
          <a v-if="phone" :href="telHref(phone)" class="btn btn-light"><AppIcon name="phone" /> Gọi báo giá {{ phone }}</a>
        </div>
      </template>
      <template v-else-if="loading">
        <div class="skeleton sk-title" /><div class="skeleton sk-sub" />
      </template>
      <h1 v-else-if="notFound">Không tìm thấy danh mục</h1>
      <h1 v-else>Chưa tải được danh mục</h1>
    </div>
    <svg class="wave" viewBox="0 0 1440 80" preserveAspectRatio="none" aria-hidden="true">
      <path class="w1" d="M0 40c160-30 320-30 480 0s320 30 480 0 320-30 480 0v40H0z" />
      <path class="w2" d="M0 56c180-22 360-22 540 0s360 22 540 0 270-16 360-8v32H0z" />
    </svg>
  </section>

  <section id="ds-san-pham" class="container products">
    <div v-if="loading" class="grid" aria-busy="true">
      <div v-for="i in 8" :key="i" class="sk-card">
        <div class="skeleton" style="aspect-ratio:1/1;border-radius:0" />
        <div style="padding:14px"><div class="skeleton" style="height:14px;width:85%" /><div class="skeleton" style="height:18px;width:45%;margin-top:18px" /></div>
      </div>
    </div>
    <div v-else-if="notFound" class="panel state">
      <AppIcon name="search" />
      <h3>Danh mục này không tồn tại hoặc đã ngừng hiển thị</h3>
      <RouterLink to="/" class="btn btn-ghost btn-sm">Về trang chủ</RouterLink>
    </div>
    <div v-else-if="error" class="panel state">
      <AppIcon name="alert" />
      <h3>Chưa tải được danh sách sản phẩm</h3>
      <p>{{ error }}</p>
      <button class="btn btn-ghost btn-sm" @click="load">Tải lại</button>
    </div>
    <template v-else>
      <div v-reveal class="sec-head">
        <div>
          <h2>Sản phẩm {{ category.name.toLowerCase() }}</h2>
          <p class="muted">{{ products.length }} sản phẩm · Giá niêm yết, liên hệ để có giá dự án theo số lượng</p>
        </div>
        <div v-if="products.length > 4" class="search">
          <AppIcon name="search" :size="18" />
          <label class="sr-only" for="cat-search">Tìm trong danh mục</label>
          <input id="cat-search" v-model="q" class="input" type="search" placeholder="Tìm trong danh mục" />
        </div>
      </div>
      <div v-if="!products.length" class="panel state">
        <AppIcon name="box" />
        <h3>Đang cập nhật sản phẩm</h3>
        <p>Vui lòng gọi {{ phone || 'cho chúng tôi' }} để được tư vấn và báo giá.</p>
      </div>
      <div v-else-if="!filtered.length" class="panel state">
        <AppIcon name="search" />
        <h3>Không có sản phẩm khớp “{{ q }}”</h3>
        <button class="btn btn-ghost btn-sm" @click="q = ''">Xem tất cả</button>
      </div>
      <div v-else class="grid">
        <ProductCard v-for="(p, i) in filtered" :key="p.id" v-reveal="i" :product="p" @open="open" />
      </div>
    </template>
  </section>

  <ProductModal v-if="openSlug" :slug="openSlug" :preview="preview" @close="close" />
</template>

<style scoped>
.hero { position: relative; color: #fff; overflow: hidden; padding: 40px 0 110px; background:
  radial-gradient(60% 80% at 85% 10%, rgba(47,124,240,.55), transparent 60%),
  linear-gradient(135deg, var(--sea-600) 0%, var(--sea-800) 55%, var(--sea-900) 100%); }
/* Có ảnh bìa: phủ lớp màu để chữ trắng luôn đọc được */
.hero.hero.has-cover { background:
  linear-gradient(100deg, rgba(7,42,99,.92) 0%, rgba(7,42,99,.78) 45%, rgba(7,42,99,.35) 100%),
  var(--cover) center / cover no-repeat, var(--sea-900); }
@media (min-width: 960px) { .hero { padding: 56px 0 130px; } }
.crumbs { display: flex; gap: 8px; color: rgba(255,255,255,.7); }
.crumbs a { color: rgba(255,255,255,.85); }
h1 { color: #fff; font-size: clamp(30px, 4.6vw, 48px); font-weight: 800; line-height: 1.12; letter-spacing: -.025em; margin-top: 14px; max-width: 20ch; }
.sub { margin-top: 14px; font-size: 17px; color: rgba(255,255,255,.85); max-width: 60ch; line-height: 1.65; }
.ctas { margin-top: 26px; display: flex; flex-wrap: wrap; gap: 12px; }
.ctas .btn { min-height: 50px; padding: 0 22px; font-size: 16px; }
.sk-title { height: 44px; width: min(420px, 80%); margin-top: 16px; opacity: .25; }
.sk-sub { height: 18px; width: min(560px, 90%); margin-top: 16px; opacity: .2; }
.wave { position: absolute; left: 0; bottom: -1px; width: 100%; height: 80px; }
.wave .w1 { fill: rgba(243,246,250,.35); }
.wave .w2 { fill: var(--mist); }

.products { padding-top: 24px; padding-bottom: 48px; scroll-margin-top: 80px; }
.sec-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 20px; }
.sec-head h2 { font-size: 26px; font-weight: 700; }
.search { position: relative; flex: 0 1 340px; }
.search svg { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--ink-400); }
.search .input { padding-left: 38px; }
.grid { display: grid; gap: 16px; grid-template-columns: repeat(2, minmax(0, 1fr)); }
@media (min-width: 720px) { .grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (min-width: 1024px) { .grid { grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 20px; } }
@media (max-width: 480px) { .grid { gap: 10px; } .search { flex-basis: 100%; } }
.sk-card { background: #fff; border: 1px solid var(--line); border-radius: var(--r-lg); overflow: hidden; }
</style>
