<script setup lang="ts">
import { computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ChevronRight } from 'lucide-vue-next'
import EventBanner from '@/components/store/EventBanner.vue'
import ProductGrid from '@/components/store/ProductGrid.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { useAsync } from '@/composables/useAsync'
import { useLocalized } from '@/composables/useLocalized'
import { usePageTitle } from '@/composables/usePageTitle'
import { isApiError } from '@/api/client'
import { catalogApi } from '@/api/catalog'

const props = defineProps<{ slug: string }>()
const { loc, dateTime } = useLocalized()

const ev = useAsync((slug: string) => catalogApi.event(slug))
watch(
  () => props.slug,
  (s) => void ev.run(s),
  { immediate: true },
)

const title = computed(() => (ev.data.value ? loc(ev.data.value, 'title') : ''))
const notFound = computed(() => isApiError(ev.error.value) && ev.error.value.status === 404)
usePageTitle(title)
</script>

<template>
  <div class="container page">
    <nav class="crumbs" :aria-label="$t('nav.breadcrumb')">
      <RouterLink :to="{ name: 'home' }">{{ $t('common.home') }}</RouterLink>
      <ChevronRight :size="14" aria-hidden="true" />
      <RouterLink :to="{ name: 'events' }">{{ $t('nav.events') }}</RouterLink>
    </nav>

    <EmptyState v-if="notFound" :title="$t('errors.not_found')">
      <BaseButton :to="{ name: 'events' }">{{ $t('nav.events') }}</BaseButton>
    </EmptyState>
    <ErrorState v-else-if="ev.error.value" :error="ev.error.value" @retry="ev.retry" />
    <template v-else>
      <SkeletonBlock v-if="!ev.data.value" height="300px" radius="var(--radius-xl)" />
      <template v-else>
        <EventBanner :event="ev.data.value" :link="false" />
        <p class="ends muted">
          {{ $t('catalog.eventEnds', { date: dateTime(ev.data.value.ends_at) }) }}
        </p>
      </template>

      <section class="section" aria-labelledby="ev-products">
        <h2 id="ev-products">{{ $t('catalog.eventProducts') }}</h2>
        <ProductGrid :products="ev.data.value?.products ?? []" :loading="!ev.data.value" />
        <EmptyState
          v-if="ev.data.value && ev.data.value.products.length === 0"
          compact
          :title="$t('home.empty')"
        />
      </section>
    </template>
  </div>
</template>

<style scoped>
.crumbs {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: var(--fs-sm);
  margin-bottom: var(--sp-3);
}
.crumbs a {
  color: var(--c-muted);
}
.ends {
  margin-top: var(--sp-3);
  font-size: var(--fs-sm);
}
</style>
