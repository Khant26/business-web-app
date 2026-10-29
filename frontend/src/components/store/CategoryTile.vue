<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { useLocalized } from '@/composables/useLocalized'
import type { Category } from '@/types/api'

const props = defineProps<{ category: Category; showCount?: boolean }>()
const { loc } = useLocalized()
const name = computed(() => loc(props.category, 'name'))
</script>

<template>
  <RouterLink :to="{ name: 'category', params: { slug: category.slug } }" class="tile">
    <ImagePlaceholder
      :src="category.image_url"
      :name="name"
      alt=""
      ratio="1 / 1"
      rounded="50%"
      small
    />
    <span class="name">{{ name }}</span>
    <span v-if="showCount" class="count">{{
      $t('catalog.productCount', category.product_count)
    }}</span>
  </RouterLink>
</template>

<style scoped>
.tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-3) var(--sp-2);
  border-radius: var(--radius-lg);
  text-align: center;
  color: var(--c-text);
  transition: background var(--dur) var(--ease);
}
.tile:hover {
  background: var(--c-surface);
  text-decoration: none;
  box-shadow: var(--shadow-sm);
}
.tile :deep(.ph) {
  width: 72px;
  box-shadow: 0 0 0 4px var(--c-surface);
  transition: transform var(--dur) var(--ease);
}
@media (min-width: 768px) {
  .tile :deep(.ph) {
    width: 96px;
  }
}
.tile:hover :deep(.ph) {
  transform: scale(1.05);
}
.name {
  font-size: var(--fs-sm);
  font-weight: var(--fw-semibold);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.count {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
</style>
