<script setup>
import { computed } from 'vue'
const props = defineProps({ socials: { type: Object, default: () => ({}) } })
// Thứ tự & màu thương hiệu theo phương án 7.2
const defs = [
  { key: 'facebook', label: 'Facebook', color: '#1877F2', d: 'M14 8h3V4h-3a4 4 0 0 0-4 4v2H8v4h2v8h4v-8h3l1-4h-4V8z' },
  { key: 'tiktok', label: 'TikTok', color: '#000000', d: 'M16 3c.4 2.4 1.9 3.9 4 4.2v3.3a7.5 7.5 0 0 1-4-1.3v6.3A5.5 5.5 0 1 1 10.5 10v3.4a2.2 2.2 0 1 0 2.2 2.2V3H16z' },
  { key: 'zalo', label: 'Zalo', color: '#0068FF', text: 'Zalo' },
  { key: 'shopee', label: 'Shopee', color: '#EE4D2D', d: 'M6 8h12l-1 12H7L6 8zM9 8a3 3 0 0 1 6 0M10.2 16.4c.5.5 1.1.7 1.9.7 1 0 1.7-.5 1.7-1.2 0-1.6-3.5-1-3.5-2.8 0-.7.7-1.2 1.6-1.2.6 0 1.1.2 1.5.5' },
]
const items = computed(() => defs.filter((d) => props.socials?.[d.key]).map((d) => ({ ...d, url: props.socials[d.key] })))
</script>

<template>
  <ul v-if="items.length" class="socials">
    <li v-for="s in items" :key="s.key">
      <a :href="s.url" target="_blank" rel="noopener noreferrer" class="social" :style="{ '--c': s.color }" :aria-label="s.label + ' (mở tab mới)'">
        <span class="ico">
          <svg v-if="s.d" viewBox="0 0 24 24" width="20" height="20" :fill="s.key === 'shopee' ? 'none' : 'currentColor'" :stroke="s.key === 'shopee' ? 'currentColor' : 'none'" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path :d="s.d" /></svg>
          <span v-else class="zalo-word">Zalo</span>
        </span>
        <span class="lbl">{{ s.label }}</span>
      </a>
    </li>
  </ul>
</template>

<style scoped>
.socials { list-style: none; padding: 0; margin: 0; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
.social { display: flex; align-items: center; gap: 10px; min-height: 48px; padding: 6px 12px 6px 6px; border-radius: 999px; background: rgba(255,255,255,.06); border: 1px solid rgba(255,255,255,.14); color: #fff; font-weight: 600; font-size: 14px; text-decoration: none !important; transition: background-color .15s, border-color .15s; }
.social:hover { background: rgba(255,255,255,.12); border-color: rgba(255,255,255,.3); }
.ico { width: 36px; height: 36px; border-radius: 50%; display: grid; place-items: center; background: var(--c); color: #fff; flex: none; box-shadow: 0 0 0 2px rgba(255,255,255,.15); }
.zalo-word { font-size: 11px; font-weight: 800; letter-spacing: -.02em; }
</style>
