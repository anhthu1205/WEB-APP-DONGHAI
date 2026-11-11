<script setup>
import { computed } from 'vue'
import AppIcon from './AppIcon.vue'

const props = defineProps({ page: Number, pageSize: Number, total: Number })
const emit = defineEmits(['change'])
const pages = computed(() => Math.max(1, Math.ceil((props.total || 0) / (props.pageSize || 20))))
const from = computed(() => (props.total ? (props.page - 1) * props.pageSize + 1 : 0))
const to = computed(() => Math.min(props.total, props.page * props.pageSize))
</script>

<template>
  <div v-if="total > 0" class="pager">
    <span class="muted small">{{ from }}–{{ to }} trong {{ total }}</span>
    <div class="row">
      <button class="btn btn-ghost btn-sm" :disabled="page <= 1" aria-label="Trang trước" @click="emit('change', page - 1)"><AppIcon name="left" :size="16" /></button>
      <span class="small num">Trang {{ page }}/{{ pages }}</span>
      <button class="btn btn-ghost btn-sm" :disabled="page >= pages" aria-label="Trang sau" @click="emit('change', page + 1)"><AppIcon name="right" :size="16" /></button>
    </div>
  </div>
</template>

<style scoped>
.pager { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 12px 16px; border-top: 1px solid var(--line); }
</style>
