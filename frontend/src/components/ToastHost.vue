<script setup>
import { useToastStore } from '../stores/toast'
import AppIcon from './AppIcon.vue'
const toast = useToastStore()
</script>

<template>
  <div class="toasts" role="status" aria-live="polite">
    <TransitionGroup name="toast">
      <div v-for="t in toast.items" :key="t.id" class="toast" :class="t.type">
        <AppIcon :name="t.type === 'error' ? 'alert' : 'check'" :size="18" />
        <span>{{ t.message }}</span>
        <button class="x" aria-label="Đóng thông báo" @click="toast.dismiss(t.id)"><AppIcon name="close" :size="16" /></button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toasts { position: fixed; z-index: 200; right: 16px; bottom: 16px; left: 16px; display: flex; flex-direction: column; align-items: flex-end; gap: 8px; pointer-events: none; }
.toast { pointer-events: auto; display: flex; align-items: center; gap: 10px; max-width: 420px; padding: 12px 10px 12px 14px; border-radius: 12px; background: var(--ink-900); color: #fff; box-shadow: var(--shadow-3); font-size: 14px; }
.toast.success svg:first-child { color: #5fd3a2; }
.toast.error { background: #7f1d1d; }
.x { background: none; border: 0; color: rgba(255,255,255,.7); cursor: pointer; padding: 4px; border-radius: 6px; display: flex; }
.x:hover { color: #fff; }
.toast-enter-active, .toast-leave-active { transition: all .2s ease; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateY(8px); }
</style>
