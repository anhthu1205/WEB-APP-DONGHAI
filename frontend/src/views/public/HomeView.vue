<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publicApi } from '../../api/public'
import { useCompanyStore } from '../../stores/company'
import { normalizePhone } from '../../utils/phone'
import AppIcon from '../../components/AppIcon.vue'
import ProductCard from '../../components/ProductCard.vue'
import ProductModal from '../../components/ProductModal.vue'
import PhoneInput from '../../components/PhoneInput.vue'
import HeroLights from '../../components/HeroLights.vue'

const route = useRoute()
const router = useRouter()
const company = useCompanyStore()
const products = ref([])
const loading = ref(true)
const error = ref(null)
const q = ref('')
const quickPhone = ref('')
const quickErr = ref('')

const c = computed(() => company.data)
const openSlug = computed(() => (typeof route.query.sp === 'string' ? route.query.sp : ''))
const preview = computed(() => products.value.find((p) => p.slug === openSlug.value) || null)

function strip(s) {
  return s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/đ/g, 'd')
}
const filtered = computed(() => {
  const k = strip(q.value.trim())
  if (!k) return products.value
  return products.value.filter((p) => strip(p.name + ' ' + p.short_description).includes(k))
})

async function load() {
  loading.value = true
  error.value = null
  try {
    products.value = await publicApi.products()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
onMounted(load)

function open(p) { router.replace({ query: { ...route.query, sp: p.slug } }) }
function close() {
  const { sp, ...rest } = route.query
  router.replace({ query: rest })
}
watch(openSlug, (s) => { if (s) document.title = 'Đông Hải — ' + (preview.value?.name || 'Sản phẩm') })

function quickLookup() {
  const p = normalizePhone(quickPhone.value)
  if (!p) { quickErr.value = 'Số điện thoại chưa đúng, ví dụ 0912 345 678'; return }
  quickErr.value = ''
  // Truyền SĐT qua history state (không đưa lên URL)
  router.push({ name: 'warranty', state: { phone: quickPhone.value } })
}
</script>

<template>
  <section class="hero">
    <HeroLights />
    <div class="container hero-in">
      <div class="copy">
        <p class="kicker enter">{{ c?.name || 'Đông Hải Lighting' }}</p>
        <h1 class="enter" style="--d:80ms">{{ c?.slogan || 'Đèn đánh cá, đèn công trình, dự án chính hãng' }}</h1>
        <p class="sub enter" style="--d:180ms">Bảo hành 2 - 5 năm, 1 đổi 1 trong 6 tháng - giá niêm yết, Miễn phí vận chuyển.</p>
        <div class="ctas enter" style="--d:280ms">
          <a href="#san-pham" class="btn btn-accent btn-shine">Xem sản phẩm</a>
          <RouterLink to="/bao-hanh" class="btn btn-light"><AppIcon name="shield" /> Tra cứu bảo hành</RouterLink>
        </div>
        <ul class="promises enter" style="--d:380ms">
          <li><AppIcon name="badge" :size="18" /> Hàng chính hãng</li>
          <li><AppIcon name="shield" :size="18" /> Bảo hành 2 - 5 năm</li>
          <li><AppIcon name="truck" :size="18" /> Miễn phí vận chuyển</li>
        </ul>
      </div>

       <form class="ticket enter" style="--d:300ms" aria-labelledby="ticket-title" @submit.prevent="quickLookup">
        <div class="stamp" aria-hidden="true"><strong>2</strong><span>năm</span></div>
        <h2 id="ticket-title">Phiếu bảo hành điện tử</h2>
        <p>Nhập số điện thoại bạn đã dùng khi mua hàng để xem hạn bảo hành của từng đơn.</p>
        <label class="sr-only" for="quick-phone">Số điện thoại</label>
        <PhoneInput id="quick-phone" v-model="quickPhone" :invalid="!!quickErr" />
        <p v-if="quickErr" class="err" role="alert">{{ quickErr }}</p>
        <button class="btn btn-primary btn-block" type="submit"><AppIcon name="search" /> Tra cứu ngay</button>
        <div class="perf" aria-hidden="true" />
        <p class="fine">Không cần đăng nhập. Thông tin cá nhân được che bớt khi hiển thị.</p>
      </form>
    </div>
    <svg class="wave" viewBox="0 0 1440 80" preserveAspectRatio="none" aria-hidden="true">
      <path class="w1" d="M0 40c160-30 320-30 480 0s320 30 480 0 320-30 480 0v40H0z" />
      <path class="w2" d="M0 56c180-22 360-22 540 0s360 22 540 0 270-16 360-8v32H0z" />
    </svg>
  </section>

  <section id="san-pham" class="container products">
    <div v-reveal class="sec-head">
      <div>
        <h2>Sản phẩm đang bán</h2>
        <p v-if="!loading && !error" class="muted">{{ products.length }} sản phẩm, giá niêm yết tại cửa hàng</p>
      </div>
      <div class="search">
        <AppIcon name="search" :size="18" />
        <label class="sr-only" for="product-search">Tìm sản phẩm</label>
        <input id="product-search" v-model="q" class="input" type="search" placeholder="Tìm theo tên sản phẩm" />
      </div>
    </div>

    <div v-if="loading" class="grid" aria-busy="true">
      <div v-for="i in 8" :key="i" class="sk-card">
        <div class="skeleton" style="aspect-ratio:1/1;border-radius:0" />
        <div style="padding:14px"><div class="skeleton" style="height:14px;width:85%" /><div class="skeleton" style="height:12px;width:60%;margin-top:10px" /><div class="skeleton" style="height:18px;width:45%;margin-top:18px" /></div>
      </div>
    </div>
    <div v-else-if="error" class="panel state">
      <AppIcon name="alert" />
      <h3>Chưa tải được danh sách sản phẩm</h3>
      <p>{{ error }}</p>
      <button class="btn btn-ghost btn-sm" @click="load">Tải lại</button>
    </div>
    <div v-else-if="!products.length" class="panel state">
      <AppIcon name="box" />
      <h3>Đang cập nhật sản phẩm</h3>
      <p>Vui lòng quay lại sau hoặc gọi {{ c?.phone || 'cho chúng tôi' }} để được tư vấn.</p>
    </div>
    <template v-else>
      <div v-if="!filtered.length" class="panel state">
        <AppIcon name="search" />
        <h3>Không có sản phẩm khớp “{{ q }}”</h3>
        <button class="btn btn-ghost btn-sm" @click="q = ''">Xem tất cả sản phẩm</button>
      </div>
      <div v-else class="grid">
        <ProductCard v-for="(p, i) in filtered" :key="p.id" v-reveal="i" :product="p" @open="open" />
      </div>
    </template>
  </section>

  <ProductModal v-if="openSlug" :slug="openSlug" :preview="preview" @close="close" />
</template>

<style scoped>
.hero { position: relative; color: #fff; background:
  radial-gradient(60% 80% at 85% 10%, rgba(47,124,240,.55), transparent 60%),
  linear-gradient(135deg, var(--sea-600) 0%, var(--sea-800) 55%, var(--sea-900) 100%);
  padding: 56px 0 120px; overflow: hidden; }
.hero-in { display: grid; gap: 40px; align-items: center; }
@media (min-width: 960px) { .hero-in { grid-template-columns: 1.25fr .9fr; gap: 64px; } .hero { padding: 80px 0 140px; } }
.kicker { font-weight: 600; color: var(--amber-500); font-size: 15px; }
h1 { color: #fff; font-size: clamp(30px, 4.6vw, 50px); font-weight: 800; line-height: 1.12; letter-spacing: -.025em; margin-top: 10px; max-width: 16ch; }
.sub { margin-top: 18px; font-size: 17px; color: rgba(255,255,255,.82); max-width: 46ch; line-height: 1.65; }
.ctas { margin-top: 28px; display: flex; flex-wrap: wrap; gap: 12px; }
.ctas .btn { min-height: 50px; padding: 0 22px; font-size: 16px; }
.promises { list-style: none; margin: 30px 0 0; padding: 0; display: flex; flex-wrap: wrap; gap: 10px 26px; color: rgba(255,255,255,.88); font-size: 14px; font-weight: 500; }
.promises li { display: flex; align-items: center; gap: 8px; }
.promises svg { color: var(--amber-500); }

.ticket { animation-name: enter-up, float; animation-duration: .7s, 6s; animation-timing-function: cubic-bezier(.2,.7,.2,1), ease-in-out; animation-delay: var(--d), 1.2s; animation-iteration-count: 1, infinite; animation-direction: normal, alternate; animation-fill-mode: both, none; }
@keyframes float { from { transform: translateY(0); } to { transform: translateY(-8px); } }
.stamp { transition: transform .3s cubic-bezier(.3,1.6,.5,1); }
.ticket:hover .stamp { transform: rotate(4deg) scale(1.06); }
.kicker { display: inline-flex; align-items: center; gap: 8px; }
.kicker::before { content: ''; width: 8px; height: 8px; border-radius: 50%; background: var(--amber-500); box-shadow: 0 0 0 0 rgba(245,158,11,.6); animation: blip 2.4s ease-out infinite; }
@keyframes blip { 0% { box-shadow: 0 0 0 0 rgba(245,158,11,.6); } 80%, 100% { box-shadow: 0 0 0 10px rgba(245,158,11,0); } }
h1 { background: linear-gradient(100deg, #fff 40%, #ffe7a3 50%, #fff 60%) 0 0 / 250% 100%; -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent; animation: enter-up .7s cubic-bezier(.2,.7,.2,1) both, glint 7s ease-in-out 1.5s infinite; animation-delay: var(--d, 0ms), 1.5s; }
@keyframes glint { 0%, 70% { background-position: 100% 0; } 100% { background-position: -50% 0; } }
.ticket { position: relative; background: #fff; color: var(--ink-900); border-radius: var(--r-xl); padding: 28px 24px 22px; box-shadow: var(--shadow-3); display: flex; flex-direction: column; gap: 12px; max-width: 440px; width: 100%; justify-self: end; }
@media (max-width: 959px) { .ticket { justify-self: stretch; max-width: none; } }
.ticket h2 { font-size: 20px; font-weight: 700; padding-right: 76px; }
.ticket > p { color: var(--ink-500); font-size: 14px; padding-right: 40px; }
.ticket .err { color: var(--danger-600); font-size: 13px; margin-top: -4px; }
.stamp { position: absolute; top: -18px; right: 18px; width: 82px; height: 82px; border-radius: 50%; background: var(--amber-500); color: var(--ink-900); display: flex; flex-direction: column; align-items: center; justify-content: center; line-height: 1; transform: rotate(-10deg); box-shadow: 0 0 0 5px rgba(245,158,11,.25), 0 10px 20px -8px rgba(0,0,0,.35); }
.stamp::after { content: ''; position: absolute; inset: 6px; border-radius: 50%; border: 1.5px dashed rgba(19,35,58,.45); }
.stamp strong { font-size: 30px; font-weight: 800; }
.stamp span { font-size: 12px; font-weight: 700; }
.perf { height: 0; border-top: 2px dashed var(--line-strong); margin: 6px -24px 0; position: relative; }
.perf::before, .perf::after { content: ''; position: absolute; top: -11px; width: 20px; height: 20px; border-radius: 50%; background: var(--sea-800); }
.perf::before { left: -10px; } .perf::after { right: -10px; }
.fine { font-size: 12px; color: var(--ink-400); }

.wave { position: absolute; left: 0; bottom: -1px; width: 100%; height: 80px; }
.wave .w1 { fill: rgba(243,246,250,.35); }
.wave .w2 { fill: var(--mist); }

.products { padding-top: 24px; scroll-margin-top: 80px; }
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
