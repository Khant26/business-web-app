<script setup lang="ts">
import { computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ChevronRight } from 'lucide-vue-next'
import ProductBrowser from '@/components/store/ProductBrowser.vue'
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
const { loc } = useLocalized()
const cat = useAsync((slug: string) => catalogApi.category(slug))
watch(
  () => props.slug,
  (s) => void cat.run(s),
  { immediate: true },
)

const name = computed(() => (cat.data.value ? loc(cat.data.value, 'name') : ''))
const notFound = computed(() => isApiError(cat.error.value) && cat.error.value.status === 404)
usePageTitle(name)
</script>

<template>
  <div class="container page">
    <nav class="crumbs" :aria-label="$t('nav.breadcrumb')">
      <RouterLink :to="{ name: 'home' }">{{ $t('common.home') }}</RouterLink>
      <ChevronRight :size="14" aria-hidden="true" />
      <RouterLink :to="{ name: 'categories' }">{{ $t('nav.categories') }}</RouterLink>
      <template v-if="name">
        <ChevronRight :size="14" aria-hidden="true" />
        <span aria-current="page">{{ name }}</span>
      </template>
    </nav>

    <EmptyState v-if="notFound" :title="$t('errors.not_found')">
      <BaseButton :to="{ name: 'categories' }">{{ $t('nav.categories') }}</BaseButton>
    </EmptyState>
    <ErrorState v-else-if="cat.error.value" :error="cat.error.value" @retry="cat.retry" />
    <template v-else>
      <h1 class="page-title">
        <SkeletonBlock v-if="!cat.data.value" width="220px" height="32px" />
        <template v-else>{{ name }}</template>
      </h1>
      <ProductBrowser :fixed-category="slug" />
    </template>
  </div>
</template>

<style scoped>
.crumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  font-size: var(--fs-sm);
  color: var(--c-muted);
  margin-bottom: var(--sp-3);
}
.crumbs a {
  color: var(--c-muted);
}
.crumbs span {
  color: var(--c-text);
  font-weight: var(--fw-medium);
}
</style>
