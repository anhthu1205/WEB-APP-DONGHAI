<script setup>
import { computed } from 'vue'
// Ảnh thay thế khi SP chưa có ảnh: nền nhạt theo id + nét vẽ thiết bị + mã SP.
const props = defineProps({ seed: { type: [Number, String], default: 0 }, label: { type: String, default: '' } })
const hues = [214, 199, 228, 190, 205, 220, 236, 184]
const hue = computed(() => hues[Math.abs(Number(props.seed) || 0) % hues.length])
</script>

<template>
  <div class="ph" :style="{ '--h': hue }" role="img" :aria-label="'Chưa có ảnh ' + label">
    <svg viewBox="0 0 120 120" aria-hidden="true">
      <g fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
        <rect x="38" y="22" width="44" height="70" rx="7" />
        <path d="M38 42h44M48 32h7M70 32h4" />
        <circle cx="60" cy="66" r="13" />
        <path d="M53 66c3-3.5 11-3.5 14 0M45 92v5M75 92v5" />
      </g>
    </svg>
    <span v-if="label" class="code">{{ label }}</span>
  </div>
</template>

<style scoped>
.ph { width: 100%; height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 4px;
  background: radial-gradient(120% 90% at 30% 15%, hsl(var(--h) 90% 97%), hsl(var(--h) 70% 91%)); color: hsl(var(--h) 45% 62%); }
svg { width: 46%; height: auto; }
.code { font-size: 12px; font-weight: 700; color: hsl(var(--h) 40% 45%); letter-spacing: .02em; }
</style>
