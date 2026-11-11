<script setup>
import AppIcon from '../../components/AppIcon.vue'
</script>

<template>
  <div class="container nf">
    <p class="code num">404</p>
    <h1>Không tìm thấy trang</h1>
    <p class="muted">Đường dẫn có thể đã thay đổi hoặc không tồn tại.</p>
    <div class="row">
      <RouterLink to="/" class="btn btn-primary"><AppIcon name="home" /> Về trang chủ</RouterLink>
      <RouterLink to="/bao-hanh" class="btn btn-ghost">Tra cứu bảo hành</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.nf { padding: 80px 16px 20px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.code { font-size: 72px; font-weight: 800; color: var(--sea-100); line-height: 1; }
h1 { font-size: 26px; }
.row { margin-top: 12px; justify-content: center; }
</style>
