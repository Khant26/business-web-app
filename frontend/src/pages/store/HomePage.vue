<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { ArrowRight, ShieldCheck, Truck, QrCode } from 'lucide-vue-next'
import HeroCarousel from '@/components/store/HeroCarousel.vue'
import CategoryTile from '@/components/store/CategoryTile.vue'
import ProductGrid from '@/components/store/ProductGrid.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { useAsync } from '@/composables/useAsync'
import { catalogApi } from '@/api/catalog'
import type { ProductSort } from '@/types/api'

const home = useAsync(() => catalogApi.home(), { immediate: true })
const { data, loading, error, retry } = home

const blocks: {
  key: 'popular' | 'featured' | 'new_arrivals'
  title: string
  sort?: ProductSort
}[] = [
  { key: 'popular', title: 'home.popular', sort: 'popular' },
  { key: 'featured', title: 'home.featured' },
  { key: 'new_arrivals', title: 'home.newArrivals', sort: 'newest' },
]
</script>

<template>
  <div class="container page home">
    <ErrorState v-if="error && !data" :error="error" @retry="retry" />

    <template v-else>
      <SkeletonBlock v-if="loading && !data" height="300px" radius="var(--radius-xl)" />
      <HeroCarousel v-else-if="data?.events.length" :events="data.events" />
      <section v-else class="welcome">
        <h1>{{ $t('home.welcomeTitle') }}</h1>
        <p>{{ $t('home.welcomeText') }}</p>
        <BaseButton :to="{ name: 'products' }" size="lg" variant="secondary">
          {{ $t('home.browse') }}<ArrowRight :size="18" aria-hidden="true" />
        </BaseButton>
      </section>

      <ul class="perks">
        <li><Truck :size="20" aria-hidden="true" />{{ $t('home.perk1') }}</li>
        <li><QrCode :size="20" aria-hidden="true" />{{ $t('home.perk2') }}</li>
        <li><ShieldCheck :size="20" aria-hidden="true" />{{ $t('home.perk3') }}</li>
      </ul>

      <section v-if="loading || data?.categories.length" class="section" aria-labelledby="h-cats">
        <div class="section-head">
          <h2 id="h-cats">{{ $t('home.shopByCategory') }}</h2>
          <RouterLink :to="{ name: 'categories' }" class="see-all">
            {{ $t('common.seeAll') }}<ArrowRight :size="16" aria-hidden="true" />
          </RouterLink>
        </div>
        <ul v-if="loading && !data" class="cats" aria-busy="true">
          <li v-for="i in 6" :key="i" class="cat-sk">
            <SkeletonBlock width="72px" height="72px" circle />
            <SkeletonBlock width="60px" height="10px" />
          </li>
        </ul>
        <ul v-else class="cats">
          <li v-for="c in data?.categories ?? []" :key="c.id">
            <CategoryTile :category="c" />
          </li>
        </ul>
      </section>

      <template v-for="block in blocks" :key="block.key">
        <section
          v-if="(loading && !data) || (data?.[block.key].length ?? 0) > 0"
          class="section"
          :aria-labelledby="`h-${block.key}`"
        >
          <div class="section-head">
            <h2 :id="`h-${block.key}`">{{ $t(block.title) }}</h2>
            <RouterLink
              :to="{ name: 'products', query: block.sort ? { sort: block.sort } : {} }"
              class="see-all"
            >
              {{ $t('common.seeAll') }}<ArrowRight :size="16" aria-hidden="true" />
            </RouterLink>
          </div>
          <ProductGrid
            :products="data?.[block.key] ?? []"
            :loading="loading && !data"
            :skeletons="4"
            scroller
          />
        </section>
      </template>
    </template>
  </div>
</template>

<style scoped>
.home {
  padding-top: var(--sp-4);
}
@media (min-width: 768px) {
  .home {
    padding-top: var(--sp-6);
  }
}
.welcome {
  border-radius: var(--radius-xl);
  padding: var(--sp-8) var(--sp-5);
  background:
    radial-gradient(circle at 90% 10%, rgb(251 146 60 / 0.55) 0%, transparent 40%),
    linear-gradient(135deg, var(--c-primary-800), var(--c-primary) 60%, var(--c-primary-600));
  color: #fff;
}
@media (min-width: 768px) {
  .welcome {
    padding: var(--sp-16) var(--sp-12);
  }
}
.welcome h1 {
  font-size: var(--fs-2xl);
  font-weight: 800;
  max-width: 18ch;
}
@media (min-width: 768px) {
  .welcome h1 {
    font-size: 2.75rem;
  }
}
.welcome p {
  max-width: 48ch;
  color: rgb(255 255 255 / 0.88);
  margin-bottom: var(--sp-5);
}
.perks {
  list-style: none;
  padding: 0;
  margin: var(--sp-4) 0 0;
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--sp-2);
}
@media (min-width: 640px) {
  .perks {
    grid-template-columns: repeat(3, 1fr);
    gap: var(--sp-3);
  }
}
.perks li {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-3) var(--sp-4);
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-lg);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
}
.perks :deep(svg) {
  color: var(--c-primary);
  flex: none;
}
.see-all {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: var(--fs-sm);
  font-weight: var(--fw-semibold);
  white-space: nowrap;
}
.cats {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--sp-1);
}
@media (min-width: 640px) {
  .cats {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
@media (min-width: 1024px) {
  .cats {
    grid-template-columns: repeat(8, minmax(0, 1fr));
  }
}
.cat-sk {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-3);
}
</style>
