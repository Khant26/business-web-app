<script setup lang="ts">
import SiteHeader from '@/components/store/SiteHeader.vue'
import SiteFooter from '@/components/store/SiteFooter.vue'
import MobileTabBar from '@/components/store/MobileTabBar.vue'
</script>

<template>
  <div class="store">
    <a href="#main" class="skip-link">{{ $t('app.skip') }}</a>
    <SiteHeader />
    <main id="main" tabindex="-1" class="main">
      <RouterView v-slot="{ Component, route }">
        <Transition name="fade" mode="out-in">
          <component
            :is="Component"
            :key="route.matched[1]?.path === '/account' ? 'account' : route.path"
          />
        </Transition>
      </RouterView>
    </main>
    <SiteFooter />
    <MobileTabBar />
  </div>
</template>

<style scoped>
.store {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  padding-bottom: calc(var(--tabbar-h) + env(safe-area-inset-bottom));
}
@media (min-width: 900px) {
  .store {
    padding-bottom: 0;
  }
}
.main {
  flex: 1;
  outline: none;
}
</style>
