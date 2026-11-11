<script setup>
import { useCompanyStore } from '../stores/company'
import SiteHeader from '../components/SiteHeader.vue'
import SiteFooter from '../components/SiteFooter.vue'
useCompanyStore().load()
</script>

<template>
  <div class="public">
    <SiteHeader />
    <main id="main"><RouterView /></main>
    <SiteFooter />
  </div>
</template>

<style scoped>
.public { min-height: 100vh; display: flex; flex-direction: column; }
main { flex: 1; }
</style>
