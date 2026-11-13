<script setup>
import { formatVND } from '../utils/format'
import ProductPlaceholder from './ProductPlaceholder.vue'
defineProps({ product: { type: Object, required: true } })
defineEmits(['open'])
</script>

<template>
  <article class="card">
    <span v-if="product.badge" class="badge-tag" :title="product.badge">{{ product.badge }}</span>
    <button class="hit" :aria-label="'Xem chi tiết ' + product.name" @click="$emit('open', product)">
      <div class="media">
        <img v-if="product.main_image_url" :src="product.main_image_url" :alt="product.name" loading="lazy" width="400" height="400" />
        <ProductPlaceholder v-else :seed="product.id" :label="product.name.split(' ').slice(-1)[0]" />
      </div>
      <div class="body">
        <h3 class="name">{{ product.name }}</h3>
        <p class="desc">{{ product.short_description }}</p>
        <div class="foot">
          <span class="price">{{ formatVND(product.price) }}</span>
          <span class="unit">/ {{ product.unit }}</span>
        </div>
        <span class="more">Xem chi tiết</span>
      </div>
    </button>
  </article>
</template>

<style scoped>
.card { position: relative; background: var(--surface); border: 1px solid var(--line); border-radius: var(--r-lg); overflow: hidden;
  transition: border-color .2s, box-shadow .25s, transform .25s cubic-bezier(.2,.7,.2,1); }
.card:hover, .card:focus-within { border-color: var(--sea-100); box-shadow: 0 18px 36px -18px rgba(11,61,145,.35), var(--shadow-2); transform: translateY(-4px); }
/* Khi đang hiện dần (v-reveal) thì không ghi đè transform của reveal */
.card.reveal:not(.is-in) { transform: translateY(18px); }
/* Vệt sáng lướt qua ảnh khi rê chuột */
.media { position: relative; }
.media::after { content: ''; position: absolute; inset: 0; pointer-events: none;
  background: linear-gradient(115deg, transparent 30%, rgba(255,255,255,.45) 48%, transparent 66%);
  transform: translateX(-130%); }
.card:hover .media::after { transform: translateX(130%); transition: transform .8s ease; }
/* Nhãn góc trên bên phải (vd. "Mua 5 tặng 1") */
.badge-tag { position: absolute; top: 10px; right: 10px; z-index: 2; max-width: calc(100% - 20px);
  padding: 6px 12px; border-radius: 999px; font-size: 13px; font-weight: 700; line-height: 1.2; color: #fff;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; pointer-events: none;
  background: linear-gradient(135deg, #ef4444, #f97316); box-shadow: 0 6px 14px -6px rgba(220,38,38,.65);
  animation: tag-pop .5s cubic-bezier(.3,1.6,.5,1) .3s both; }
.badge-tag::after { content: ''; position: absolute; inset: 0; border-radius: inherit;
  background: linear-gradient(110deg, transparent 35%, rgba(255,255,255,.55) 50%, transparent 65%);
  transform: translateX(-120%); animation: shine 3.8s ease-in-out 1.5s infinite; }
@keyframes tag-pop { from { opacity: 0; transform: scale(.6) rotate(-6deg); } }
.card:hover .badge-tag { transform: scale(1.05); transition: transform .2s; }
@media (max-width: 480px) { .badge-tag { top: 8px; right: 8px; padding: 4px 9px; font-size: 11.5px; } }
.more { display: inline-flex; align-items: center; gap: 4px; }
.more::after { content: '→'; transition: transform .2s; }
.card:hover .more::after { transform: translateX(4px); }
.hit { display: flex; flex-direction: column; width: 100%; height: 100%; padding: 0; border: 0; background: none; text-align: left; cursor: pointer; }
.media { aspect-ratio: 1 / 1; background: var(--mist); overflow: hidden; }
.media img { width: 100%; height: 100%; object-fit: cover; transition: transform .5s cubic-bezier(.2,.7,.2,1); }
.card:hover .media img { transform: scale(1.07); }
.media :deep(svg) { transition: transform .4s cubic-bezier(.2,.7,.2,1); }
.card:hover .media :deep(svg) { transform: scale(1.08) rotate(-2deg); }
.body { padding: 14px 14px 16px; display: flex; flex-direction: column; gap: 6px; flex: 1; }
.name { font-size: 15px; font-weight: 600; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; min-height: 2.8em; }
.desc { font-size: 13px; color: var(--ink-500); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; min-height: 3em; }
.foot { margin-top: auto; padding-top: 6px; display: flex; align-items: baseline; gap: 4px; }
.price { font-size: 18px; }
.unit { font-size: 13px; color: var(--ink-400); }
.more { font-size: 14px; font-weight: 600; color: var(--sea-600); }
@media (max-width: 480px) {
  .body { padding: 10px 10px 12px; }
  .name { font-size: 14px; }
  .desc { display: none; }
  .price { font-size: 16px; }
}
</style>
