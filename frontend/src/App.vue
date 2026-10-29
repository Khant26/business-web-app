<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import ToastHost from '@/components/ui/ToastHost.vue'
import { onUnauthorized } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useWishlistStore } from '@/stores/wishlist'
import { useSettingsStore } from '@/stores/settings'
import { useUiStore } from '@/stores/ui'

const router = useRouter()
const { t } = useI18n()
const auth = useAuthStore()
const wishlist = useWishlistStore()
const ui = useUiStore()
void useSettingsStore().load()

// Session ended server-side (expired/revoked): clear local user and send to login if needed.
const off = onUnauthorized(() => {
  const wasLoggedIn = auth.isLoggedIn
  auth.clear()
  const route = router.currentRoute.value
  if (route.matched.some((r) => r.meta.requiresAuth)) {
    if (wasLoggedIn) ui.warning(t('auth.sessionExpired'))
    void router.replace({ name: 'login', query: { redirect: route.fullPath } })
  }
})
onBeforeUnmount(off)

watch(
  () => auth.user?.id,
  (id) => {
    if (id) void wishlist.load()
    else wishlist.reset()
  },
  { immediate: true },
)
</script>

<template>
  <RouterView />
  <ToastHost />
</template>
