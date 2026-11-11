<script setup>
import { ref, onMounted } from 'vue'
import { productsApi } from '../../api/products'
import { formatVND, formatDate } from '../../utils/format'
import AppIcon from '../../components/AppIcon.vue'
import StateBlock from '../../components/StateBlock.vue'

// Trang ẩn: KM đang/sắp diễn ra + bảng giá nội bộ (API trả 401 nếu chưa đăng nhập)
const data = ref(null)
const loading = ref(true)
const error = ref(null)
async function load() {
  loading.value = true; error.value = null
  try { data.value = await productsApi.priceList() } catch (e) { error.value = e.message } finally { loading.value = false }
}
onMounted(load)
</script>

<template>
  <div>
    <div class="page-head">
      <div><h1>Khuyến mãi & bảng giá nội bộ</h1><p>Chỉ dành cho nhân viên, không chia sẻ ra ngoài.</p></div>
      <span v-if="data" class="badge badge-blue"><AppIcon name="calendar" :size="14" /> Áp dụng ngày {{ formatDate(data.today) }}</span>
    </div>
    <StateBlock v-if="loading" kind="loading" />
    <StateBlock v-else-if="error" kind="error" title="Chưa tải được bảng giá" :text="error" @retry="load" />
    <template v-else>
      <h2 class="sec-title">Chương trình khuyến mãi</h2>
      <div v-if="!data.promotions.length" class="panel"><StateBlock title="Hiện không có chương trình khuyến mãi" text="Chương trình đang chạy hoặc sắp diễn ra sẽ hiện ở đây." /></div>
      <div v-else class="promos">
        <article v-for="p in data.promotions" :key="p.id" class="panel promo" :class="p.status">
          <header>
            <span class="badge" :class="p.status === 'active' ? 'badge-ok' : 'badge-amber'"><span class="dot" />{{ p.status === 'active' ? 'Đang diễn ra' : 'Sắp diễn ra' }}</span>
            <h3>{{ p.title }}</h3>
            <p class="muted small num">{{ formatDate(p.start_date) }} – {{ formatDate(p.end_date) }}</p>
          </header>
          <p v-if="p.content" class="pre-line content">{{ p.content }}</p>
          <table v-if="p.items.length" class="table mini">
            <thead><tr><th>Sản phẩm</th><th class="right">Giá KM</th></tr></thead>
            <tbody>
              <tr v-for="it in p.items" :key="it.product_id">
                <td>{{ it.product_name }}<span v-if="it.note" class="muted small"><br />{{ it.note }}</span></td>
                <td class="right num"><span class="strike small">{{ formatVND(it.price_public) }}</span><br /><span class="price">{{ formatVND(it.promo_price) }}</span></td>
              </tr>
            </tbody>
          </table>
        </article>
      </div>

      <h2 class="sec-title">Bảng giá sản phẩm</h2>
      <div class="panel table-wrap">
        <table class="table responsive">
          <thead><tr><th>Sản phẩm</th><th class="right">Giá niêm yết</th><th class="right">Giá nội bộ</th><th class="right">Giá KM hiện hành</th></tr></thead>
          <tbody>
            <tr v-for="p in data.products" :key="p.id">
              <td data-label="Sản phẩm"><strong>{{ p.name }}</strong><span class="muted small"> / {{ p.unit }}</span></td>
              <td data-label="Giá niêm yết" class="right num"><span :class="{ strike: p.current_promo_price !== null }">{{ formatVND(p.price_public) }}</span></td>
              <td data-label="Giá nội bộ" class="right num"><strong>{{ formatVND(p.price_internal) }}</strong></td>
              <td data-label="Giá KM hiện hành" class="right num">
                <template v-if="p.current_promo_price !== null"><span class="price">{{ formatVND(p.current_promo_price) }}</span><br /><span class="muted small">{{ p.current_promotion_title }}</span></template>
                <span v-else class="muted">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.sec-title { font-size: 17px; margin: 8px 0 12px; }
.promos { display: grid; gap: 14px; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); margin-bottom: 28px; }
.promo { padding: 18px; border-top: 4px solid var(--ok-700); }
.promo.upcoming { border-top-color: var(--amber-500); }
.promo h3 { font-size: 17px; margin-top: 10px; }
.content { margin-top: 10px; color: var(--ink-700); font-size: 14px; }
.mini { margin-top: 12px; }
.mini td, .mini th { padding: 8px 6px; }
</style>
