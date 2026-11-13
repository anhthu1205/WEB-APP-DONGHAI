<script setup>
import { useCompanyStore } from '../stores/company'
import SiteHeader from '../components/SiteHeader.vue'
import SiteFooter from '../components/SiteFooter.vue'
import SeaDecor from '../components/SeaDecor.vue'
useCompanyStore().load()
</script>

<template>
  <div class="public">
    <SeaDecor />
    <SiteHeader />
    <main id="main"><RouterView /></main>
    <SiteFooter />
  </div>
</template>

<style scoped>
.public { min-height: 100vh; display: flex; flex-direction: column; }
main { flex: 1; }
/* Nội dung nằm trên lớp trang trí đáy biển (SeaDecor, z-index 0) */
main, .public > :deep(footer) { position: relative; z-index: 1; }
</style>
