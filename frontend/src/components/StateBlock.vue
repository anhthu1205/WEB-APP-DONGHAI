<script setup>
import AppIcon from './AppIcon.vue'
defineProps({
  kind: { type: String, default: 'empty' }, // empty | error | loading
  title: { type: String, default: '' },
  text: { type: String, default: '' },
})
defineEmits(['retry'])
</script>

<template>
  <div class="state">
    <span v-if="kind === 'loading'" class="spinner" style="width:28px;height:28px;color:var(--sea-600)" />
    <AppIcon v-else :name="kind === 'error' ? 'alert' : 'box'" />
    <h3 v-if="title">{{ title }}</h3>
    <p v-if="text">{{ text }}</p>
    <slot />
    <button v-if="kind === 'error'" class="btn btn-ghost btn-sm" @click="$emit('retry')">Thử lại</button>
  </div>
</template>
