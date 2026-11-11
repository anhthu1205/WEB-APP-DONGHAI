<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'

const props = defineProps({
  title: { type: String, required: true },
  message: { type: String, default: '' },
  confirmText: { type: String, default: 'Xác nhận' },
  danger: { type: Boolean, default: false },
  busy: { type: Boolean, default: false },
})
const emit = defineEmits(['confirm', 'cancel'])
const btn = ref(null)
function onKey(e) { if (e.key === 'Escape' && !props.busy) emit('cancel') }
onMounted(() => { document.addEventListener('keydown', onKey); btn.value?.focus() })
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="overlay" @click.self="!busy && emit('cancel')">
    <div class="dialog" role="alertdialog" aria-modal="true" :aria-label="title">
      <h2>{{ title }}</h2>
      <p v-if="message" class="msg">{{ message }}</p>
      <slot />
      <div class="actions">
        <button class="btn btn-ghost" :disabled="busy" @click="emit('cancel')">Hủy</button>
        <button ref="btn" class="btn" :class="danger ? 'btn-danger' : 'btn-primary'" :disabled="busy" @click="emit('confirm')">
          <span v-if="busy" class="spinner" />{{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay { position: fixed; inset: 0; z-index: 150; background: rgba(7, 26, 52, .55); display: grid; place-items: center; padding: 16px; }
.dialog { width: 100%; max-width: 420px; background: #fff; border-radius: var(--r-lg); padding: 22px; box-shadow: var(--shadow-3); }
h2 { font-size: 18px; }
.msg { margin-top: 8px; color: var(--ink-700); }
.actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
</style>
