<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  CheckCircle2,
  ChevronRight,
  ShoppingCart,
  Truck,
  XCircle,
  AlertCircle,
} from 'lucide-vue-next'
import ProductGallery from '@/components/store/ProductGallery.vue'
import ProductReviews from '@/components/store/ProductReviews.vue'
import PriceTag from '@/components/store/PriceTag.vue'
import WishlistButton from '@/components/store/WishlistButton.vue'
import RatingStars from '@/components/ui/RatingStars.vue'
import QuantityStepper from '@/components/ui/QuantityStepper.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { useAsync } from '@/composables/useAsync'
import { useLocalized } from '@/composables/useLocalized'
import { usePageTitle } from '@/composables/usePageTitle'
import { useCartStore, MAX_LINE_QTY } from '@/stores/cart'
import { useUiStore } from '@/stores/ui'
import { isApiError } from '@/api/client'
import { catalogApi } from '@/api/catalog'
import type { Variant } from '@/types/api'

const props = defineProps<{ slug: string }>()
const { loc } = useLocalized()
const { t } = useI18n()
const cart = useCartStore()
const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const product = useAsync((slug: string) => catalogApi.product(slug))
watch(
  () => props.slug,
  (s) => void product.run(s),
  { immediate: true },
)

const p = computed(() => product.data.value)
const name = computed(() => (p.value ? loc(p.value, 'name') : ''))
const description = computed(() => (p.value ? loc(p.value, 'description') : ''))
const notFound = computed(
  () => isApiError(product.error.value) && product.error.value.status === 404,
)
usePageTitle(name)

const variants = computed<Variant[]>(() =>
  [...(p.value?.variants ?? [])]
    .filter((v) => v.is_active)
    .sort((a, b) => a.sort_order - b.sort_order),
)
const selectedId = ref<number | null>(null)
watch(
  variants,
  (vs) => {
    const fromQuery = Number(route.query.v)
    const pick = vs.find((v) => v.id === fromQuery) ?? vs.find((v) => v.stock > 0) ?? vs[0] ?? null
    selectedId.value = pick?.id ?? null
  },
  { immediate: true },
)
const selected = computed(() => variants.value.find((v) => v.id === selectedId.value) ?? null)

const inCart = computed(() => (selected.value ? (cart.find(selected.value.id)?.quantity ?? 0) : 0))
const maxQty = computed(() =>
  selected.value ? Math.max(0, Math.min(selected.value.stock, MAX_LINE_QTY) - inCart.value) : 0,
)
const qty = ref(1)
watch([selectedId, maxQty], () => {
  if (qty.value > maxQty.value) qty.value = Math.max(1, maxQty.value)
})

const stockState = computed<'out' | 'low' | 'in'>(() => {
  const s = selected.value?.stock ?? 0
  if (s <= 0) return 'out'
  if (s <= 5) return 'low'
  return 'in'
})

function selectVariant(v: Variant) {
  selectedId.value = v.id
  void router.replace({ query: { ...route.query, v: String(v.id) } })
}

function addToCart(): boolean {
  const v = selected.value
  if (!p.value || !v || v.stock <= 0 || maxQty.value <= 0) {
    if (v && maxQty.value <= 0 && v.stock > 0) ui.warning(t('cart.maxReached'))
    return false
  }
  cart.add(v.id, Math.min(qty.value, maxQty.value), {
    product_slug: p.value.slug,
    product_name_en: p.value.name_en,
    product_name_my: p.value.name_my,
    variant_name_en: v.name_en,
    variant_name_my: v.name_my,
    image_url: p.value.images[0]?.url ?? p.value.image_url,
    unit_price_satang: v.price_satang,
    available: v.stock,
  })
  return true
}

function onAdd() {
  if (addToCart()) {
    ui.success(t('product.added'))
    qty.value = 1
  }
}

function onBuyNow() {
  if (addToCart() || inCart.value > 0) void router.push({ name: 'checkout' })
}
</script>

<template>
  <div class="container page">
    <EmptyState v-if="notFound" :title="$t('product.notFound')">
      <BaseButton :to="{ name: 'products' }">{{ $t('home.browse') }}</BaseButton>
    </EmptyState>
    <ErrorState
      v-else-if="product.error.value"
      :error="product.error.value"
      @retry="product.retry"
    />

    <div v-else-if="!p" class="layout" aria-busy="true">
      <SkeletonBlock height="0" radius="var(--radius-xl)" class="sk-img" />
      <div class="stack">
        <SkeletonBlock width="30%" height="14px" />
        <SkeletonBlock height="32px" />
        <SkeletonBlock width="40%" height="28px" />
        <SkeletonBlock :lines="3" />
        <SkeletonBlock height="50px" />
      </div>
    </div>

    <template v-else>
      <nav class="crumbs" :aria-label="$t('nav.breadcrumb')">
        <RouterLink :to="{ name: 'home' }">{{ $t('common.home') }}</RouterLink>
        <ChevronRight :size="14" aria-hidden="true" />
        <template v-if="p.category">
          <RouterLink :to="{ name: 'category', params: { slug: p.category.slug } }">
            {{ loc(p.category, 'name') }}
          </RouterLink>
          <ChevronRight :size="14" aria-hidden="true" />
        </template>
        <span aria-current="page">{{ name }}</span>
      </nav>

      <div class="layout">
        <ProductGallery :images="p.images" :name="name" />

        <div class="info">
          <h1 class="title">{{ name }}</h1>
          <div class="meta">
            <a v-if="p.rating_count" href="#reviews" class="rating">
              <RatingStars :value="p.rating_avg" :size="16" />
              <span>{{ p.rating_avg.toFixed(1) }}</span>
              <span class="muted">({{ $t('product.reviewsCount', p.rating_count) }})</span>
            </a>
            <span v-if="p.sold_count" class="muted">{{
              $t('product.sold', { n: p.sold_count })
            }}</span>
          </div>

          <div class="price-box">
            <PriceTag
              size="lg"
              show-badge
              :price="selected?.price_satang ?? p.price_satang"
              :compare-at="selected ? selected.compare_at_satang : p.compare_at_satang"
            />
          </div>

          <fieldset
            v-if="variants.length > 1 || (variants[0] && variants[0].name_en)"
            class="variants"
          >
            <legend>
              {{ $t('product.variant') }}:
              <strong>{{ selected ? loc(selected, 'name') : $t('product.chooseVariant') }}</strong>
            </legend>
            <div class="chips" role="radiogroup" :aria-label="$t('product.variant')">
              <button
                v-for="v in variants"
                :key="v.id"
                type="button"
                role="radio"
                class="vchip"
                :class="{ on: v.id === selectedId, oos: v.stock <= 0 }"
                :aria-checked="v.id === selectedId"
                @click="selectVariant(v)"
              >
                {{ loc(v, 'name') || v.sku }}
              </button>
            </div>
          </fieldset>

          <p class="stock" :class="`s-${stockState}`">
            <CheckCircle2 v-if="stockState === 'in'" :size="18" aria-hidden="true" />
            <AlertCircle v-else-if="stockState === 'low'" :size="18" aria-hidden="true" />
            <XCircle v-else :size="18" aria-hidden="true" />
            <template v-if="stockState === 'in'">{{ $t('product.inStock') }}</template>
            <template v-else-if="stockState === 'low'">{{
              $t('product.lowStock', { n: selected?.stock ?? 0 })
            }}</template>
            <template v-else>{{ $t('product.outOfStock') }}</template>
          </p>

          <div class="buy">
            <QuantityStepper
              v-if="stockState !== 'out'"
              v-model="qty"
              :min="1"
              :max="Math.max(1, maxQty)"
              :disabled="maxQty <= 0"
            />
            <BaseButton
              size="lg"
              class="grow"
              :disabled="stockState === 'out' || maxQty <= 0"
              @click="onAdd"
            >
              <ShoppingCart :size="18" aria-hidden="true" />{{ $t('product.addToCart') }}
            </BaseButton>
            <WishlistButton :product-id="p.id" variant="outline" />
          </div>
          <BaseButton
            size="lg"
            variant="accent"
            block
            :disabled="stockState === 'out' || (maxQty <= 0 && inCart === 0)"
            @click="onBuyNow"
          >
            {{ $t('product.buyNow') }}
          </BaseButton>
          <p v-if="inCart > 0" class="in-cart muted text-sm">
            <RouterLink :to="{ name: 'cart' }"
              >{{ $t('nav.cart') }}: {{ $t('common.items', inCart) }}</RouterLink
            >
          </p>

          <p class="ship muted text-sm">
            <Truck :size="16" aria-hidden="true" />{{ $t('home.perk1') }}
          </p>

          <dl class="facts text-sm">
            <template v-if="selected">
              <dt>{{ $t('product.sku') }}</dt>
              <dd>{{ selected.sku }}</dd>
            </template>
            <template v-if="p.category">
              <dt>{{ $t('product.category') }}</dt>
              <dd>
                <RouterLink :to="{ name: 'category', params: { slug: p.category.slug } }">
                  {{ loc(p.category, 'name') }}
                </RouterLink>
              </dd>
            </template>
          </dl>
        </div>
      </div>

      <section class="section card card-pad" aria-labelledby="h-desc">
        <h2 id="h-desc">{{ $t('product.description') }}</h2>
        <p v-if="description" class="desc">{{ description }}</p>
        <p v-else class="muted">{{ $t('product.noDescription') }}</p>
      </section>

      <div class="card card-pad">
        <ProductReviews :slug="p.slug" :rating-avg="p.rating_avg" :rating-count="p.rating_count" />
      </div>
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
  margin-bottom: var(--sp-4);
}
.crumbs a {
  color: var(--c-muted);
}
.crumbs span {
  color: var(--c-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 60vw;
}
.layout {
  display: grid;
  gap: var(--sp-6);
}
@media (min-width: 900px) {
  .layout {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: var(--sp-10);
    align-items: start;
  }
}
.sk-img :deep(.sk) {
  aspect-ratio: 1;
  height: auto !important;
}
.info {
  display: flex;
  flex-direction: column;
  gap: var(--sp-4);
}
@media (min-width: 900px) {
  .info {
    position: sticky;
    top: calc(var(--header-h) + var(--sp-4));
  }
}
.title {
  margin: 0;
  font-size: var(--fs-2xl);
}
.meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sp-4);
  align-items: center;
  font-size: var(--fs-sm);
  margin-top: calc(var(--sp-2) * -1);
}
.rating {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--c-text);
  font-weight: var(--fw-semibold);
}
.price-box {
  padding: var(--sp-4);
  background: var(--c-surface-2);
  border-radius: var(--radius-lg);
}
.variants {
  border: 0;
  margin: 0;
  padding: 0;
}
legend {
  font-size: var(--fs-sm);
  color: var(--c-text-2);
  margin-bottom: var(--sp-2);
  padding: 0;
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sp-2);
}
.vchip {
  min-height: 40px;
  padding: 0 var(--sp-4);
  border-radius: var(--radius-md);
  border: 1.5px solid var(--c-border-strong);
  background: var(--c-surface);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  cursor: pointer;
}
.vchip:hover {
  border-color: var(--c-primary-500);
}
.vchip.on {
  border-color: var(--c-primary);
  background: var(--c-primary-50);
  color: var(--c-primary-hover);
  font-weight: var(--fw-semibold);
}
.vchip.oos {
  color: var(--c-muted);
  text-decoration: line-through;
  border-style: dashed;
}
.stock {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  margin: 0;
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
}
.s-in {
  color: var(--c-success);
}
.s-low {
  color: var(--c-warning);
}
.s-out {
  color: var(--c-danger);
}
.buy {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sp-2);
  align-items: stretch;
}
.buy :deep(.stepper) {
  flex: none;
}
.grow {
  flex: 1 1 140px;
}
.in-cart,
.ship {
  margin: 0;
}
.ship {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
}
.facts {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--sp-1) var(--sp-4);
  margin: 0;
  padding-top: var(--sp-4);
  border-top: 1px solid var(--c-border);
}
dt {
  color: var(--c-muted);
}
dd {
  margin: 0;
}
.desc {
  white-space: pre-line;
  color: var(--c-text-2);
  margin: 0;
}
</style>
