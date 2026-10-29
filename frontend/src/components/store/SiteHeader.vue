<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Heart, LayoutDashboard, LogOut, Search, ShoppingCart, User } from 'lucide-vue-next'
import SearchBox from './SearchBox.vue'
import LanguageSwitch from '@/components/ui/LanguageSwitch.vue'
import BrandLogo from './BrandLogo.vue'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { useUiStore } from '@/stores/ui'

const auth = useAuthStore()
const cart = useCartStore()
const ui = useUiStore()
const router = useRouter()
const { t } = useI18n()

const mobileSearch = ref(false)
const menuOpen = ref(false)

async function logout() {
  menuOpen.value = false
  await auth.logout()
  ui.info(t('auth.signedOut'))
  void router.push({ name: 'home' })
}

function onMenuFocusOut(e: FocusEvent) {
  const next = e.relatedTarget as Node | null
  if (!next || !(e.currentTarget as HTMLElement).contains(next)) menuOpen.value = false
}
</script>

<template>
  <header class="hdr">
    <div class="container bar">
      <RouterLink :to="{ name: 'home' }" class="brand" :aria-label="$t('app.name')">
        <BrandLogo />
      </RouterLink>

      <nav class="desk-nav" :aria-label="$t('nav.mainNav')">
        <RouterLink :to="{ name: 'products' }">{{ $t('nav.products') }}</RouterLink>
        <RouterLink :to="{ name: 'categories' }">{{ $t('nav.categories') }}</RouterLink>
        <RouterLink :to="{ name: 'events' }" class="deals">{{ $t('nav.events') }}</RouterLink>
      </nav>

      <div class="search-desk"><SearchBox /></div>

      <div class="actions">
        <button
          type="button"
          class="icon-btn only-mobile"
          :aria-label="$t('common.search')"
          :aria-expanded="mobileSearch"
          aria-controls="mobile-search"
          @click="mobileSearch = !mobileSearch"
        >
          <Search :size="22" aria-hidden="true" />
        </button>

        <LanguageSwitch class="lang" compact />

        <RouterLink
          v-if="auth.isLoggedIn"
          :to="{ name: 'account-wishlist' }"
          class="icon-btn only-desk"
          :aria-label="$t('nav.wishlist')"
        >
          <Heart :size="22" aria-hidden="true" />
        </RouterLink>

        <RouterLink
          :to="{ name: 'cart' }"
          class="icon-btn cart only-desk"
          :aria-label="`${$t('nav.cart')} (${cart.count})`"
        >
          <ShoppingCart :size="22" aria-hidden="true" />
          <span v-if="cart.count" class="count" aria-hidden="true">{{
            cart.count > 99 ? '99+' : cart.count
          }}</span>
        </RouterLink>

        <div v-if="auth.isLoggedIn" class="acct only-desk" @focusout="onMenuFocusOut">
          <button
            type="button"
            class="acct-btn"
            :aria-expanded="menuOpen"
            aria-haspopup="menu"
            @click="menuOpen = !menuOpen"
            @keydown.esc="menuOpen = false"
          >
            <span class="avatar" aria-hidden="true">{{
              (auth.user?.name || auth.user?.email || '?').charAt(0).toUpperCase()
            }}</span>
            <span class="acct-name">{{ auth.user?.name || $t('nav.account') }}</span>
          </button>
          <div v-if="menuOpen" class="menu" role="menu" @keydown.esc="menuOpen = false">
            <p class="menu-email">{{ auth.user?.email }}</p>
            <RouterLink role="menuitem" :to="{ name: 'account-profile' }" @click="menuOpen = false">
              <User :size="16" aria-hidden="true" />{{ $t('nav.account') }}
            </RouterLink>
            <RouterLink role="menuitem" :to="{ name: 'account-orders' }" @click="menuOpen = false">
              <ShoppingCart :size="16" aria-hidden="true" />{{ $t('nav.orders') }}
            </RouterLink>
            <RouterLink
              v-if="auth.isStaff"
              role="menuitem"
              :to="{ name: 'admin-dashboard' }"
              @click="menuOpen = false"
            >
              <LayoutDashboard :size="16" aria-hidden="true" />{{ $t('nav.admin') }}
            </RouterLink>
            <button role="menuitem" type="button" @click="logout">
              <LogOut :size="16" aria-hidden="true" />{{ $t('nav.logout') }}
            </button>
          </div>
        </div>
        <RouterLink v-else :to="{ name: 'login' }" class="login only-desk">
          <User :size="18" aria-hidden="true" />{{ $t('nav.login') }}
        </RouterLink>
      </div>
    </div>
    <div v-if="mobileSearch" id="mobile-search" class="container mobile-search">
      <SearchBox autofocus @submitted="mobileSearch = false" />
    </div>
  </header>
</template>

<style scoped>
.hdr {
  position: sticky;
  top: 0;
  z-index: var(--z-header);
  background: rgb(255 255 255 / 0.92);
  backdrop-filter: saturate(1.6) blur(12px);
  -webkit-backdrop-filter: saturate(1.6) blur(12px);
  border-bottom: 1px solid var(--c-border);
}
.bar {
  height: var(--header-h);
  display: flex;
  align-items: center;
  gap: var(--sp-4);
}
.brand {
  display: flex;
  align-items: center;
  flex: none;
  text-decoration: none;
}
.desk-nav {
  display: none;
  gap: var(--sp-1);
}
.desk-nav a {
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  color: var(--c-text-2);
  font-weight: var(--fw-medium);
  font-size: var(--fs-sm);
  white-space: nowrap;
}
.desk-nav a:hover {
  background: var(--c-surface-2);
  text-decoration: none;
  color: var(--c-text);
}
.desk-nav a.router-link-active {
  color: var(--c-primary);
  background: var(--c-primary-50);
}
.desk-nav .deals {
  color: var(--c-accent-strong);
}
.search-desk {
  display: none;
  flex: 1;
  max-width: 440px;
}
.actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: var(--sp-1);
}
.icon-btn {
  position: relative;
  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-full);
  border: 0;
  background: transparent;
  color: var(--c-text-2);
  cursor: pointer;
}
.icon-btn:hover {
  background: var(--c-surface-2);
  color: var(--c-primary);
}
.count {
  position: absolute;
  top: 3px;
  right: 1px;
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: 999px;
  background: var(--c-accent-strong);
  color: #fff;
  font-size: 11px;
  font-weight: var(--fw-bold);
  display: grid;
  place-items: center;
  border: 2px solid #fff;
  font-variant-numeric: tabular-nums;
}
.lang {
  margin-inline: var(--sp-1);
}
.login {
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
  padding: 0 var(--sp-4);
  height: 40px;
  border-radius: var(--radius-full);
  background: var(--c-primary);
  color: #fff;
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
  margin-left: var(--sp-2);
}
.login:hover {
  background: var(--c-primary-hover);
  text-decoration: none;
}
.acct {
  position: relative;
  margin-left: var(--sp-1);
}
.acct-btn {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  border-radius: var(--radius-full);
  padding: 3px var(--sp-3) 3px 3px;
  cursor: pointer;
  max-width: 200px;
}
.acct-btn:hover {
  border-color: var(--c-primary-300);
}
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--c-primary-100);
  color: var(--c-primary);
  display: grid;
  place-items: center;
  font-weight: var(--fw-bold);
  font-size: var(--fs-sm);
  flex: none;
}
.acct-name {
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.menu {
  position: absolute;
  right: 0;
  top: calc(100% + 8px);
  min-width: 230px;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: var(--sp-2);
  display: flex;
  flex-direction: column;
}
.menu-email {
  margin: 0;
  padding: var(--sp-2) var(--sp-3);
  font-size: var(--fs-xs);
  color: var(--c-muted);
  border-bottom: 1px solid var(--c-border);
  margin-bottom: var(--sp-1);
  overflow: hidden;
  text-overflow: ellipsis;
}
.menu a,
.menu button {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  color: var(--c-text);
  font-size: var(--fs-sm);
  border: 0;
  background: transparent;
  cursor: pointer;
  text-align: left;
}
.menu a:hover,
.menu button:hover {
  background: var(--c-surface-2);
  text-decoration: none;
}
.mobile-search {
  padding-bottom: var(--sp-3);
}
.only-desk {
  display: none;
}
@media (min-width: 900px) {
  .desk-nav,
  .search-desk {
    display: flex;
  }
  .only-desk {
    display: grid;
  }
  .login.only-desk {
    display: inline-flex;
  }
  .acct.only-desk {
    display: block;
  }
  .only-mobile,
  .mobile-search {
    display: none;
  }
}
</style>
