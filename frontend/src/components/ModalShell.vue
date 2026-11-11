<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import AppIcon from './AppIcon.vue'
const props = defineProps({ title: String, busy: Boolean, wide: Boolean })
const emit = defineEmits(['close'])
function onKey(e) { if (e.key === 'Escape' && !props.busy) emit('close') }
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="overlay" @click.self="!busy && emit('close')">
    <div class="box" :class="{ wide }" role="dialog" aria-modal="true" :aria-label="title">
      <header><h2>{{ title }}</h2><button class="x" aria-label="Đóng" :disabled="busy" @click="emit('close')"><AppIcon name="close" /></button></header>
      <div class="body"><slot /></div>
      <footer v-if="$slots.footer"><slot name="footer" /></footer>
    </div>
  </div>
</template>

<style scoped>
.overlay { position: fixed; inset: 0; z-index: 120; background: rgba(7,26,52,.55); display: grid; place-items: center; padding: 16px; }
.box { width: 100%; max-width: 520px; max-height: calc(100vh - 32px); overflow-y: auto; background: #fff; border-radius: var(--r-lg); box-shadow: var(--shadow-3); }
.box.wide { max-width: 720px; }
header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid var(--line); }
h2 { font-size: 18px; }
.x { width: 40px; height: 40px; border: 0; background: none; border-radius: 8px; cursor: pointer; display: grid; place-items: center; color: var(--ink-500); }
.x:hover { background: var(--mist); }
.body { padding: 20px; }
footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 1px solid var(--line); }
</style>
