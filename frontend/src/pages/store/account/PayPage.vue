<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, ArrowLeft, Download, Hourglass, Timer, Upload } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import FileDrop from '@/components/ui/FileDrop.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { ordersApi } from '@/api/orders'
import { isApiError } from '@/api/client'
import { useNow } from '@/composables/useNow'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useLocalized } from '@/composables/useLocalized'
import { useUiStore } from '@/stores/ui'
import { formatBaht } from '@/utils/money'
import { formatCountdown, msUntil } from '@/utils/date'
import type { Order, PromptPayInfo } from '@/types/api'

const props = defineProps<{ code: string }>()
const ui = useUiStore()
const { t } = useI18n()
const { errorText } = useErrorMessage()
const { dateTime } = useLocalized()
const now = useNow(1000)

const order = ref<Order | null>(null)
const info = ref<PromptPayInfo | null>(null)
const loading = ref(true)
const error = ref<unknown>(null)
/** Error from the promptpay endpoint (promptpay_not_configured / not_payable). */
const payError = ref<unknown>(null)
const qrUrl = ref('')

async function renderQr(payload: string) {
  qrUrl.value = ''
  const QRCode = await import('qrcode')
  qrUrl.value = await QRCode.toDataURL(payload, {
    width: 360,
    margin: 2,
    errorCorrectionLevel: 'M',
    color: { dark: '#161827', light: '#ffffff' },
  })
}

async function load() {
  loading.value = true
  error.value = null
  payError.value = null
  info.value = null
  try {
    const o = await ordersApi.get(props.code)
    order.value = o
    if (o.status === 'pending_payment') {
      try {
        const p = await ordersApi.promptpay(o.code)
        info.value = p
        await renderQr(p.payload)
      } catch (e) {
        payError.value = e
      }
    }
  } catch (e) {
    error.value = e
  } finally {
    loading.value = false
  }
}

watch(
  () => props.code,
  () => void load(),
  { immediate: true },
)

const remaining = computed(() => (info.value ? msUntil(info.value.due_at, now.value) : 0))
const expired = computed(() => !!info.value && remaining.value <= 0)
const lowTime = computed(() => remaining.value > 0 && remaining.value < 5 * 60 * 1000)

const lastRejected = computed(() => {
  const slips = order.value?.slips ?? []
  const latest = [...slips].sort((a, b) => b.uploaded_at.localeCompare(a.uploaded_at))[0]
  return latest?.status === 'rejected' ? latest : null
})

const payErrorCode = computed(() => (isApiError(payError.value) ? payError.value.code : ''))

// Slip upload
const file = ref<File | null>(null)
const uploading = ref(false)
const drop = ref<InstanceType<typeof FileDrop> | null>(null)

async function upload() {
  if (!file.value || !order.value) return
  uploading.value = true
  try {
    const o = await ordersApi.uploadSlip(order.value.code, file.value)
    order.value = o
    info.value = null
    file.value = null
    drop.value?.clearPreview()
    ui.success(t('pay.slipUploaded'))
  } catch (e) {
    ui.error(errorText(e))
    if (isApiError(e) && e.code === 'not_payable') void load()
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div>
    <RouterLink :to="{ name: 'account-order', params: { code } }" class="back">
      <ArrowLeft :size="16" aria-hidden="true" />{{ $t('pay.orderCode', { code }) }}
    </RouterLink>
    <h1 class="h">{{ $t('pay.title') }}</h1>

    <div v-if="loading" class="grid" aria-busy="true">
      <SkeletonBlock height="420px" radius="var(--radius-lg)" />
      <SkeletonBlock height="300px" radius="var(--radius-lg)" />
    </div>
    <ErrorState v-else-if="error" :error="error" @retry="load" />

    <template v-else-if="order">
      <!-- Waiting for verification -->
      <div v-if="order.status === 'payment_review'" class="card card-pad">
        <EmptyState :icon="Hourglass" :title="$t('pay.waitingTitle')" :text="$t('pay.waitingText')">
          <BaseButton :to="{ name: 'account-order', params: { code: order.code } }">
            {{ $t('pay.viewOrder') }}
          </BaseButton>
        </EmptyState>
      </div>

      <!-- Not payable -->
      <div
        v-else-if="order.status !== 'pending_payment' || payErrorCode === 'not_payable'"
        class="card card-pad"
      >
        <EmptyState :title="$t('pay.notPayableTitle')">
          <p class="state">
            {{ $t('pay.notPayableText', { status: '' }) }}<StatusBadge :status="order.status" />
          </p>
          <BaseButton :to="{ name: 'account-order', params: { code: order.code } }">
            {{ $t('pay.viewOrder') }}
          </BaseButton>
        </EmptyState>
      </div>

      <!-- PromptPay not configured / other error -->
      <div v-else-if="payError" class="card card-pad">
        <EmptyState
          v-if="payErrorCode === 'promptpay_not_configured'"
          :icon="AlertTriangle"
          :title="$t('pay.notConfiguredTitle')"
          :text="$t('errors.promptpay_not_configured')"
        >
          <BaseButton :to="{ name: 'account-order', params: { code: order.code } }">
            {{ $t('pay.viewOrder') }}
          </BaseButton>
        </EmptyState>
        <ErrorState v-else :error="payError" @retry="load" />
      </div>

      <!-- Pay -->
      <div v-else-if="info" class="grid">
        <section class="card qr-card" aria-labelledby="h-qr">
          <div class="qr-head">
            <span class="pp-logo" aria-hidden="true">PromptPay</span>
            <h2 id="h-qr" class="sr-only">{{ $t('pay.scan') }}</h2>
          </div>
          <div class="qr-box">
            <img
              v-if="qrUrl"
              :src="qrUrl"
              :alt="$t('pay.qrAlt', { amount: formatBaht(info.amount_satang) })"
              class="qr"
              :class="{ faded: expired }"
              width="300"
              height="300"
            />
            <div v-else class="qr-loading" role="status">{{ $t('pay.generating') }}</div>
          </div>
          <p class="scan muted text-sm">{{ $t('pay.scan') }}</p>
          <dl class="facts">
            <div>
              <dt>{{ $t('pay.amount') }}</dt>
              <dd class="amount tabular">{{ formatBaht(info.amount_satang) }}</dd>
            </div>
            <div v-if="info.promptpay_name">
              <dt>{{ $t('pay.accountName') }}</dt>
              <dd>{{ info.promptpay_name }}</dd>
            </div>
          </dl>
          <div class="timer" :class="{ low: lowTime, over: expired }" role="timer" aria-live="off">
            <Timer :size="18" aria-hidden="true" />
            <template v-if="!expired">
              <span>{{ $t('pay.timeLeft') }}</span>
              <strong class="tabular">{{ formatCountdown(remaining) }}</strong>
            </template>
            <span v-else>{{ $t('pay.expired') }}</span>
          </div>
          <p class="muted text-xs due">{{ $t('orders.dueBy', { time: dateTime(info.due_at) }) }}</p>
          <BaseButton
            v-if="qrUrl && !expired"
            :href="qrUrl"
            :download="`promptpay-${order.code}.png`"
            variant="secondary"
            size="sm"
          >
            <Download :size="16" aria-hidden="true" />{{ $t('pay.download') }}
          </BaseButton>
        </section>

        <section class="card card-pad" aria-labelledby="h-up">
          <p v-if="lastRejected" class="rejected" role="alert">
            <AlertTriangle :size="16" aria-hidden="true" />
            {{
              lastRejected.note
                ? $t('pay.rejected', { note: lastRejected.note })
                : $t('pay.rejectedNoNote')
            }}
          </p>
          <ol class="steps">
            <li>{{ $t('pay.step1') }}</li>
            <li>{{ $t('pay.step2') }}</li>
            <li>{{ $t('pay.step3') }}</li>
          </ol>
          <h2 id="h-up" class="sec">{{ $t('pay.uploadTitle') }}</h2>
          <FileDrop
            ref="drop"
            :label="$t('pay.chooseFile')"
            :hint="$t('pay.uploadHint')"
            :disabled="uploading || expired"
            @select="file = $event"
          />
          <BaseButton
            block
            size="lg"
            class="send"
            :loading="uploading"
            :disabled="!file || expired"
            @click="upload"
          >
            <Upload :size="18" aria-hidden="true" />
            {{ uploading ? $t('pay.uploading') : $t('pay.submit') }}
          </BaseButton>
        </section>
      </div>
    </template>
  </div>
</template>

<style scoped>
.back {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-sm);
  margin-bottom: var(--sp-3);
}
.h {
  font-size: var(--fs-xl);
}
.grid {
  display: grid;
  gap: var(--sp-4);
}
@media (min-width: 900px) {
  .grid {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    align-items: start;
  }
}
.qr-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 0 0 var(--sp-6);
  overflow: hidden;
  gap: var(--sp-2);
}
.qr-head {
  width: 100%;
  background: #0e3d6b;
  padding: var(--sp-3);
  margin-bottom: var(--sp-4);
}
.pp-logo {
  color: #fff;
  font-weight: 800;
  letter-spacing: 0.02em;
  font-size: var(--fs-lg);
  font-family: 'Inter Variable', system-ui, sans-serif;
}
.qr-box {
  width: min(300px, 78vw);
  aspect-ratio: 1;
  display: grid;
  place-items: center;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-md);
  background: #fff;
}
.qr {
  width: 100%;
  height: auto;
  image-rendering: pixelated;
}
.faded {
  opacity: 0.2;
  filter: grayscale(1);
}
.qr-loading {
  color: var(--c-muted);
  font-size: var(--fs-sm);
}
.scan {
  margin: 0;
}
.facts {
  margin: var(--sp-2) 0 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-1);
}
.facts dt {
  font-size: var(--fs-xs);
  color: var(--c-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.facts dd {
  margin: 0;
  font-weight: var(--fw-semibold);
}
.amount {
  font-size: var(--fs-3xl);
  font-weight: var(--fw-bold) !important;
  color: var(--c-primary);
  line-height: 1.2;
}
.timer {
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-2) var(--sp-4);
  border-radius: var(--radius-full);
  background: var(--c-primary-50);
  color: var(--c-primary);
  font-size: var(--fs-sm);
  margin-top: var(--sp-2);
}
.timer strong {
  font-size: var(--fs-md);
}
.timer.low {
  background: var(--c-warning-bg);
  color: var(--c-warning);
}
.timer.over {
  background: var(--c-danger-bg);
  color: var(--c-danger);
}
.due {
  margin: 0 0 var(--sp-2);
}
.steps {
  margin: 0 0 var(--sp-5);
  padding-left: var(--sp-5);
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
  font-size: var(--fs-sm);
  color: var(--c-text-2);
}
.sec {
  font-size: var(--fs-md);
}
.send {
  margin-top: var(--sp-4);
}
.rejected {
  display: flex;
  gap: var(--sp-2);
  align-items: flex-start;
  padding: var(--sp-3);
  border-radius: var(--radius-md);
  background: var(--c-danger-bg);
  color: var(--c-danger);
  font-size: var(--fs-sm);
  margin: 0 0 var(--sp-4);
}
.rejected :deep(svg) {
  flex: none;
  margin-top: 3px;
}
.state {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  justify-content: center;
  margin: 0 0 var(--sp-3);
  color: var(--c-text-2);
}
</style>
