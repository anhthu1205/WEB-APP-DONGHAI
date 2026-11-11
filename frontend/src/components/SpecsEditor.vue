<script setup>
import AppIcon from './AppIcon.vue'
const props = defineProps({ modelValue: { type: Array, default: () => [] } })
const emit = defineEmits(['update:modelValue'])
function update(i, key, val) {
  const next = props.modelValue.map((s, j) => (j === i ? { ...s, [key]: val } : s))
  emit('update:modelValue', next)
}
function add() { emit('update:modelValue', [...props.modelValue, { label: '', value: '' }]) }
function remove(i) { emit('update:modelValue', props.modelValue.filter((_, j) => j !== i)) }
</script>

<template>
  <div class="specs">
    <div v-for="(s, i) in modelValue" :key="i" class="spec-row">
      <input class="input" :value="s.label" placeholder="Tên thông số (VD: Công suất)" :aria-label="'Tên thông số ' + (i + 1)" @input="update(i, 'label', $event.target.value)" />
      <input class="input" :value="s.value" placeholder="Giá trị (VD: 1.500 W)" :aria-label="'Giá trị thông số ' + (i + 1)" @input="update(i, 'value', $event.target.value)" />
      <button type="button" class="btn btn-ghost btn-icon" aria-label="Xóa thông số" @click="remove(i)"><AppIcon name="trash" :size="18" /></button>
    </div>
    <button type="button" class="btn btn-ghost btn-sm" @click="add"><AppIcon name="plus" :size="16" /> Thêm thông số</button>
  </div>
</template>

<style scoped>
.specs { display: flex; flex-direction: column; gap: 8px; align-items: flex-start; }
.spec-row { display: grid; grid-template-columns: 1fr 1fr 44px; gap: 8px; width: 100%; }
@media (max-width: 560px) { .spec-row { grid-template-columns: 1fr 44px; } .spec-row .input:nth-child(2) { grid-column: 1; } .spec-row .btn { grid-row: 1 / span 2; grid-column: 2; align-self: center; } }
</style>
