<script setup>
// Lớp trang trí cho phần đầu trang: 3 vệt đèn quét chậm + đốm sáng bay lên (chủ đề chiếu sáng).
// Chỉ là trang trí: aria-hidden, không nhận chuột; tự tắt khi người dùng bật "giảm chuyển động".
const sparks = Array.from({ length: 12 }, (_, i) => ({
  left: `${(i * 37 + 11) % 100}%`,
  t: `${8 + ((i * 7) % 6)}s`,
  delay: `${(i * 1.3) % 9}s`,
  dx: `${((i % 5) - 2) * 18}px`,
  size: `${4 + (i % 3) * 2}px`,
}))
</script>

<template>
  <div class="lights" aria-hidden="true">
    <span class="beam b1" /><span class="beam b2" /><span class="beam b3" />
    <span v-for="(s, i) in sparks" :key="i" class="spark"
      :style="{ left: s.left, width: s.size, height: s.size, '--t': s.t, '--delay': s.delay, '--dx': s.dx }" />
  </div>
</template>
