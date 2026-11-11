<script setup>
import { formatVND } from '../utils/format'
import ProductPlaceholder from './ProductPlaceholder.vue'
defineProps({ product: { type: Object, required: true } })
defineEmits(['open'])
</script>

<template>
  <article class="card">
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
.card { background: var(--surface); border: 1px solid var(--line); border-radius: var(--r-lg); overflow: hidden; transition: border-color .15s, box-shadow .15s; }
.card:hover { border-color: var(--sea-100); box-shadow: var(--shadow-2); }
.hit { display: flex; flex-direction: column; width: 100%; height: 100%; padding: 0; border: 0; background: none; text-align: left; cursor: pointer; }
.media { aspect-ratio: 1 / 1; background: var(--mist); overflow: hidden; }
.media img { width: 100%; height: 100%; object-fit: cover; transition: transform .3s; }
.card:hover .media img { transform: scale(1.03); }
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
