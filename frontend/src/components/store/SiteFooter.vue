<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { Mail, Phone, QrCode, Truck } from 'lucide-vue-next'
import BrandLogo from './BrandLogo.vue'
import { useSettingsStore } from '@/stores/settings'

const settings = useSettingsStore()
const s = computed(() => settings.settings)
const year = new Date().getFullYear()
</script>

<template>
  <footer class="ftr">
    <div class="container grid">
      <div class="about">
        <BrandLogo light />
        <p>{{ $t('footer.about') }}</p>
      </div>
      <nav :aria-label="$t('footer.shop')">
        <h2>{{ $t('footer.shop') }}</h2>
        <ul>
          <li>
            <RouterLink :to="{ name: 'products' }">{{ $t('nav.products') }}</RouterLink>
          </li>
          <li>
            <RouterLink :to="{ name: 'categories' }">{{ $t('nav.categories') }}</RouterLink>
          </li>
          <li>
            <RouterLink :to="{ name: 'events' }">{{ $t('nav.events') }}</RouterLink>
          </li>
        </ul>
      </nav>
      <nav :aria-label="$t('footer.help')">
        <h2>{{ $t('footer.help') }}</h2>
        <ul>
          <li>
            <RouterLink :to="{ name: 'account-orders' }">{{ $t('nav.orders') }}</RouterLink>
          </li>
          <li>
            <RouterLink :to="{ name: 'cart' }">{{ $t('nav.cart') }}</RouterLink>
          </li>
          <li>
            <RouterLink :to="{ name: 'account-profile' }">{{ $t('nav.account') }}</RouterLink>
          </li>
        </ul>
      </nav>
      <div>
        <h2>{{ $t('footer.contact') }}</h2>
        <ul class="contact">
          <li v-if="s.support_email">
            <Mail :size="16" aria-hidden="true" /><a :href="`mailto:${s.support_email}`">{{
              s.support_email
            }}</a>
          </li>
          <li v-if="s.support_phone">
            <Phone :size="16" aria-hidden="true" /><a :href="`tel:${s.support_phone}`">{{
              s.support_phone
            }}</a>
          </li>
        </ul>
        <h2 class="mt">{{ $t('footer.payments') }}</h2>
        <div class="pay">
          <span><QrCode :size="16" aria-hidden="true" />{{ $t('footer.promptpay') }}</span>
          <span><Truck :size="16" aria-hidden="true" />{{ $t('footer.cod') }}</span>
        </div>
      </div>
    </div>
    <div class="container legal">{{ $t('footer.rights', { year, shop: s.shop_name }) }}</div>
  </footer>
</template>

<style scoped>
.ftr {
  margin-top: var(--sp-16);
  background: var(--c-primary-900);
  color: rgb(255 255 255 / 0.78);
  padding-top: var(--sp-10);
  font-size: var(--fs-sm);
}
.grid {
  display: grid;
  gap: var(--sp-8);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.about {
  grid-column: 1 / -1;
}
.about p {
  margin-top: var(--sp-3);
  max-width: 42ch;
}
@media (min-width: 900px) {
  .grid {
    grid-template-columns: 2fr 1fr 1fr 1.4fr;
  }
  .about {
    grid-column: auto;
  }
}
h2 {
  font-size: var(--fs-sm);
  color: #fff;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: var(--sp-3);
}
.mt {
  margin-top: var(--sp-5);
}
ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
a {
  color: rgb(255 255 255 / 0.85);
}
a:hover {
  color: #fff;
}
.contact li {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  word-break: break-all;
}
.pay {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sp-2);
}
.pay span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  background: rgb(255 255 255 / 0.1);
  color: #fff;
  font-size: var(--fs-xs);
}
.legal {
  margin-top: var(--sp-10);
  padding-block: var(--sp-5);
  border-top: 1px solid rgb(255 255 255 / 0.12);
  font-size: var(--fs-xs);
  color: rgb(255 255 255 / 0.6);
}
</style>
