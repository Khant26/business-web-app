<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { Home, LayoutGrid, ShoppingCart, User } from 'lucide-vue-next'
import { useCartStore } from '@/stores/cart'

const cart = useCartStore()
const route = useRoute()

const tabs = computed(() => [
  { to: { name: 'home' }, label: 'nav.home', icon: Home, active: route.name === 'home' },
  {
    to: { name: 'categories' },
    label: 'nav.categories',
    icon: LayoutGrid,
    active: ['categories', 'category', 'products', 'product', 'events', 'event'].includes(
      String(route.name),
    ),
  },
  {
    to: { name: 'cart' },
    label: 'nav.cart',
    icon: ShoppingCart,
    active: route.name === 'cart' || route.name === 'checkout',
    badge: cart.count,
  },
  {
    to: { name: 'account-profile' },
    label: 'nav.account',
    icon: User,
    active: String(route.name ?? '').startsWith('account') || route.name === 'login',
  },
])
</script>

<template>
  <nav class="tabbar" :aria-label="$t('nav.tabNav')">
    <RouterLink
      v-for="t in tabs"
      :key="t.label"
      :to="t.to"
      class="tab"
      :class="{ active: t.active }"
      :aria-current="t.active ? 'page' : undefined"
    >
      <span class="ic">
        <component
          :is="t.icon"
          :size="22"
          :stroke-width="t.active ? 2.4 : 1.8"
          aria-hidden="true"
        />
        <span v-if="t.badge" class="badge" aria-hidden="true">{{
          t.badge > 99 ? '99+' : t.badge
        }}</span>
      </span>
      <span class="lbl">
        {{ $t(t.label) }}
        <span v-if="t.badge" class="sr-only">({{ t.badge }})</span>
      </span>
    </RouterLink>
  </nav>
</template>

<style scoped>
.tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: var(--z-tabbar);
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  background: rgb(255 255 255 / 0.96);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-top: 1px solid var(--c-border);
  height: calc(var(--tabbar-h) + env(safe-area-inset-bottom));
  padding-bottom: env(safe-area-inset-bottom);
}
.tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  color: var(--c-muted);
  font-size: 11px;
  font-weight: var(--fw-medium);
  text-decoration: none;
  line-height: 1.2;
}
.tab:hover {
  text-decoration: none;
}
.tab.active {
  color: var(--c-primary);
}
.ic {
  position: relative;
  display: grid;
  place-items: center;
  width: 48px;
  height: 28px;
  border-radius: var(--radius-full);
  transition: background var(--dur) var(--ease);
}
.active .ic {
  background: var(--c-primary-50);
}
.badge {
  position: absolute;
  top: -4px;
  right: 4px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  border-radius: 999px;
  background: var(--c-accent-strong);
  color: #fff;
  font-size: 10px;
  font-weight: var(--fw-bold);
  display: grid;
  place-items: center;
  border: 2px solid #fff;
}
.lbl {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-inline: 2px;
}
@media (min-width: 900px) {
  .tabbar {
    display: none;
  }
}
</style>
