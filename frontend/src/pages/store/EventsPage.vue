<script setup lang="ts">
import EventBanner from '@/components/store/EventBanner.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { CalendarClock } from 'lucide-vue-next'
import { useAsync } from '@/composables/useAsync'
import { catalogApi } from '@/api/catalog'

const events = useAsync(() => catalogApi.events(), { immediate: true })
</script>

<template>
  <div class="container page">
    <h1 class="page-title">{{ $t('catalog.eventsTitle') }}</h1>
    <ErrorState v-if="events.error.value" :error="events.error.value" @retry="events.retry" />
    <div v-else-if="events.loading.value && !events.data.value" class="list" aria-busy="true">
      <SkeletonBlock v-for="i in 2" :key="i" height="240px" radius="var(--radius-xl)" />
    </div>
    <EmptyState
      v-else-if="!events.data.value?.length"
      :title="$t('catalog.noEvents')"
      :icon="CalendarClock"
    >
      <BaseButton :to="{ name: 'products' }">{{ $t('home.browse') }}</BaseButton>
    </EmptyState>
    <div v-else class="list">
      <EventBanner v-for="e in events.data.value" :key="e.id" :event="e" hero />
    </div>
  </div>
</template>

<style scoped>
.list {
  display: grid;
  gap: var(--sp-5);
}
@media (min-width: 1024px) {
  .list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
