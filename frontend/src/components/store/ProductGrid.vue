<script setup lang="ts">
import ProductCard from './ProductCard.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import type { ProductCard as Card } from '@/types/api'

withDefaults(
  defineProps<{
    products: Card[]
    loading?: boolean
    skeletons?: number
    dense?: boolean
    scroller?: boolean
  }>(),
  { skeletons: 8 },
)
defineEmits<{ wishlistChanged: [productId: number, on: boolean] }>()
</script>

<template>
  <ul v-if="loading" class="grid" :class="{ dense, scroller }" aria-busy="true">
    <li v-for="i in skeletons" :key="i" class="sk-card">
      <SkeletonBlock height="0" radius="0" class="sk-img" />
      <div class="sk-body">
        <SkeletonBlock width="40%" height="10px" />
        <SkeletonBlock :lines="2" height="12px" />
        <SkeletonBlock width="50%" height="16px" />
      </div>
    </li>
  </ul>
  <ul v-else class="grid" :class="{ dense, scroller }">
    <li v-for="p in products" :key="p.id">
      <ProductCard :product="p" @wishlist-changed="$emit('wishlistChanged', p.id, $event)" />
    </li>
  </ul>
</template>

<style scoped>
.grid {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--sp-3);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
@media (min-width: 640px) {
  .grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--sp-4);
  }
}
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--sp-5);
  }
  .grid.dense {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 1200px) {
  .grid.dense {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
/* Horizontal snap scroller on small screens */
@media (max-width: 639px) {
  .grid.scroller {
    display: flex;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    margin-inline: calc(var(--sp-4) * -1);
    padding-inline: var(--sp-4);
    padding-bottom: var(--sp-2);
    scrollbar-width: none;
  }
  .grid.scroller::-webkit-scrollbar {
    display: none;
  }
  .grid.scroller > li {
    flex: 0 0 44%;
    scroll-snap-align: start;
  }
}
.sk-card {
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
.sk-img :deep(.sk) {
  aspect-ratio: 1 / 1;
  height: auto !important;
}
.sk-body {
  padding: var(--sp-3);
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
</style>
