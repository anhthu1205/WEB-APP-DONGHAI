<script setup>
import { computed } from 'vue'
import { useCompanyStore } from '../stores/company'
import { telHref } from '../utils/format'
import AppIcon from './AppIcon.vue'
import SocialButtons from './SocialButtons.vue'

const company = useCompanyStore()
const c = computed(() => company.data)
const year = new Date().getFullYear()
</script>

<template>
  <footer id="lien-he" class="ftr">
    <svg class="wave" viewBox="0 0 1440 60" preserveAspectRatio="none" aria-hidden="true"><path d="M0 30c120-20 240-20 360 0s240 20 360 0 240-20 360 0 240 20 360 0v30H0z" /></svg>
    <div class="container grid">
      <section class="about">
        <div class="brand"><img src="/favicon.svg" alt="" width="40" height="40" /><h2>{{ c?.name || 'Công ty Đông Hải' }}</h2></div>
        <p v-if="c?.slogan" class="slogan">{{ c.slogan }}</p>
        <p v-if="c?.about" class="txt">{{ c.about }}</p>
      </section>
      <section>
        <h3>Thông tin liên hệ</h3>
        <ul class="info">
          <li v-if="c?.address"><AppIcon name="map" :size="18" /><span>{{ c.address }}</span></li>
          <li v-if="c?.phone"><AppIcon name="phone" :size="18" /><a :href="telHref(c.phone)" class="num">{{ c.phone }}</a></li>
          <li v-if="c?.email"><AppIcon name="mail" :size="18" /><a :href="'mailto:' + c.email">{{ c.email }}</a></li>
          <li v-if="c?.working_hours"><AppIcon name="clock" :size="18" /><span>{{ c.working_hours }}</span></li>
        </ul>
      </section>
      <section>
        <h3>Kết nối với chúng tôi</h3>
        <SocialButtons :socials="c?.socials" />
      </section>
    </div>
    <div class="container bottom">
      <span>© {{ year }} {{ c?.name || 'Công ty Đông Hải' }}</span>
      <RouterLink to="/dang-nhap" class="staff">Đăng nhập nội bộ</RouterLink>
    </div>
  </footer>
</template>

<style scoped>
.ftr { position: relative; margin-top: 64px; background: var(--sea-900); color: rgba(255,255,255,.78); padding-top: 56px; }
.wave { position: absolute; top: -59px; left: 0; width: 100%; height: 60px; fill: var(--sea-900); }
.grid { display: grid; gap: 36px; padding-bottom: 40px; }
@media (min-width: 900px) { .grid { grid-template-columns: 1.3fr 1fr 1fr; gap: 48px; } }
.brand { display: flex; align-items: center; gap: 12px; }
h2 { color: #fff; font-size: 19px; }
h3 { color: #fff; font-size: 15px; font-weight: 700; margin-bottom: 14px; }
.slogan { margin-top: 12px; color: var(--amber-500); font-weight: 600; }
.txt { margin-top: 10px; font-size: 14px; line-height: 1.7; max-width: 52ch; }
.info { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 12px; font-size: 14px; }
.info li { display: flex; gap: 10px; align-items: flex-start; }
.info svg { flex: none; margin-top: 2px; color: var(--amber-500); }
.info a { color: #fff; }
.bottom { display: flex; justify-content: space-between; gap: 12px; flex-wrap: wrap; padding-top: 18px; padding-bottom: 22px; font-size: 13px;
  background: linear-gradient(rgba(255,255,255,.12), rgba(255,255,255,.12)) top center / calc(100% - 2 * var(--gutter)) 1px no-repeat; }
.staff { color: rgba(255,255,255,.6); }
.staff:hover { color: #fff; }
</style>
