<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, ArrowRight, ShoppingCart, Trash2 } from 'lucide-vue-next'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import QuantityStepper from '@/components/ui/QuantityStepper.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { useCartStore, MAX_LINE_QTY, type CartItem } from '@/stores/cart'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { formatBaht } from '@/utils/money'

const cart = useCartStore()
const auth = useAuthStore()
const ui = useUiStore()
const router = useRouter()
const { t } = useI18n()
const { loc } = useLocalized()
const { errorText } = useErrorMessage()

onMounted(async () => {
  try {
    await cart.validate()
  } catch (e) {
    ui.error(errorText(e))
  }
})

function name(item: CartItem) {
  return item.info ? loc(item.info, 'product_name') : t('cart.unknownItem', { id: item.variant_id })
}
function variantName(item: CartItem) {
  return item.info ? loc(item.info, 'variant_name') : ''
}
function maxFor(item: CartItem) {
  const avail = cart.problemFor(item.variant_id)?.available ?? item.info?.available
  return Math.max(1, Math.min(MAX_LINE_QTY, avail ?? MAX_LINE_QTY))
}

let timer: ReturnType<typeof setTimeout> | undefined
function setQty(item: CartItem, q: number) {
  cart.setQuantity(item.variant_id, q)
  clearTimeout(timer)
  timer = setTimeout(() => void cart.validate().catch(() => undefined), 500)
}
function remove(item: CartItem) {
  cart.remove(item.variant_id)
  ui.info(t('cart.removed'))
}

const blocking = computed(() => cart.problems.length > 0)

function checkout() {
  if (!auth.isLoggedIn) void router.push({ name: 'login', query: { redirect: '/checkout' } })
  else void router.push({ name: 'checkout' })
}
</script>

<template>
  <div class="container page">
    <h1 class="page-title">{{ $t('cart.title') }}</h1>

    <EmptyState
      v-if="cart.items.length === 0"
      :icon="ShoppingCart"
      :title="$t('cart.empty')"
      :text="$t('cart.emptyHint')"
    >
      <BaseButton :to="{ name: 'products' }" size="lg">{{ $t('cart.continue') }}</BaseButton>
    </EmptyState>

    <div v-else class="layout">
      <section class="card lines" :aria-label="$t('cart.title')" :aria-busy="cart.validating">
        <ul>
          <li
            v-for="item in cart.items"
            :key="item.variant_id"
            class="line"
            :class="{ bad: !!cart.problemFor(item.variant_id) }"
          >
            <RouterLink
              v-if="item.info?.product_slug"
              :to="{ name: 'product', params: { slug: item.info.product_slug } }"
              class="thumb"
              tabindex="-1"
              aria-hidden="true"
            >
              <ImagePlaceholder :src="item.info?.image_url" :name="name(item)" small />
            </RouterLink>
            <div v-else class="thumb"><ImagePlaceholder :name="name(item)" small /></div>

            <div class="main">
              <RouterLink
                v-if="item.info?.product_slug"
                :to="{ name: 'product', params: { slug: item.info.product_slug } }"
                class="name"
              >
                {{ name(item) }}
              </RouterLink>
              <span v-else class="name">{{ name(item) }}</span>
              <p v-if="variantName(item)" class="variant muted">{{ variantName(item) }}</p>
              <p v-if="item.info" class="unit muted">
                {{ $t('cart.unitPrice', { price: formatBaht(item.info.unit_price_satang) }) }}
              </p>

              <p v-if="cart.problemFor(item.variant_id)" class="problem" role="alert">
                <AlertTriangle :size="16" aria-hidden="true" />
                <span>{{
                  $t(`cart.problem.${cart.problemFor(item.variant_id)!.code}`, {
                    n: cart.problemFor(item.variant_id)!.available,
                  })
                }}</span>
                <button
                  v-if="
                    cart.problemFor(item.variant_id)!.code === 'insufficient_stock' &&
                    cart.problemFor(item.variant_id)!.available > 0
                  "
                  type="button"
                  class="fix"
                  @click="setQty(item, cart.problemFor(item.variant_id)!.available)"
                >
                  {{ $t('cart.adjust', { n: cart.problemFor(item.variant_id)!.available }) }}
                </button>
              </p>

              <div class="controls">
                <QuantityStepper
                  :model-value="item.quantity"
                  size="sm"
                  :max="maxFor(item)"
                  @update:model-value="setQty(item, $event)"
                />
                <strong class="total tabular">{{
                  formatBaht((item.info?.unit_price_satang ?? 0) * item.quantity)
                }}</strong>
                <button
                  type="button"
                  class="rm"
                  :aria-label="$t('cart.remove', { name: name(item) })"
                  @click="remove(item)"
                >
                  <Trash2 :size="18" aria-hidden="true" />
                </button>
              </div>
            </div>
          </li>
        </ul>
      </section>

      <aside class="card card-pad summary" :aria-label="$t('cart.summary')">
        <h2>{{ $t('cart.summary') }}</h2>
        <div class="sum-row">
          <span>{{ $t('cart.subtotal') }} ({{ $t('common.items', cart.count) }})</span>
          <strong>{{ formatBaht(cart.subtotal) }}</strong>
        </div>
        <p class="muted text-xs">{{ $t('cart.shippingNote') }}</p>
        <p v-if="cart.validating" class="muted text-xs" aria-live="polite">
          {{ $t('cart.refreshing') }}
        </p>
        <p v-if="blocking" class="problem" role="alert">
          <AlertTriangle :size="16" aria-hidden="true" />{{ $t('cart.fixProblems') }}
        </p>
        <BaseButton size="lg" block :disabled="blocking || cart.validating" @click="checkout">
          {{ auth.isLoggedIn ? $t('cart.checkout') : $t('cart.loginToCheckout') }}
          <ArrowRight :size="18" aria-hidden="true" />
        </BaseButton>
        <BaseButton :to="{ name: 'products' }" variant="ghost" block>{{
          $t('cart.continue')
        }}</BaseButton>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.layout {
  display: grid;
  gap: var(--sp-5);
  align-items: start;
}
@media (min-width: 900px) {
  .layout {
    grid-template-columns: minmax(0, 1fr) 340px;
  }
  .summary {
    position: sticky;
    top: calc(var(--header-h) + var(--sp-4));
  }
}
.lines ul {
  list-style: none;
  margin: 0;
  padding: 0;
}
.line {
  display: flex;
  gap: var(--sp-3);
  padding: var(--sp-4);
  border-bottom: 1px solid var(--c-border);
}
.line:last-child {
  border-bottom: 0;
}
.line.bad {
  background: #fffaf5;
}
.thumb {
  flex: none;
  width: 76px;
}
@media (min-width: 640px) {
  .thumb {
    width: 96px;
  }
}
.main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.name {
  font-weight: var(--fw-semibold);
  color: var(--c-text);
  line-height: 1.4;
}
.variant,
.unit {
  margin: 0;
  font-size: var(--fs-xs);
}
.controls {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  margin-top: var(--sp-2);
  flex-wrap: wrap;
}
.total {
  margin-left: auto;
}
.rm {
  width: 36px;
  height: 36px;
  border: 0;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--c-muted);
  display: grid;
  place-items: center;
  cursor: pointer;
}
.rm:hover {
  background: var(--c-danger-bg);
  color: var(--c-danger);
}
.problem {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin: var(--sp-2) 0 0;
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  background: var(--c-warning-bg);
  color: var(--c-warning);
  font-size: var(--fs-xs);
  font-weight: var(--fw-medium);
}
.fix {
  border: 0;
  background: none;
  color: var(--c-primary);
  font-weight: var(--fw-semibold);
  text-decoration: underline;
  cursor: pointer;
  font-size: var(--fs-xs);
}
.summary {
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.summary h2 {
  font-size: var(--fs-lg);
  margin: 0;
}
.summary p {
  margin: 0;
}
</style>
