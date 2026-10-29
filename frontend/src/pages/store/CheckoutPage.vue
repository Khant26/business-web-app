<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, Lock, MapPin, QrCode, Tag, Truck, X } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import AddressForm from '@/components/store/AddressForm.vue'
import AddressCard from '@/components/store/AddressCard.vue'
import OrderSummaryCard from '@/components/store/OrderSummaryCard.vue'
import {
  emptyAddress,
  useAddressValidation,
  type AddressFields,
} from '@/components/store/checkout/address'
import { accountApi } from '@/api/account'
import { checkoutApi } from '@/api/orders'
import { isApiError } from '@/api/client'
import { useAsync } from '@/composables/useAsync'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useLocalized } from '@/composables/useLocalized'
import { useCartStore } from '@/stores/cart'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import { useUiStore } from '@/stores/ui'
import { formatBaht } from '@/utils/money'
import type { PaymentMethod, PlaceOrderInput, Quote } from '@/types/api'

const cart = useCartStore()
const auth = useAuthStore()
const settings = useSettingsStore()
const ui = useUiStore()
const router = useRouter()
const { t } = useI18n()
const { loc } = useLocalized()
const { errorText, codeText } = useErrorMessage()
const { validate: validateAddress, fromApiFields } = useAddressValidation()

// ----- Addresses -----
const addresses = useAsync(() => accountApi.addresses())
const selected = ref<number | 'new'>('new')
const newAddress = ref<AddressFields>({
  ...emptyAddress(),
  recipient_name: auth.user?.name ?? '',
  phone: auth.user?.phone ?? '',
})
const saveAddress = ref(true)
const addressErrors = ref<Record<string, string>>({})

const savedList = computed(() => addresses.data.value ?? [])
const selectedAddress = computed(() =>
  selected.value === 'new' ? null : (savedList.value.find((a) => a.id === selected.value) ?? null),
)
const province = computed(() =>
  selected.value === 'new' ? newAddress.value.province : (selectedAddress.value?.province ?? ''),
)

async function loadAddresses() {
  const list = await addresses.run()
  if (list && list.length > 0) {
    const def = list.find((a) => a.is_default) ?? list[0]
    if (def) selected.value = def.id
  }
}

// ----- Payment / coupon / note -----
const payment = ref<PaymentMethod>('promptpay')
const couponInput = ref('')
const appliedCoupon = ref('')
const couponError = ref('')
const note = ref('')

// ----- Quote (debounced, latest response wins) -----
const quote = ref<Quote | null>(null)
const quoteError = ref<unknown>(null)
const quoting = ref(false)
let quoteSeq = 0
let timer: ReturnType<typeof setTimeout> | undefined

async function requote() {
  if (cart.items.length === 0) return
  const id = ++quoteSeq
  quoting.value = true
  try {
    const q = await checkoutApi.quote({
      items: cart.payload,
      province: province.value,
      coupon_code: appliedCoupon.value || undefined,
    })
    if (id !== quoteSeq) return
    quote.value = q
    quoteError.value = null
    cart.applyLines(q.lines)
    cart.problems = q.problems
    if (q.coupon && !q.coupon.valid) {
      couponError.value = codeText(q.coupon.message_code || 'coupon_invalid')
      appliedCoupon.value = ''
    }
    if (!q.cod_allowed && payment.value === 'cod') payment.value = 'promptpay'
  } catch (e) {
    if (id === quoteSeq) quoteError.value = e
  } finally {
    if (id === quoteSeq) quoting.value = false
  }
}

function scheduleQuote(delay = 350) {
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => void requote(), delay)
}

watch([() => JSON.stringify(cart.payload), province, appliedCoupon], () => scheduleQuote())
onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})

onMounted(async () => {
  await loadAddresses()
  scheduleQuote(0)
})

function applyCoupon() {
  const code = couponInput.value.trim().toUpperCase()
  couponError.value = ''
  if (!code) return
  appliedCoupon.value = code
}
function removeCoupon() {
  appliedCoupon.value = ''
  couponInput.value = ''
  couponError.value = ''
}

const validCoupon = computed(() =>
  quote.value?.coupon?.valid && appliedCoupon.value ? quote.value.coupon.code : '',
)

// ----- Derived summary -----
const lines = computed(() => quote.value?.lines ?? [])
const problemFor = (variantId: number) =>
  quote.value?.problems.find((p) => p.variant_id === variantId)
const hasProblems = computed(() => (quote.value?.problems.length ?? 0) > 0)

const freeShippingGap = computed(() => {
  const q = quote.value
  const free = q?.zone?.free_over_satang
  if (!q || free == null) return null
  const net = q.subtotal_satang - q.discount_satang
  return free > net ? free - net : 0
})

const codAllowed = computed(() => quote.value?.cod_allowed ?? true)

// ----- Place order -----
const placing = ref(false)
const formError = ref('')

async function placeOrder() {
  if (placing.value) return
  formError.value = ''
  addressErrors.value = {}
  const body: PlaceOrderInput = {
    items: cart.payload,
    payment_method: payment.value,
    coupon_code: validCoupon.value || undefined,
    note: note.value.trim() || undefined,
  }
  if (selected.value === 'new') {
    const r = validateAddress(newAddress.value)
    if (!r.ok) {
      addressErrors.value = r.errors
      formError.value = t('errors.validation')
      document.getElementById('checkout-address')?.scrollIntoView({ behavior: 'smooth' })
      return
    }
    body.address = r.data
    body.save_address = saveAddress.value
  } else {
    body.address_id = selected.value
  }

  placing.value = true
  try {
    const order = await checkoutApi.placeOrder(body)
    cart.clear()
    if (order.payment_method === 'promptpay' && order.status === 'pending_payment') {
      await router.replace({ name: 'account-pay', params: { code: order.code } })
    } else {
      await router.replace({
        name: 'account-order',
        params: { code: order.code },
        query: { placed: '1' },
      })
    }
  } catch (e) {
    formError.value = errorText(e)
    if (isApiError(e)) {
      if (e.code === 'validation' && Object.keys(e.fields).length) {
        addressErrors.value = fromApiFields(e.fields)
      }
      if (['out_of_stock', 'insufficient_stock', 'unavailable'].includes(e.code)) {
        void requote()
      }
      if (e.code === 'coupon_invalid') {
        couponError.value = errorText(e)
        appliedCoupon.value = ''
      }
      if (e.code === 'cod_not_allowed') payment.value = 'promptpay'
    }
    ui.error(formError.value)
  } finally {
    placing.value = false
  }
}

const canPlace = computed(
  () => !!quote.value && !hasProblems.value && !quoting.value && cart.items.length > 0,
)
</script>

<template>
  <div class="container page">
    <h1 class="page-title">{{ $t('checkout.title') }}</h1>

    <EmptyState
      v-if="cart.items.length === 0"
      :title="$t('checkout.emptyTitle')"
      :text="$t('checkout.emptyText')"
    >
      <BaseButton :to="{ name: 'products' }">{{ $t('checkout.browse') }}</BaseButton>
    </EmptyState>

    <form v-else class="layout" novalidate @submit.prevent="placeOrder">
      <div class="main">
        <!-- Address -->
        <section id="checkout-address" class="card card-pad" aria-labelledby="h-addr">
          <h2 id="h-addr" class="sec-title">
            <MapPin :size="20" aria-hidden="true" />{{ $t('checkout.address') }}
          </h2>

          <div v-if="addresses.loading.value" class="stack">
            <SkeletonBlock height="72px" radius="var(--radius-md)" />
            <SkeletonBlock height="72px" radius="var(--radius-md)" />
          </div>
          <ErrorState
            v-else-if="addresses.error.value"
            compact
            :error="addresses.error.value"
            @retry="loadAddresses"
          />
          <template v-else>
            <fieldset v-if="savedList.length" class="choices">
              <legend class="sr-only">{{ $t('checkout.savedAddresses') }}</legend>
              <label
                v-for="a in savedList"
                :key="a.id"
                class="choice"
                :class="{ on: selected === a.id }"
              >
                <input v-model="selected" type="radio" name="addr" :value="a.id" class="radio" />
                <AddressCard :address="a" />
              </label>
              <label class="choice" :class="{ on: selected === 'new' }">
                <input v-model="selected" type="radio" name="addr" value="new" class="radio" />
                <span class="new-lbl">{{ $t('checkout.useNew') }}</span>
              </label>
            </fieldset>

            <div v-if="selected === 'new'" class="new-form">
              <AddressForm v-model="newAddress" :errors="addressErrors" show-label />
              <BaseCheckbox
                v-model="saveAddress"
                class="save-cb"
                :label="$t('checkout.saveAddress')"
              />
            </div>
          </template>
        </section>

        <!-- Payment -->
        <section class="card card-pad" aria-labelledby="h-pay">
          <h2 id="h-pay" class="sec-title">{{ $t('checkout.payment') }}</h2>
          <fieldset class="pay-grid">
            <legend class="sr-only">{{ $t('checkout.payment') }}</legend>
            <label class="pay-opt" :class="{ on: payment === 'promptpay' }">
              <input v-model="payment" type="radio" name="pay" value="promptpay" class="radio" />
              <span class="pay-ic"><QrCode :size="22" aria-hidden="true" /></span>
              <span class="pay-txt">
                <strong>{{ $t('payment.promptpay') }}</strong>
                <span class="muted text-sm">{{ $t('checkout.promptpayHint') }}</span>
              </span>
            </label>
            <label class="pay-opt" :class="{ on: payment === 'cod', disabled: !codAllowed }">
              <input
                v-model="payment"
                type="radio"
                name="pay"
                value="cod"
                class="radio"
                :disabled="!codAllowed"
                :aria-describedby="!codAllowed ? 'cod-reason' : undefined"
              />
              <span class="pay-ic"><Truck :size="22" aria-hidden="true" /></span>
              <span class="pay-txt">
                <strong>{{ $t('payment.cod') }}</strong>
                <span class="muted text-sm">{{ $t('checkout.codHint') }}</span>
                <span v-if="!codAllowed" id="cod-reason" class="text-danger text-xs">
                  {{
                    $t('checkout.codDisabled', {
                      max: formatBaht(settings.settings.cod_max_satang),
                    })
                  }}
                </span>
              </span>
            </label>
          </fieldset>
        </section>

        <!-- Coupon + note -->
        <section class="card card-pad" aria-labelledby="h-extra">
          <h2 id="h-extra" class="sr-only">{{ $t('checkout.coupon') }}</h2>
          <div v-if="validCoupon" class="coupon-on" role="status">
            <Tag :size="18" aria-hidden="true" />
            <span>{{ $t('checkout.couponApplied', { code: validCoupon }) }}</span>
            <button
              type="button"
              class="x"
              :aria-label="$t('checkout.removeCoupon')"
              @click="removeCoupon"
            >
              <X :size="16" aria-hidden="true" />
            </button>
          </div>
          <label v-if="!validCoupon" for="coupon-code" class="coupon-label">{{
            $t('checkout.coupon')
          }}</label>
          <div v-if="!validCoupon" class="coupon-row">
            <BaseInput
              id="coupon-code"
              v-model="couponInput"
              class="coupon-in"
              :placeholder="$t('checkout.couponPlaceholder')"
              :error="couponError"
              autocapitalize="characters"
              autocomplete="off"
              maxlength="40"
              @keydown.enter.prevent="applyCoupon"
            />
            <BaseButton
              variant="secondary"
              class="coupon-btn"
              :loading="quoting && !!appliedCoupon"
              :disabled="!couponInput.trim()"
              @click="applyCoupon"
            >
              {{ $t('checkout.applyCoupon') }}
            </BaseButton>
          </div>
          <BaseInput
            v-model="note"
            class="note"
            multiline
            :rows="3"
            :label="$t('checkout.note')"
            :placeholder="$t('checkout.notePlaceholder')"
            optional
            maxlength="500"
          />
        </section>
      </div>

      <!-- Summary -->
      <aside class="side">
        <section class="card card-pad summary" aria-labelledby="h-sum" :aria-busy="quoting">
          <div class="sum-head">
            <h2 id="h-sum" class="sec-title">{{ $t('checkout.summary') }}</h2>
            <RouterLink :to="{ name: 'cart' }" class="text-sm">{{
              $t('checkout.editCart')
            }}</RouterLink>
          </div>

          <ErrorState
            v-if="quoteError && !quote"
            compact
            :title="$t('checkout.quoteError')"
            :error="quoteError"
            @retry="requote"
          />
          <div v-else-if="!quote" class="stack">
            <SkeletonBlock v-for="i in 3" :key="i" height="48px" />
            <SkeletonBlock :lines="4" />
          </div>
          <template v-else>
            <p class="sr-only" aria-live="polite">{{ quoting ? $t('checkout.updating') : '' }}</p>
            <ul class="lines" :class="{ dim: quoting }">
              <li v-for="l in lines" :key="l.variant_id" class="line">
                <div class="thumb">
                  <ImagePlaceholder
                    :src="l.image_url"
                    :name="loc(l, 'product_name')"
                    small
                    rounded="var(--radius-sm)"
                  />
                  <span class="q" aria-hidden="true">{{ l.quantity }}</span>
                </div>
                <div class="l-info">
                  <span class="l-name">{{ loc(l, 'product_name') }}</span>
                  <span class="muted text-xs">
                    {{ loc(l, 'variant_name') }} · {{ $t('checkout.qty', { n: l.quantity }) }}
                  </span>
                  <span v-if="problemFor(l.variant_id)" class="text-danger text-xs prob">
                    <AlertTriangle :size="12" aria-hidden="true" />
                    {{
                      $t(`cart.problem.${problemFor(l.variant_id)?.code}`, {
                        n: problemFor(l.variant_id)?.available ?? 0,
                      })
                    }}
                  </span>
                </div>
                <span class="l-total tabular">{{ formatBaht(l.line_total_satang) }}</span>
              </li>
            </ul>

            <p v-if="hasProblems" class="alert" role="alert">
              <AlertTriangle :size="16" aria-hidden="true" />{{ $t('checkout.problems') }}
            </p>

            <p v-if="!province" class="hint muted">{{ $t('checkout.provinceFirst') }}</p>
            <p
              v-else-if="freeShippingGap !== null"
              class="ship-hint"
              :class="{ ok: freeShippingGap === 0 }"
            >
              <Truck :size="16" aria-hidden="true" />
              {{
                freeShippingGap === 0
                  ? $t('checkout.freeShippingUnlocked')
                  : $t('checkout.freeShippingHint', { amount: formatBaht(freeShippingGap) })
              }}
            </p>

            <OrderSummaryCard
              :class="{ dim: quoting }"
              :subtotal="quote.subtotal_satang"
              :discount="quote.discount_satang"
              :shipping="quote.shipping_satang"
              :vat-rate-bp="quote.vat_rate_bp"
              :vat="quote.vat_satang"
              :total="quote.total_satang"
              :coupon-code="validCoupon"
              :zone-name="quote.zone ? loc(quote.zone, 'name') : ''"
              :shipping-pending="!quote.zone"
            />
          </template>

          <p v-if="formError" class="form-err" role="alert">{{ formError }}</p>

          <BaseButton
            type="submit"
            size="lg"
            block
            class="place-desk"
            :loading="placing"
            :disabled="!canPlace"
          >
            {{ placing ? $t('checkout.placing') : $t('checkout.placeOrder') }}
          </BaseButton>
          <p class="secure muted text-xs">
            <Lock :size="12" aria-hidden="true" />{{ $t('checkout.secure') }}
          </p>
        </section>
      </aside>

      <!-- Mobile sticky bar -->
      <div class="mobile-bar">
        <div class="mb-total">
          <span class="muted text-xs">{{ $t('checkout.total') }}</span>
          <strong class="tabular">{{ quote ? formatBaht(quote.total_satang) : '—' }}</strong>
        </div>
        <BaseButton type="submit" size="lg" :loading="placing" :disabled="!canPlace">
          {{ $t('checkout.placeOrder') }}
        </BaseButton>
      </div>
    </form>
  </div>
</template>

<style scoped>
.layout {
  display: grid;
  gap: var(--sp-5);
  padding-bottom: 90px;
}
@media (min-width: 900px) {
  .layout {
    grid-template-columns: minmax(0, 1fr) 380px;
    align-items: start;
    padding-bottom: 0;
  }
}
@media (min-width: 1200px) {
  .layout {
    grid-template-columns: minmax(0, 1fr) 420px;
  }
}
.main {
  display: flex;
  flex-direction: column;
  gap: var(--sp-5);
  min-width: 0;
}
.side {
  min-width: 0;
}
@media (min-width: 900px) {
  .summary {
    position: sticky;
    top: calc(var(--header-h) + var(--sp-4));
  }
}
.sec-title {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  font-size: var(--fs-lg);
  margin-bottom: var(--sp-4);
}
.sec-title :deep(svg) {
  color: var(--c-primary);
}
.choices,
.pay-grid {
  border: 0;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--sp-3);
  min-width: 0;
}
.choice,
.pay-opt {
  display: flex;
  gap: var(--sp-3);
  align-items: flex-start;
  padding: var(--sp-4);
  border: 1.5px solid var(--c-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  min-width: 0;
  transition:
    border-color var(--dur-fast),
    background var(--dur-fast);
}
.choice:hover,
.pay-opt:hover {
  border-color: var(--c-primary-300);
}
.choice.on,
.pay-opt.on {
  border-color: var(--c-primary);
  background: var(--c-primary-50);
}
.choice:focus-within,
.pay-opt:focus-within {
  box-shadow: var(--ring);
}
.pay-opt.disabled {
  cursor: not-allowed;
  opacity: 0.75;
  background: var(--c-surface-2);
}
.radio {
  flex: none;
  width: 18px;
  height: 18px;
  margin-top: 3px;
  accent-color: var(--c-primary);
}
.radio:focus-visible {
  outline: none;
}
.new-lbl {
  font-weight: var(--fw-semibold);
}
.new-form {
  margin-top: var(--sp-4);
}
.save-cb {
  margin-top: var(--sp-4);
}
@media (min-width: 640px) {
  .pay-grid {
    grid-template-columns: 1fr 1fr;
  }
}
.pay-ic {
  flex: none;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  display: grid;
  place-items: center;
  color: var(--c-primary);
}
.pay-txt {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.coupon-label {
  display: block;
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
  margin-bottom: var(--sp-1);
}
.coupon-row {
  display: flex;
  gap: var(--sp-2);
  align-items: flex-start;
}
.coupon-in {
  flex: 1;
}
.coupon-btn {
  min-height: 44px;
}
.coupon-btn {
  flex: none;
}
.coupon-on {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-3) var(--sp-4);
  background: var(--c-success-bg);
  color: var(--c-success);
  border-radius: var(--radius-md);
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
}
.coupon-on .x {
  margin-left: auto;
  border: 0;
  background: transparent;
  color: inherit;
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  cursor: pointer;
}
.coupon-on .x:hover {
  background: rgb(0 0 0 / 0.06);
}
.note {
  margin-top: var(--sp-4);
}
.sum-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: var(--sp-3);
}
.lines {
  list-style: none;
  margin: 0 0 var(--sp-4);
  padding: 6px 0 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
  max-height: 340px;
  overflow-y: auto;
}
.line {
  display: flex;
  gap: var(--sp-3);
  align-items: flex-start;
}
.thumb {
  position: relative;
  flex: none;
  width: 52px;
}
.q {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: 999px;
  background: var(--c-text-2);
  color: #fff;
  font-size: 11px;
  font-weight: var(--fw-bold);
  display: grid;
  place-items: center;
}
.l-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  font-size: var(--fs-sm);
}
.l-name {
  font-weight: var(--fw-medium);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.prob {
  display: flex;
  align-items: center;
  gap: 4px;
}
.l-total {
  font-size: var(--fs-sm);
  font-weight: var(--fw-semibold);
  white-space: nowrap;
}
.dim {
  opacity: 0.55;
  transition: opacity var(--dur);
}
.alert,
.form-err {
  display: flex;
  gap: var(--sp-2);
  align-items: flex-start;
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  background: var(--c-danger-bg);
  color: var(--c-danger);
  font-size: var(--fs-sm);
  margin: 0 0 var(--sp-3);
}
.form-err {
  margin-top: var(--sp-3);
}
.alert :deep(svg) {
  flex: none;
  margin-top: 3px;
}
.hint {
  font-size: var(--fs-sm);
}
.ship-hint {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  background: var(--c-accent-50);
  color: var(--c-accent-strong);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
}
.ship-hint.ok {
  background: var(--c-success-bg);
  color: var(--c-success);
}
.place-desk {
  margin-top: var(--sp-4);
}
.secure {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  margin: var(--sp-2) 0 0;
}
.mobile-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: calc(var(--tabbar-h) + env(safe-area-inset-bottom));
  z-index: calc(var(--z-tabbar) - 1);
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-3) var(--sp-4);
  background: var(--c-surface);
  border-top: 1px solid var(--c-border);
  box-shadow: 0 -6px 18px -8px rgb(22 24 39 / 0.18);
}
.mb-total {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}
.mb-total strong {
  font-size: var(--fs-lg);
}
@media (max-width: 899px) {
  .place-desk {
    display: none;
  }
}
@media (min-width: 900px) {
  .mobile-bar {
    display: none;
  }
}
</style>
