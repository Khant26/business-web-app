<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ArrowLeft,
  CheckCircle2,
  Clock,
  ExternalLink,
  FileImage,
  MessageSquare,
  QrCode,
  Truck,
} from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import OrderTimeline from '@/components/store/OrderTimeline.vue'
import AddressCard from '@/components/store/AddressCard.vue'
import OrderSummaryCard from '@/components/store/OrderSummaryCard.vue'
import { useAsync } from '@/composables/useAsync'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { usePageTitle } from '@/composables/usePageTitle'
import { ordersApi } from '@/api/orders'
import { useUiStore } from '@/stores/ui'
import { formatBaht } from '@/utils/money'

const props = defineProps<{ code: string }>()
const route = useRoute()
const ui = useUiStore()
const { t } = useI18n()
const { loc, dateTime } = useLocalized()
const { errorText } = useErrorMessage()

const state = useAsync((code: string) => ordersApi.get(code))
watch(
  () => props.code,
  (c) => void state.run(c),
  { immediate: true },
)
const order = computed(() => state.data.value)
usePageTitle(() => t('orders.title', { code: props.code }))

const placed = computed(() => route.query.placed === '1')
const isPending = computed(() => order.value?.status === 'pending_payment')
const canPay = computed(() => isPending.value && order.value?.payment_method === 'promptpay')

// Cancel
const cancelOpen = ref(false)
const cancelReason = ref('')
const cancelling = ref(false)
async function confirmCancel() {
  if (!order.value) return
  cancelling.value = true
  try {
    const o = await ordersApi.cancel(order.value.code, cancelReason.value.trim() || undefined)
    state.set(o)
    cancelOpen.value = false
    cancelReason.value = ''
    ui.success(t('orders.cancelled'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    cancelling.value = false
  }
}

const safeTrackingUrl = computed(() => {
  const u = order.value?.tracking_url ?? ''
  return /^https?:\/\//i.test(u) ? u : ''
})
</script>

<template>
  <div>
    <RouterLink :to="{ name: 'account-orders' }" class="back">
      <ArrowLeft :size="16" aria-hidden="true" />{{ $t('orders.back') }}
    </RouterLink>

    <div v-if="state.loading.value && !order" class="stack" aria-busy="true">
      <SkeletonBlock height="32px" width="50%" />
      <SkeletonBlock height="160px" radius="var(--radius-lg)" />
      <SkeletonBlock height="220px" radius="var(--radius-lg)" />
    </div>
    <ErrorState v-else-if="state.error.value" :error="state.error.value" @retry="state.retry" />

    <template v-else-if="order">
      <div v-if="placed && order.status !== 'cancelled'" class="banner ok" role="status">
        <CheckCircle2 :size="22" aria-hidden="true" />
        <div>
          <strong>{{ $t('orders.placedTitle') }}</strong>
          <p>{{ canPay ? $t('orders.placedPromptpay') : $t('orders.placedText') }}</p>
        </div>
      </div>

      <header class="head">
        <div class="title-wrap">
          <h1 class="h">{{ $t('orders.title', { code: order.code }) }}</h1>
          <StatusBadge :status="order.status" />
        </div>
        <p class="muted text-sm meta">
          {{ $t('orders.placedOn', { date: dateTime(order.created_at) }) }} ·
          {{ $t(`payment.${order.payment_method}`) }}
        </p>
      </header>

      <div v-if="canPay" class="banner pay" role="status">
        <Clock :size="22" aria-hidden="true" />
        <div class="grow">
          <strong>{{ $t('status.pending_payment') }}</strong>
          <p v-if="order.payment_due_at">
            {{ $t('orders.dueBy', { time: dateTime(order.payment_due_at) }) }}
          </p>
        </div>
        <BaseButton
          v-if="order.can_upload_slip"
          :to="{ name: 'account-pay', params: { code: order.code } }"
        >
          <QrCode :size="18" aria-hidden="true" />{{ $t('orders.payNow') }}
        </BaseButton>
      </div>
      <p v-if="order.status === 'expired'" class="banner muted-b">{{ $t('orders.expiredText') }}</p>

      <div class="grid">
        <div class="col">
          <section class="card card-pad" aria-labelledby="h-items">
            <h2 id="h-items" class="sec">{{ $t('orders.items') }}</h2>
            <ul class="items">
              <li v-for="it in order.items" :key="it.id" class="item">
                <div class="thumb">
                  <ImagePlaceholder :src="it.image_url" :name="loc(it, 'product_name')" small />
                </div>
                <div class="i-info">
                  <RouterLink
                    v-if="it.product_slug"
                    :to="{ name: 'product', params: { slug: it.product_slug } }"
                    class="i-name"
                  >
                    {{ loc(it, 'product_name') }}
                  </RouterLink>
                  <span v-else class="i-name">{{ loc(it, 'product_name') }}</span>
                  <span class="muted text-xs">
                    {{ loc(it, 'variant_name') }} · {{ $t('orders.qty', { n: it.quantity }) }} ·
                    {{ formatBaht(it.unit_price_satang) }}
                  </span>
                  <RouterLink
                    v-if="it.can_review && it.product_slug"
                    :to="{
                      name: 'product',
                      params: { slug: it.product_slug },
                      hash: '#reviews',
                    }"
                    class="review text-sm"
                  >
                    <MessageSquare :size="14" aria-hidden="true" />{{ $t('orders.writeReview') }}
                  </RouterLink>
                </div>
                <span class="i-total tabular">{{ formatBaht(it.line_total_satang) }}</span>
              </li>
            </ul>
            <hr class="divider" />
            <OrderSummaryCard
              :subtotal="order.subtotal_satang"
              :discount="order.discount_satang"
              :shipping="order.shipping_satang"
              :vat-rate-bp="order.vat_rate_bp"
              :vat="order.vat_satang"
              :total="order.total_satang"
              :coupon-code="order.coupon_code"
            />
          </section>

          <section class="card card-pad" aria-labelledby="h-ship">
            <h2 id="h-ship" class="sec">
              <Truck :size="18" aria-hidden="true" />{{ $t('orders.shipment') }}
            </h2>
            <template v-if="order.tracking_number">
              <dl class="kv">
                <div v-if="order.courier">
                  <dt>{{ $t('orders.courier') }}</dt>
                  <dd>{{ order.courier }}</dd>
                </div>
                <div>
                  <dt>{{ $t('orders.trackingNumber') }}</dt>
                  <dd class="tabular mono">{{ order.tracking_number }}</dd>
                </div>
              </dl>
              <BaseButton
                v-if="safeTrackingUrl"
                :href="safeTrackingUrl"
                target="_blank"
                rel="noopener noreferrer"
                variant="secondary"
                size="sm"
                class="track"
              >
                {{ $t('orders.trackParcel') }}<ExternalLink :size="14" aria-hidden="true" />
              </BaseButton>
            </template>
            <p v-else class="muted text-sm m0">{{ $t('orders.notShipped') }}</p>
          </section>

          <section v-if="order.slips.length" class="card card-pad" aria-labelledby="h-slips">
            <h2 id="h-slips" class="sec">
              <FileImage :size="18" aria-hidden="true" />{{ $t('orders.slips') }}
            </h2>
            <ul class="slips">
              <li v-for="s in order.slips" :key="s.id" class="slip">
                <div>
                  <StatusBadge :status="s.status" kind="slip" />
                  <p class="muted text-xs m0">
                    {{ $t('orders.uploadedAt', { date: dateTime(s.uploaded_at) }) }}
                  </p>
                  <p v-if="s.note" class="text-sm m0">
                    {{ $t('orders.slipNote', { note: s.note }) }}
                  </p>
                </div>
              </li>
            </ul>
          </section>
        </div>

        <div class="col">
          <section class="card card-pad" aria-labelledby="h-prog">
            <h2 id="h-prog" class="sec">{{ $t('orders.progress') }}</h2>
            <OrderTimeline
              :status="order.status"
              :payment-method="order.payment_method"
              :history="order.history"
            />
            <p v-if="order.cancel_reason" class="text-sm muted reason">
              {{ $t('orders.cancelReason') }}: {{ order.cancel_reason }}
            </p>
          </section>

          <section class="card card-pad" aria-labelledby="h-addr">
            <h2 id="h-addr" class="sec">{{ $t('orders.address') }}</h2>
            <AddressCard :address="order.shipping_address" />
            <p class="text-sm muted contact">{{ order.email }}</p>
            <template v-if="order.note">
              <h3 class="sub">{{ $t('orders.note') }}</h3>
              <p class="text-sm m0 pre">{{ order.note }}</p>
            </template>
          </section>

          <BaseButton
            v-if="order.can_cancel"
            variant="secondary"
            block
            class="cancel"
            @click="cancelOpen = true"
          >
            {{ $t('orders.cancel') }}
          </BaseButton>
        </div>
      </div>

      <ConfirmDialog
        v-model:open="cancelOpen"
        :title="$t('orders.cancelTitle')"
        :message="$t('orders.cancelMessage')"
        :confirm-label="$t('orders.cancelConfirm')"
        :loading="cancelling"
        danger
        @confirm="confirmCancel"
      >
        <BaseInput
          v-model="cancelReason"
          class="reason-in"
          multiline
          :rows="2"
          :label="$t('orders.cancelReason')"
          :placeholder="$t('orders.cancelReasonPlaceholder')"
          optional
          maxlength="300"
        />
      </ConfirmDialog>
    </template>
  </div>
</template>

<style scoped>
.back {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-sm);
  margin-bottom: var(--sp-4);
}
.head {
  margin-bottom: var(--sp-4);
}
.title-wrap {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  flex-wrap: wrap;
}
.h {
  font-size: var(--fs-xl);
  margin: 0;
}
.meta {
  margin: var(--sp-1) 0 0;
}
.banner {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  flex-wrap: wrap;
  padding: var(--sp-4);
  border-radius: var(--radius-lg);
  margin-bottom: var(--sp-4);
}
.banner p {
  margin: 0;
  font-size: var(--fs-sm);
}
.banner.ok {
  background: var(--c-success-bg);
  color: var(--c-success);
}
.banner.ok p {
  color: var(--c-text-2);
}
.banner.pay {
  background: var(--c-warning-bg);
  color: var(--c-warning);
}
.banner.pay p {
  color: var(--c-text-2);
}
.banner.muted-b {
  background: var(--c-surface-3);
  color: var(--c-text-2);
  font-size: var(--fs-sm);
}
.banner :deep(svg) {
  flex: none;
}
.grow {
  flex: 1;
  min-width: 160px;
}
.grid {
  display: grid;
  gap: var(--sp-4);
}
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
    align-items: start;
  }
}
.col {
  display: flex;
  flex-direction: column;
  gap: var(--sp-4);
  min-width: 0;
}
.sec {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  font-size: var(--fs-md);
  margin-bottom: var(--sp-4);
}
.sec :deep(svg) {
  color: var(--c-primary);
}
.sub {
  font-size: var(--fs-sm);
  margin: var(--sp-4) 0 var(--sp-1);
}
.items {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-4);
}
.item {
  display: flex;
  gap: var(--sp-3);
  align-items: flex-start;
}
.thumb {
  flex: none;
  width: 64px;
}
.i-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.i-name {
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
  color: var(--c-text);
}
a.i-name:hover {
  color: var(--c-primary);
}
.review {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 2px;
  font-weight: var(--fw-medium);
}
.i-total {
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
  white-space: nowrap;
}
.kv {
  margin: 0;
  display: grid;
  gap: var(--sp-2);
  font-size: var(--fs-sm);
}
.kv div {
  display: flex;
  justify-content: space-between;
  gap: var(--sp-3);
}
.kv dt {
  color: var(--c-muted);
}
.kv dd {
  margin: 0;
  font-weight: var(--fw-semibold);
  text-align: right;
  word-break: break-all;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
.track {
  margin-top: var(--sp-3);
}
.slips {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.slip {
  padding-bottom: var(--sp-3);
  border-bottom: 1px solid var(--c-border);
}
.slip:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}
.slip p {
  margin-top: 4px;
}
.m0 {
  margin: 0;
}
.reason {
  margin: var(--sp-3) 0 0;
}
.contact {
  margin: var(--sp-2) 0 0;
}
.pre {
  white-space: pre-wrap;
}
.reason-in {
  margin-top: var(--sp-4);
}
</style>
