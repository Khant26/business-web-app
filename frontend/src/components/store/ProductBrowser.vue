<script setup lang="ts">
/**
 * Product listing with search / category / sort / price / stock filters, all
 * synced to the URL query (?q=&category=&sort=&min=&max=&in_stock=1&page=).
 * Prices in the URL are whole baht; the API receives satang.
 */
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { SlidersHorizontal, X } from 'lucide-vue-next'
import ProductGrid from './ProductGrid.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseDrawer from '@/components/ui/BaseDrawer.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { useAsync } from '@/composables/useAsync'
import { useLocalized } from '@/composables/useLocalized'
import { catalogApi } from '@/api/catalog'
import type { ProductQuery, ProductSort } from '@/types/api'
import type { SelectOption } from '@/types/ui'

const props = defineProps<{ fixedCategory?: string }>()

const PAGE_SIZE = 24
const SORTS: ProductSort[] = ['popular', 'newest', 'price_asc', 'price_desc', 'rating']

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { loc } = useLocalized()

const str = (v: LocationQuery[string] | undefined) => (typeof v === 'string' ? v : '')
const num = (v: LocationQuery[string] | undefined) => {
  const n = Number(str(v))
  return str(v) !== '' && Number.isFinite(n) && n >= 0 ? n : undefined
}

const state = computed(() => {
  const q = route.query
  const sort = str(q.sort) as ProductSort
  return {
    q: str(q.q).trim(),
    category: props.fixedCategory ?? str(q.category),
    sort: SORTS.includes(sort) ? sort : 'popular',
    min: num(q.min),
    max: num(q.max),
    inStock: q.in_stock === '1',
    page: Math.max(1, Number(str(q.page)) || 1),
  }
})

const list = useAsync((query: ProductQuery) => catalogApi.products(query))
const categories = useAsync(() => catalogApi.categories(), { immediate: !props.fixedCategory })

function load() {
  const s = state.value
  void list.run({
    q: s.q || undefined,
    category: s.category || undefined,
    sort: s.sort,
    min_price: s.min !== undefined ? Math.round(s.min * 100) : undefined,
    max_price: s.max !== undefined ? Math.round(s.max * 100) : undefined,
    in_stock: s.inStock || undefined,
    page: s.page,
    page_size: PAGE_SIZE,
  })
}
watch(() => JSON.stringify(state.value), load, { immediate: true })

const items = computed(() => list.data.value?.items ?? [])
const total = computed(() => list.data.value?.total ?? 0)
const pages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

function update(patch: Record<string, string | undefined>, keepPage = false) {
  const query: Record<string, string> = {}
  for (const [k, v] of Object.entries(route.query)) if (typeof v === 'string') query[k] = v
  for (const [k, v] of Object.entries(patch)) {
    if (v === undefined || v === '') delete query[k]
    else query[k] = v
  }
  if (!keepPage) delete query.page
  void router.replace({ query })
}

const sortModel = computed<ProductSort>({
  get: () => state.value.sort,
  set: (v) => update({ sort: v === 'popular' ? undefined : v }),
})
const categoryModel = computed<string>({
  get: () => state.value.category,
  set: (v) => update({ category: v || undefined }),
})
const inStockModel = computed<boolean>({
  get: () => state.value.inStock,
  set: (v) => update({ in_stock: v ? '1' : undefined }),
})
const pageModel = computed<number>({
  get: () => state.value.page,
  set: (p) => {
    update({ page: p > 1 ? String(p) : undefined }, true)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  },
})

const minInput = ref<string>('')
const maxInput = ref<string>('')
watch(
  () => [state.value.min, state.value.max],
  ([a, b]) => {
    minInput.value = a !== undefined ? String(a) : ''
    maxInput.value = b !== undefined ? String(b) : ''
  },
  { immediate: true },
)
function applyPrice() {
  let a = minInput.value.trim() === '' ? undefined : Math.max(0, Math.floor(Number(minInput.value)))
  let b = maxInput.value.trim() === '' ? undefined : Math.max(0, Math.ceil(Number(maxInput.value)))
  if (a !== undefined && Number.isNaN(a)) a = undefined
  if (b !== undefined && Number.isNaN(b)) b = undefined
  if (a !== undefined && b !== undefined && a > b) [a, b] = [b, a]
  update({
    min: a !== undefined ? String(a) : undefined,
    max: b !== undefined ? String(b) : undefined,
  })
}

const activeFilterCount = computed(() => {
  const s = state.value
  return (
    (s.category && !props.fixedCategory ? 1 : 0) +
    (s.min !== undefined || s.max !== undefined ? 1 : 0) +
    (s.inStock ? 1 : 0)
  )
})

function clearFilters() {
  update({ category: undefined, min: undefined, max: undefined, in_stock: undefined, q: undefined })
}

const sortOptions = computed<SelectOption<ProductSort>[]>(() => [
  { value: 'popular', label: t('catalog.sortPopular') },
  { value: 'newest', label: t('catalog.sortNewest') },
  { value: 'price_asc', label: t('catalog.sortPriceAsc') },
  { value: 'price_desc', label: t('catalog.sortPriceDesc') },
  { value: 'rating', label: t('catalog.sortRating') },
])

const categoryOptions = computed<SelectOption<string>[]>(() => [
  { value: '', label: t('catalog.allCategories') },
  ...(categories.data.value ?? []).map((c) => ({ value: c.slug, label: loc(c, 'name') })),
])

const drawer = ref(false)
</script>

<template>
  <div class="browser">
    <div class="toolbar">
      <p class="count" aria-live="polite">
        <template v-if="list.data.value">{{ $t('catalog.results', total) }}</template>
      </p>
      <BaseButton variant="secondary" class="filter-btn" @click="drawer = true">
        <SlidersHorizontal :size="16" aria-hidden="true" />{{ $t('catalog.filters') }}
        <span v-if="activeFilterCount" class="fcount">{{ activeFilterCount }}</span>
      </BaseButton>
      <div class="sort">
        <BaseSelect v-model="sortModel" :options="sortOptions" :aria-label="$t('catalog.sort')" />
      </div>
    </div>

    <div v-if="state.q || activeFilterCount" class="chips">
      <button v-if="state.q" type="button" class="chip" @click="update({ q: undefined })">
        “{{ state.q }}” <X :size="14" aria-hidden="true" /><span class="sr-only">{{
          $t('common.clear')
        }}</span>
      </button>
      <button
        v-if="state.category && !fixedCategory"
        type="button"
        class="chip"
        @click="update({ category: undefined })"
      >
        {{ categoryOptions.find((o) => o.value === state.category)?.label ?? state.category }}
        <X :size="14" aria-hidden="true" />
      </button>
      <button
        v-if="state.min !== undefined || state.max !== undefined"
        type="button"
        class="chip"
        @click="update({ min: undefined, max: undefined })"
      >
        ฿{{ state.min ?? 0 }} – {{ state.max !== undefined ? `฿${state.max}` : '∞' }}
        <X :size="14" aria-hidden="true" />
      </button>
      <button
        v-if="state.inStock"
        type="button"
        class="chip"
        @click="update({ in_stock: undefined })"
      >
        {{ $t('catalog.inStockOnly') }} <X :size="14" aria-hidden="true" />
      </button>
      <button type="button" class="clear-all" @click="clearFilters">
        {{ $t('catalog.clearFilters') }}
      </button>
    </div>

    <div class="layout">
      <aside class="side" :aria-label="$t('catalog.filters')">
        <div class="card card-pad filters">
          <h2 class="fh">{{ $t('catalog.filters') }}</h2>
          <BaseSelect
            v-if="!fixedCategory"
            v-model="categoryModel"
            :label="$t('catalog.category')"
            :options="categoryOptions"
          />
          <fieldset class="price">
            <legend>{{ $t('catalog.price') }}</legend>
            <div class="price-row">
              <input
                v-model="minInput"
                type="number"
                min="0"
                inputmode="numeric"
                :placeholder="$t('catalog.minPrice')"
                :aria-label="$t('catalog.minPrice')"
                @keydown.enter="applyPrice"
              />
              <span aria-hidden="true">–</span>
              <input
                v-model="maxInput"
                type="number"
                min="0"
                inputmode="numeric"
                :placeholder="$t('catalog.maxPrice')"
                :aria-label="$t('catalog.maxPrice')"
                @keydown.enter="applyPrice"
              />
            </div>
            <BaseButton size="sm" variant="secondary" block @click="applyPrice">{{
              $t('common.apply')
            }}</BaseButton>
          </fieldset>
          <BaseCheckbox v-model="inStockModel" :label="$t('catalog.inStockOnly')" toggle />
          <BaseButton v-if="activeFilterCount" variant="ghost" size="sm" @click="clearFilters">
            {{ $t('catalog.clearFilters') }}
          </BaseButton>
        </div>
      </aside>

      <div class="results">
        <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
        <template v-else>
          <ProductGrid
            :products="items"
            :loading="list.loading.value && !list.data.value"
            :skeletons="8"
            dense
          />
          <EmptyState
            v-if="list.data.value && items.length === 0 && !list.loading.value"
            :title="$t('catalog.noResults')"
            :text="$t('catalog.noResultsHint')"
          >
            <BaseButton
              v-if="state.q || activeFilterCount"
              variant="secondary"
              @click="clearFilters"
            >
              {{ $t('catalog.clearFilters') }}
            </BaseButton>
          </EmptyState>
          <div v-if="list.loading.value && list.data.value" class="refreshing" aria-hidden="true" />
          <BasePagination v-model:page="pageModel" :pages="pages" />
        </template>
      </div>
    </div>

    <BaseDrawer v-model:open="drawer" side="bottom" :title="$t('catalog.filters')">
      <div class="drawer-filters">
        <BaseSelect
          v-if="!fixedCategory"
          v-model="categoryModel"
          :label="$t('catalog.category')"
          :options="categoryOptions"
        />
        <fieldset class="price">
          <legend>{{ $t('catalog.price') }}</legend>
          <div class="price-row">
            <input
              v-model="minInput"
              type="number"
              min="0"
              inputmode="numeric"
              :placeholder="$t('catalog.minPrice')"
              :aria-label="$t('catalog.minPrice')"
            />
            <span aria-hidden="true">–</span>
            <input
              v-model="maxInput"
              type="number"
              min="0"
              inputmode="numeric"
              :placeholder="$t('catalog.maxPrice')"
              :aria-label="$t('catalog.maxPrice')"
            />
          </div>
        </fieldset>
        <BaseCheckbox v-model="inStockModel" :label="$t('catalog.inStockOnly')" toggle />
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="clearFilters">{{ $t('common.reset') }}</BaseButton>
        <BaseButton
          block
          @click="
            () => {
              applyPrice()
              drawer = false
            }
          "
        >
          {{ $t('common.apply') }}
        </BaseButton>
      </template>
    </BaseDrawer>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  margin-bottom: var(--sp-4);
}
.count {
  margin: 0;
  color: var(--c-muted);
  font-size: var(--fs-sm);
  flex: 1;
  min-width: 0;
}
.sort {
  width: 190px;
}
.filter-btn {
  position: relative;
}
.fcount {
  background: var(--c-primary);
  color: #fff;
  border-radius: 999px;
  min-width: 20px;
  height: 20px;
  font-size: 11px;
  display: grid;
  place-items: center;
}
@media (max-width: 639px) {
  .toolbar {
    flex-wrap: wrap;
  }
  .count {
    flex-basis: 100%;
  }
  .filter-btn {
    flex: 1;
  }
  .sort {
    flex: 1.4;
    width: auto;
  }
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sp-2);
  margin-bottom: var(--sp-4);
  align-items: center;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px 5px 12px;
  border-radius: var(--radius-full);
  border: 1px solid var(--c-primary-200);
  background: var(--c-primary-50);
  color: var(--c-primary-hover);
  font-size: var(--fs-xs);
  font-weight: var(--fw-semibold);
  cursor: pointer;
}
.chip:hover {
  background: var(--c-primary-100);
}
.clear-all {
  border: 0;
  background: none;
  color: var(--c-muted);
  font-size: var(--fs-xs);
  text-decoration: underline;
  cursor: pointer;
}
.layout {
  display: grid;
  gap: var(--sp-6);
}
.side {
  display: none;
}
@media (min-width: 1024px) {
  .layout {
    grid-template-columns: 250px minmax(0, 1fr);
  }
  .side {
    display: block;
  }
  .filter-btn {
    display: none;
  }
}
.filters {
  position: sticky;
  top: calc(var(--header-h) + var(--sp-4));
  display: flex;
  flex-direction: column;
  gap: var(--sp-5);
}
.fh {
  font-size: var(--fs-md);
  margin: 0;
}
.price {
  border: 0;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
legend {
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
  margin-bottom: var(--sp-1);
  padding: 0;
}
.price-row {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
}
.price-row input {
  width: 100%;
  min-width: 0;
  height: 40px;
  padding: 0 var(--sp-3);
  border: 1px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  font-size: 16px;
}
@media (min-width: 768px) {
  .price-row input {
    font-size: var(--fs-sm);
  }
}
.price-row input:focus {
  outline: none;
  border-color: var(--c-primary-600);
  box-shadow: var(--ring);
}
.results {
  position: relative;
  min-width: 0;
}
.refreshing {
  position: absolute;
  inset: 0;
  background: rgb(247 247 251 / 0.5);
  border-radius: var(--radius-lg);
}
.drawer-filters {
  display: flex;
  flex-direction: column;
  gap: var(--sp-5);
}
</style>
