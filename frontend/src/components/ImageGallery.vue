<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import AppIcon from './AppIcon.vue'
import ProductPlaceholder from './ProductPlaceholder.vue'

const props = defineProps({ images: { type: Array, default: () => [] }, alt: { type: String, default: '' }, seed: { type: Number, default: 0 } })
const idx = ref(0)
const zoom = ref(false)
const current = computed(() => props.images[idx.value])
watch(() => props.images, () => { idx.value = 0 })

function go(d) {
  const n = props.images.length
  if (n) idx.value = (idx.value + d + n) % n
}
function onKey(e) {
  if (!zoom.value) return
  if (e.key === 'Escape') { e.stopPropagation(); zoom.value = false }
  else if (e.key === 'ArrowLeft') go(-1)
  else if (e.key === 'ArrowRight') go(1)
}
watch(zoom, (z) => {
  if (z) document.addEventListener('keydown', onKey, true)
  else document.removeEventListener('keydown', onKey, true)
})
onBeforeUnmount(() => document.removeEventListener('keydown', onKey, true))
</script>

<template>
  <div class="gallery">
    <div class="main">
      <button v-if="current" class="main-btn" :aria-label="'Phóng to ảnh ' + alt" @click="zoom = true">
        <img :src="current.url" :alt="alt" />
        <span class="zoom-hint"><AppIcon name="zoom" :size="16" /> Phóng to</span>
      </button>
      <ProductPlaceholder v-else :seed="seed" />
      <template v-if="images.length > 1">
        <button class="nav prev" aria-label="Ảnh trước" @click="go(-1)"><AppIcon name="left" /></button>
        <button class="nav next" aria-label="Ảnh sau" @click="go(1)"><AppIcon name="right" /></button>
      </template>
    </div>
    <div v-if="images.length > 1" class="thumbs" role="tablist" aria-label="Ảnh sản phẩm">
      <button v-for="(im, i) in images" :key="im.url" class="thumb" :class="{ on: i === idx }" role="tab" :aria-selected="i === idx" :aria-label="'Ảnh ' + (i + 1)" @click="idx = i">
        <img :src="im.url" alt="" loading="lazy" />
      </button>
    </div>

    <Teleport to="body">
      <div v-if="zoom && current" class="lightbox" role="dialog" aria-modal="true" :aria-label="'Ảnh phóng to ' + alt" @click.self="zoom = false">
        <img :src="current.url" :alt="alt" />
        <button class="lb-close" aria-label="Đóng ảnh phóng to" @click="zoom = false"><AppIcon name="close" :size="24" /></button>
        <template v-if="images.length > 1">
          <button class="lb-nav prev" aria-label="Ảnh trước" @click="go(-1)"><AppIcon name="left" :size="28" /></button>
          <button class="lb-nav next" aria-label="Ảnh sau" @click="go(1)"><AppIcon name="right" :size="28" /></button>
          <span class="lb-count num">{{ idx + 1 }} / {{ images.length }}</span>
        </template>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.gallery { display: flex; flex-direction: column; gap: 10px; }
.main { position: relative; aspect-ratio: 1 / 1; border-radius: var(--r-lg); overflow: hidden; background: var(--mist); }
.main-btn { width: 100%; height: 100%; padding: 0; border: 0; background: none; cursor: zoom-in; position: relative; }
.main-btn img { width: 100%; height: 100%; object-fit: contain; background: #fff; }
.zoom-hint { position: absolute; right: 10px; bottom: 10px; display: inline-flex; align-items: center; gap: 6px; padding: 6px 10px; border-radius: 999px; background: rgba(19,35,58,.72); color: #fff; font-size: 12px; font-weight: 600; }
.nav { position: absolute; top: 50%; transform: translateY(-50%); width: 40px; height: 40px; border-radius: 50%; border: 0; background: rgba(255,255,255,.92); box-shadow: var(--shadow-1); display: grid; place-items: center; cursor: pointer; color: var(--ink-900); }
.nav.prev { left: 10px; } .nav.next { right: 10px; }
.thumbs { display: flex; gap: 8px; overflow-x: auto; padding-bottom: 2px; }
.thumb { flex: none; width: 64px; height: 64px; padding: 0; border-radius: 10px; overflow: hidden; border: 2px solid transparent; background: var(--mist); cursor: pointer; }
.thumb.on { border-color: var(--sea-600); }
.thumb img { width: 100%; height: 100%; object-fit: cover; }
.lightbox { position: fixed; inset: 0; z-index: 300; background: rgba(5, 14, 30, .94); display: grid; place-items: center; padding: 56px 16px; }
.lightbox img { max-width: 100%; max-height: 100%; object-fit: contain; }
.lb-close { position: absolute; top: 12px; right: 12px; }
.lb-close, .lb-nav { width: 48px; height: 48px; border-radius: 50%; border: 0; background: rgba(255,255,255,.12); color: #fff; display: grid; place-items: center; cursor: pointer; }
.lb-close:hover, .lb-nav:hover { background: rgba(255,255,255,.22); }
.lb-nav { position: absolute; top: 50%; transform: translateY(-50%); }
.lb-nav.prev { left: 12px; } .lb-nav.next { right: 12px; }
.lb-count { position: absolute; bottom: 16px; left: 50%; transform: translateX(-50%); color: rgba(255,255,255,.8); font-size: 14px; }
</style>
