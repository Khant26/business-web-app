<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Heart, LogOut, MapPin, Package, Shield, User } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'

const auth = useAuthStore()
const ui = useUiStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const links = [
  { name: 'account-profile', label: 'account.nav.profile', icon: User, match: ['account-profile'] },
  {
    name: 'account-orders',
    label: 'account.nav.orders',
    icon: Package,
    match: ['account-orders', 'account-order', 'account-pay'],
  },
  {
    name: 'account-addresses',
    label: 'account.nav.addresses',
    icon: MapPin,
    match: ['account-addresses'],
  },
  {
    name: 'account-wishlist',
    label: 'account.nav.wishlist',
    icon: Heart,
    match: ['account-wishlist'],
  },
  {
    name: 'account-security',
    label: 'account.nav.security',
    icon: Shield,
    match: ['account-security'],
  },
]

const isActive = (match: string[]) => match.includes(String(route.name))
const displayName = computed(() => auth.user?.name || auth.user?.email || '')

async function logout() {
  await auth.logout()
  ui.info(t('auth.signedOut'))
  void router.push({ name: 'home' })
}
</script>

<template>
  <div class="container page acc">
    <header class="acc-head">
      <span class="avatar" aria-hidden="true">{{ displayName.charAt(0).toUpperCase() }}</span>
      <div class="who">
        <p class="hello">
          {{
            $t('account.greeting', {
              name: auth.user?.name || auth.user?.email.split('@')[0] || $t('account.title'),
            })
          }}
        </p>
        <p class="email muted">{{ auth.user?.email }}</p>
      </div>
    </header>

    <div class="acc-grid">
      <nav class="subnav" :aria-label="$t('account.nav.label')">
        <RouterLink
          v-for="l in links"
          :key="l.name"
          :to="{ name: l.name }"
          class="item"
          :class="{ active: isActive(l.match) }"
          :aria-current="isActive(l.match) ? 'page' : undefined"
        >
          <component :is="l.icon" :size="18" aria-hidden="true" />
          <span>{{ $t(l.label) }}</span>
        </RouterLink>
        <button type="button" class="item logout" @click="logout">
          <LogOut :size="18" aria-hidden="true" />
          <span>{{ $t('account.nav.logout') }}</span>
        </button>
      </nav>

      <section class="content">
        <RouterView />
      </section>
    </div>
  </div>
</template>

<style scoped>
.acc-head {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  margin-bottom: var(--sp-5);
}
.avatar {
  flex: none;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--c-primary-500), var(--c-primary-hover));
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: var(--fw-bold);
  font-size: var(--fs-xl);
}
.who {
  min-width: 0;
}
.hello {
  margin: 0;
  font-size: var(--fs-xl);
  font-weight: var(--fw-bold);
  line-height: 1.3;
}
.email {
  margin: 0;
  font-size: var(--fs-sm);
  overflow: hidden;
  text-overflow: ellipsis;
}
.acc-grid {
  display: grid;
  gap: var(--sp-5);
}
.subnav {
  display: flex;
  gap: var(--sp-2);
  overflow-x: auto;
  scrollbar-width: none;
  margin-inline: calc(var(--sp-4) * -1);
  padding: 2px var(--sp-4) var(--sp-1);
}
.subnav::-webkit-scrollbar {
  display: none;
}
.item {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-2) var(--sp-4);
  border-radius: var(--radius-full);
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  cursor: pointer;
  white-space: nowrap;
  text-decoration: none;
}
.item:hover {
  text-decoration: none;
  border-color: var(--c-primary-300);
  color: var(--c-primary);
}
.item.active {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: #fff;
}
.logout {
  color: var(--c-danger);
}
.content {
  min-width: 0;
}
@media (min-width: 900px) {
  .acc-grid {
    grid-template-columns: 240px minmax(0, 1fr);
    align-items: start;
    gap: var(--sp-8);
  }
  .subnav {
    position: sticky;
    top: calc(var(--header-h) + var(--sp-4));
    flex-direction: column;
    margin: 0;
    padding: var(--sp-2);
    background: var(--c-surface);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-lg);
    overflow: visible;
  }
  .item {
    border: 0;
    border-radius: var(--radius-md);
    padding: var(--sp-3);
    background: transparent;
  }
  .item:hover {
    background: var(--c-surface-2);
  }
  .item.active {
    background: var(--c-primary-50);
    color: var(--c-primary);
    font-weight: var(--fw-semibold);
  }
  .logout {
    margin-top: var(--sp-2);
    border-top: 1px solid var(--c-border);
    border-radius: 0 0 var(--radius-md) var(--radius-md);
    text-align: left;
  }
}
</style>
