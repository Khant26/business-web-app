<script setup lang="ts">
import CategoryTile from '@/components/store/CategoryTile.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { useAsync } from '@/composables/useAsync'
import { catalogApi } from '@/api/catalog'

const cats = useAsync(() => catalogApi.categories(), { immediate: true })
</script>

<template>
  <div class="container page">
    <h1 class="page-title">{{ $t('catalog.categoriesTitle') }}</h1>
    <ErrorState v-if="cats.error.value" :error="cats.error.value" @retry="cats.retry" />
    <ul v-else-if="cats.loading.value && !cats.data.value" class="grid" aria-busy="true">
      <li v-for="i in 8" :key="i" class="sk card">
        <SkeletonBlock width="96px" height="96px" circle />
        <SkeletonBlock width="80px" height="12px" />
      </li>
    </ul>
    <EmptyState v-else-if="!cats.data.value?.length" :title="$t('home.empty')">
      <BaseButton :to="{ name: 'products' }">{{ $t('home.browse') }}</BaseButton>
    </EmptyState>
    <ul v-else class="grid">
      <li v-for="c in cats.data.value" :key="c.id" class="card">
        <CategoryTile :category="c" show-count />
      </li>
    </ul>
  </div>
</template>

<style scoped>
.grid {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--sp-3);
}
@media (min-width: 640px) {
  .grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: var(--sp-4);
  }
}
.grid > li :deep(.tile) {
  padding: var(--sp-5) var(--sp-3);
  height: 100%;
}
.sk {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-5);
}
</style>
