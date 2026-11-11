<script setup>
// Ô nhập SĐT: inputmode tel, chỉ cho chữ số và + . - ( ) khoảng trắng.
defineProps({ modelValue: { type: String, default: '' }, invalid: Boolean, id: String, placeholder: { type: String, default: 'VD: 0912 345 678' }, autofocus: Boolean })
const emit = defineEmits(['update:modelValue'])
function onInput(e) {
  const v = e.target.value.replace(/[^\d+\s.\-()]/g, '').slice(0, 20)
  e.target.value = v
  emit('update:modelValue', v)
}
</script>

<template>
  <input :id="id" class="input num" :class="{ invalid }" type="tel" inputmode="tel" autocomplete="tel"
         :placeholder="placeholder" :value="modelValue" :autofocus="autofocus" @input="onInput" />
</template>
