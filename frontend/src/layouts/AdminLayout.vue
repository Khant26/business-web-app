<script setup lang="ts">
import { computed, ref, watch, type Component } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ClipboardList,
  FolderTree,
  Gauge,
  History,
  LogOut,
  Menu,
  Package,
  Settings,
  Star,
  Store,
  Tag,
  Ticket,
  Truck,
  Users,
  Warehouse,
} from 'lucide-vue-next'
import BrandLogo from '@/components/store/BrandLogo.vue'
import LanguageSwitch from '@/components/ui/LanguageSwitch.vue'
import BaseDrawer from '@/components/ui/BaseDrawer.vue'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import '@/components/admin/admin.css'

interface NavItem {
  name: string
  label: string
  icon: Component
  /** Other route names that belong to this section. */
  match?: string[]
  adminOnly?: boolean
}
interface NavGroup {
  label: string
  items: NavItem[]
}

const auth = useAuthStore()
const ui = useUiStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const drawer = ref(false)

const ALL_GROUPS: NavGroup[] = [
  {
    label: 'admin.nav.sales',
    items: [
      { name: 'admin-dashboard', label: 'admin.nav.dashboard', icon: Gauge },
      {
        name: 'admin-orders',
        label: 'admin.nav.orders',
        icon: ClipboardList,
        match: ['admin-order'],
      },
    ],
  },
  {
    label: 'admin.nav.catalog',
    items: [
      {
        name: 'admin-products',
        label: 'admin.nav.products',
        icon: Package,
        match: ['admin-product-new', 'admin-product-edit'],
      },
      { name: 'admin-inventory', label: 'admin.nav.inventory', icon: Warehouse },
      { name: 'admin-categories', label: 'admin.nav.categories', icon: FolderTree },
    ],
  },
  {
    label: 'admin.nav.marketing',
    items: [
      { name: 'admin-events', label: 'admin.nav.events', icon: Tag },
      { name: 'admin-coupons', label: 'admin.nav.coupons', icon: Ticket },
      { name: 'admin-reviews', label: 'admin.nav.reviews', icon: Star },
    ],
  },
  {
    label: 'admin.nav.administration',
    items: [
      { name: 'admin-zones', label: 'admin.nav.zones', icon: Truck, adminOnly: true },
      { name: 'admin-settings', label: 'admin.nav.settings', icon: Settings, adminOnly: true },
      { name: 'admin-users', label: 'admin.nav.users', icon: Users, adminOnly: true },
      { name: 'admin-audit', label: 'admin.nav.audit', icon: History, adminOnly: true },
    ],
  },
]

const groups = computed<NavGroup[]>(() =>
  ALL_GROUPS.map((g) => ({
    ...g,
    items: g.items.filter((i) => !i.adminOnly || auth.isAdmin),
  })).filter((g) => g.items.length > 0),
)

function isActive(item: NavItem): boolean {
  const current = String(route.name ?? '')
  return current === item.name || (item.match?.includes(current) ?? false)
}

const pageTitle = computed(() => {
  const key = [...route.matched].reverse().find((r) => r.meta.title)?.meta.title
  return key ? t(key) : t('admin.title')
})

watch(
  () => route.fullPath,
  () => (drawer.value = false),
)

async function logout() {
  await auth.logout()
  ui.info(t('auth.signedOut'))
  void router.push({ name: 'home' })
}
</script>

<template>
  <div class="admin">
    <a href="#admin-main" class="skip-link">{{ $t('app.skip') }}</a>

    <aside class="sidebar">
      <RouterLink :to="{ name: 'admin-dashboard' }" class="brand">
        <BrandLogo light />
        <span class="tag">{{ $t('admin.title') }}</span>
      </RouterLink>
      <nav :aria-label="$t('admin.nav.adminNav')" class="nav">
        <div v-for="g in groups" :key="g.label" class="group">
          <p class="group-label">{{ $t(g.label) }}</p>
          <RouterLink
            v-for="item in g.items"
            :key="item.name"
            :to="{ name: item.name }"
            class="link"
            :class="{ active: isActive(item) }"
            :aria-current="isActive(item) ? 'page' : undefined"
          >
            <component :is="item.icon" :size="18" aria-hidden="true" />
            <span>{{ $t(item.label) }}</span>
          </RouterLink>
        </div>
      </nav>
      <RouterLink :to="{ name: 'home' }" class="link store-link">
        <Store :size="18" aria-hidden="true" /><span>{{ $t('admin.nav.viewStore') }}</span>
      </RouterLink>
    </aside>

    <BaseDrawer v-model:open="drawer" side="left">
      <template #header>
        <span class="drawer-brand"><BrandLogo /></span>
      </template>
      <nav :aria-label="$t('admin.nav.adminNav')" class="nav light">
        <div v-for="g in groups" :key="g.label" class="group">
          <p class="group-label">{{ $t(g.label) }}</p>
          <RouterLink
            v-for="item in g.items"
            :key="item.name"
            :to="{ name: item.name }"
            class="link"
            :class="{ active: isActive(item) }"
            :aria-current="isActive(item) ? 'page' : undefined"
          >
            <component :is="item.icon" :size="18" aria-hidden="true" />
            <span>{{ $t(item.label) }}</span>
          </RouterLink>
        </div>
        <RouterLink :to="{ name: 'home' }" class="link">
          <Store :size="18" aria-hidden="true" /><span>{{ $t('admin.nav.viewStore') }}</span>
        </RouterLink>
      </nav>
    </BaseDrawer>

    <div class="main-col">
      <header class="topbar">
        <button
          type="button"
          class="menu-btn"
          :aria-label="$t('common.openMenu')"
          :aria-expanded="drawer"
          @click="drawer = true"
        >
          <Menu :size="22" aria-hidden="true" />
        </button>
        <p class="page-title">{{ pageTitle }}</p>
        <div class="top-actions">
          <LanguageSwitch compact />
          <RouterLink
            :to="{ name: 'home' }"
            class="store-btn"
            :aria-label="$t('admin.nav.viewStore')"
          >
            <Store :size="16" aria-hidden="true" />
            <span class="hide-sm">{{ $t('admin.nav.viewStore') }}</span>
          </RouterLink>
          <div class="user">
            <span class="avatar" aria-hidden="true">
              {{ (auth.user?.name || auth.user?.email || '?').charAt(0).toUpperCase() }}
            </span>
            <span class="who hide-sm">
              <span class="nm">{{ auth.user?.name || auth.user?.email }}</span>
              <span class="role">{{ $t(`admin.roles.${auth.user?.role ?? 'staff'}`) }}</span>
            </span>
          </div>
          <button
            type="button"
            class="icon-btn"
            :aria-label="$t('admin.nav.logout')"
            :title="$t('admin.nav.logout')"
            @click="logout"
          >
            <LogOut :size="18" aria-hidden="true" />
          </button>
        </div>
      </header>
      <main id="admin-main" class="content" tabindex="-1">
        <RouterView v-slot="{ Component: view, route: r }">
          <component :is="view" :key="r.path" />
        </RouterView>
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin {
  min-height: 100dvh;
  background: var(--c-bg);
}
.sidebar {
  display: none;
}
@media (min-width: 1024px) {
  .sidebar {
    display: flex;
    flex-direction: column;
    position: fixed;
    inset: 0 auto 0 0;
    width: var(--admin-sidebar);
    background: var(--c-primary-900);
    color: #fff;
    padding: var(--sp-4) var(--sp-3);
    overflow-y: auto;
    z-index: var(--z-header);
  }
  .main-col {
    margin-left: var(--admin-sidebar);
  }
  .menu-btn {
    display: none !important;
  }
}
.brand {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-2) var(--sp-2) var(--sp-5);
  text-decoration: none;
}
.brand:hover {
  text-decoration: none;
}
.tag {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  background: rgb(255 255 255 / 0.14);
  color: #fff;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-weight: var(--fw-bold);
}
.nav {
  display: flex;
  flex-direction: column;
  gap: var(--sp-4);
  flex: 1;
}
.group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.group-label {
  margin: 0 0 var(--sp-1);
  padding: 0 var(--sp-3);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgb(255 255 255 / 0.6);
  font-weight: var(--fw-semibold);
}
.link {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  color: rgb(255 255 255 / 0.82);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  text-decoration: none;
  min-height: 40px;
}
.link:hover {
  background: rgb(255 255 255 / 0.08);
  color: #fff;
  text-decoration: none;
}
.link.active {
  background: rgb(255 255 255 / 0.14);
  color: #fff;
  box-shadow: inset 3px 0 0 var(--c-accent);
}
.link:focus-visible {
  outline: 2px solid #fff;
  outline-offset: -2px;
}
.store-link {
  margin-top: var(--sp-4);
  border-top: 1px solid rgb(255 255 255 / 0.12);
  border-radius: 0;
  padding-top: var(--sp-4);
}
.nav.light .group-label {
  color: var(--c-muted);
}
.nav.light .link {
  color: var(--c-text-2);
}
.nav.light .link:hover {
  background: var(--c-surface-2);
  color: var(--c-text);
}
.nav.light .link.active {
  background: var(--c-primary-50);
  color: var(--c-primary);
  box-shadow: inset 3px 0 0 var(--c-primary);
}
.nav.light .link:focus-visible {
  outline-color: var(--c-primary-600);
}
.drawer-brand {
  display: flex;
}
.topbar {
  position: sticky;
  top: 0;
  z-index: calc(var(--z-header) - 1);
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  height: var(--header-h);
  padding: 0 var(--sp-3);
  background: rgb(255 255 255 / 0.94);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--c-border);
}
@media (min-width: 768px) {
  .topbar {
    padding: 0 var(--sp-6);
  }
}
.menu-btn,
.icon-btn {
  width: 40px;
  height: 40px;
  border: 0;
  background: transparent;
  border-radius: var(--radius-md);
  display: grid;
  place-items: center;
  cursor: pointer;
  color: var(--c-text-2);
  flex: none;
}
.menu-btn:hover,
.icon-btn:hover {
  background: var(--c-surface-2);
  color: var(--c-primary);
}
.page-title {
  margin: 0;
  font-weight: var(--fw-semibold);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
.top-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  flex: none;
}
.store-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 var(--sp-3);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-full);
  font-size: var(--fs-sm);
  color: var(--c-text-2);
  font-weight: var(--fw-medium);
}
.store-btn:hover {
  border-color: var(--c-primary-300);
  color: var(--c-primary);
  text-decoration: none;
}
.user {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
}
.avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: var(--c-primary-100);
  color: var(--c-primary);
  font-weight: var(--fw-bold);
  display: grid;
  place-items: center;
  flex: none;
}
.who {
  display: flex;
  flex-direction: column;
  line-height: 1.25;
  max-width: 180px;
}
.nm {
  font-size: var(--fs-sm);
  font-weight: var(--fw-semibold);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.role {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
@media (max-width: 767px) {
  .hide-sm {
    display: none !important;
  }
  .store-btn {
    width: 36px;
    padding: 0;
    justify-content: center;
  }
  .user {
    display: none;
  }
}
.content {
  padding: var(--sp-5) var(--sp-4) var(--sp-12);
  max-width: 1400px;
  outline: none;
}
@media (min-width: 768px) {
  .content {
    padding: var(--sp-6) var(--sp-8) var(--sp-12);
  }
}
</style>
