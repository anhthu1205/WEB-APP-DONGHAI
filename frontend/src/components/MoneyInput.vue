<script setup>
import { computed } from 'vue'
// Ô nhập tiền VND: hiển thị 5.990.000, giá trị là số nguyên hoặc null.
const props = defineProps({
  modelValue: { type: [Number, null], default: null },
  placeholder: { type: String, default: '0' },
  invalid: Boolean,
  id: String,
})
const emit = defineEmits(['update:modelValue'])
const display = computed(() => (props.modelValue === null || props.modelValue === undefined ? '' : new Intl.NumberFormat('vi-VN').format(props.modelValue)))
function onInput(e) {
  const digits = e.target.value.replace(/\D/g, '').slice(0, 13)
  const v = digits === '' ? null : Number(digits)
  emit('update:modelValue', v)
  e.target.value = v === null ? '' : new Intl.NumberFormat('vi-VN').format(v)
}
</script>

<template>
  <div class="money">
    <input :id="id" class="input num" :class="{ invalid }" inputmode="numeric" :value="display" :placeholder="placeholder" @input="onInput" />
    <span class="unit">₫</span>
  </div>
</template>

<style scoped>
.money { position: relative; }
.money .input { padding-right: 30px; text-align: right; }
.unit { position: absolute; right: 12px; top: 50%; transform: translateY(-50%); color: var(--ink-500); pointer-events: none; }
</style>
