<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { CheckCircle2, ExternalLink, Info, PackageCheck, Truck, XCircle } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { adminApi } from '@/api/admin'
import { useAsync } from '@/composables/useAsync'
import { useForm, msg } from '@/composables/useForm'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { usePageTitle } from '@/composables/usePageTitle'
import { useUiStore } from '@/stores/ui'
import { formatBaht, formatBp } from '@/utils/money'
import type { AdminOrder, AdminSlip, OrderStatus } from '@/types/api'

const props = defineProps<{ code: string }>()
const { t } = useI18n()
const { dateTime } = useLocalized()
const { errorText } = useErrorMessage()
const ui = useUiStore()

const order = useAsync(() => adminApi.order(props.code))
onMounted(() => void order.run())
usePageTitle(() => props.code)

const o = computed(() => order.data.value)

function setOrder(next: AdminOrder) {
  order.set(next)
}

// ---------- Status transitions ----------
const FINAL: OrderStatus[] = ['shipped', 'delivered', 'cancelled', 'expired']
const CLOSED: OrderStatus[] = ['delivered', 'cancelled', 'expired']

interface Action {
  to: OrderStatus
  label: string
  disabled?: boolean
  hint?: string
}

const nextAction = computed<Action | null>(() => {
  const s = o.value?.status
  if (s === 'paid' || s === 'confirmed')
    return { to: 'packing', label: t('admin.orders.toPacking') }
  if (s === 'packing')
    return {
      to: 'shipped',
      label: t('admin.orders.toShipped'),
      disabled: !o.value?.tracking_number,
      hint: o.value?.tracking_number ? undefined : t('admin.orders.needsTracking'),
    }
  if (s === 'shipped') return { to: 'delivered', label: t('admin.orders.toDelivered') }
  return null
})
const canCancel = computed(() => !!o.value && !FINAL.includes(o.value.status))
const statusHint = computed(() => {
  const s = o.value?.status
  if (s === 'pending_payment') return t('admin.orders.awaitingSlip')
  if (s === 'payment_review') return t('admin.orders.reviewSlipHint')
  if (s && CLOSED.includes(s)) return t('admin.orders.finalState')
  return ''
})

const statusModal = ref(false)
const pendingStatus = ref<OrderStatus | null>(null)
const statusNote = ref('')
const statusError = ref('')
const statusBusy = ref(false)

function openStatus(to: OrderStatus) {
  pendingStatus.value = to
  statusNote.value = ''
  statusError.value = ''
  statusModal.value = true
}

async function submitStatus() {
  if (!pendingStatus.value || !o.value) return
  statusBusy.value = true
  statusError.value = ''
  try {
    const next = await adminApi.setOrderStatus(
      o.value.code,
      pendingStatus.value,
      statusNote.value.trim() || undefined,
    )
    setOrder(next)
    statusModal.value = false
    ui.success(t('admin.orders.statusUpdated'))
  } catch (e) {
    statusError.value = errorText(e)
  } finally {
    statusBusy.value = false
  }
}

// ---------- Tracking ----------
const trackingSchema = z.object({
  courier: z
    .string()
    .trim()
    .min(1, msg('validation.required'))
    .max(100, msg('validation.maxLength', { max: 100 })),
  tracking_number: z
    .string()
    .trim()
    .min(1, msg('validation.required'))
    .max(100, msg('validation.maxLength', { max: 100 })),
  tracking_url: z
    .string()
    .trim()
    .refine((v) => v === '' || /^https?:\/\/\S+$/i.test(v), msg('validation.url')),
})
const tracking = useForm(trackingSchema, { courier: '', tracking_number: '', tracking_url: '' })

watch(
  o,
  (v) => {
    if (v)
      tracking.reset({
        courier: v.courier,
        tracking_number: v.tracking_number,
        tracking_url: v.tracking_url,
      })
  },
  { immediate: true },
)

const showTracking = computed(
  () =>
    !!o.value &&
    !['cancelled', 'expired', 'pending_payment', 'payment_review'].includes(o.value.status),
)

async function saveTracking() {
  const data = tracking.validate()
  if (!data || !o.value) return
  tracking.submitting.value = true
  try {
    const next = await adminApi.setTracking(o.value.code, {
      courier: data.courier,
      tracking_number: data.tracking_number,
      tracking_url: data.tracking_url || undefined,
    })
    setOrder(next)
    ui.success(t('admin.orders.trackingSaved'))
  } catch (e) {
    if (!tracking.applyApiError(e)) ui.error(errorText(e))
  } finally {
    tracking.submitting.value = false
  }
}

// ---------- Slips ----------
const viewer = ref<AdminSlip | null>(null)
const viewerOpen = ref(false)
function openViewer(s: AdminSlip) {
  viewer.value = s
  viewerOpen.value = true
}

const slipModal = ref(false)
const slipAction = ref<'approve' | 'reject'>('approve')
const slipTarget = ref<AdminSlip | null>(null)
const slipNote = ref('')
const slipError = ref('')
const slipBusy = ref(false)

function openSlip(s: AdminSlip, action: 'approve' | 'reject') {
  slipTarget.value = s
  slipAction.value = action
  slipNote.value = ''
  slipError.value = ''
  slipModal.value = true
  viewerOpen.value = false
}

async function submitSlip() {
  const s = slipTarget.value
  if (!s) return
  const note = slipNote.value.trim()
  if (slipAction.value === 'reject' && !note) {
    slipError.value = t('admin.orders.rejectNoteRequired')
    return
  }
  slipBusy.value = true
  slipError.value = ''
  try {
    const next =
      slipAction.value === 'approve'
        ? await adminApi.approveSlip(s.id, note || undefined)
        : await adminApi.rejectSlip(s.id, note)
    setOrder(next)
    slipModal.value = false
    ui.success(
      slipAction.value === 'approve'
        ? t('admin.orders.slipApproved')
        : t('admin.orders.slipRejected'),
    )
  } catch (e) {
    slipError.value = errorText(e)
  } finally {
    slipBusy.value = false
  }
}

const timeline = computed(() => [...(o.value?.history ?? [])].reverse())
</script>

<template>
  <div>
    <AdminPageHeader :title="code" :back="{ name: 'admin-orders' }">
      <template #meta>
        <div v-if="o" class="meta">
          <StatusBadge :status="o.status" />
          <BaseBadge tone="neutral">{{ $t(`payment.${o.payment_method}`) }}</BaseBadge>
          <span class="muted text-sm">{{ dateTime(o.created_at) }}</span>
        </div>
      </template>
    </AdminPageHeader>

    <ErrorState v-if="order.error.value" :error="order.error.value" @retry="order.retry" />

    <div v-else-if="!o" class="admin-cols">
      <div class="stack">
        <div class="card card-pad"><SkeletonBlock :lines="6" height="20px" /></div>
        <div class="card card-pad"><SkeletonBlock :lines="4" height="20px" /></div>
      </div>
      <div class="card card-pad"><SkeletonBlock :lines="5" height="20px" /></div>
    </div>

    <div v-else class="admin-cols">
      <div class="stack">
        <!-- Items + totals -->
        <section class="card card-pad">
          <h2 class="admin-section-title">{{ $t('admin.orders.orderItems') }}</h2>
          <ul class="items">
            <li v-for="it in o.items" :key="it.id" class="item">
              <ImagePlaceholder
                class="thumb"
                :src="it.image_url"
                :name="it.product_name_en"
                small
              />
              <div class="info">
                <RouterLink
                  v-if="it.product_id"
                  :to="{ name: 'admin-product-edit', params: { id: it.product_id } }"
                  class="nm"
                >
                  {{ it.product_name_en }}
                </RouterLink>
                <span v-else class="nm">{{ it.product_name_en }}</span>
                <span class="sub"
                  >{{ it.variant_name_en }} · <span class="admin-mono">{{ it.sku }}</span></span
                >
                <span class="sub tabular"
                  >{{ formatBaht(it.unit_price_satang) }} × {{ it.quantity }}</span
                >
              </div>
              <strong class="tabular">{{ formatBaht(it.line_total_satang) }}</strong>
            </li>
          </ul>
          <div class="totals">
            <div class="sum-row">
              <span>{{ $t('admin.orders.subtotal') }}</span
              ><span>{{ formatBaht(o.subtotal_satang) }}</span>
            </div>
            <div v-if="o.discount_satang > 0" class="sum-row discount">
              <span>{{
                o.coupon_code
                  ? $t('admin.orders.discountCode', { code: o.coupon_code })
                  : $t('admin.orders.discount')
              }}</span>
              <span>−{{ formatBaht(o.discount_satang) }}</span>
            </div>
            <div class="sum-row">
              <span>{{ $t('admin.orders.shipping') }}</span
              ><span>{{ formatBaht(o.shipping_satang) }}</span>
            </div>
            <div class="sum-row">
              <span>{{ $t('admin.orders.vat', { rate: formatBp(o.vat_rate_bp) }) }}</span>
              <span>{{ formatBaht(o.vat_satang) }}</span>
            </div>
            <div class="sum-row total">
              <span>{{ $t('admin.orders.total') }}</span
              ><span>{{ formatBaht(o.total_satang) }}</span>
            </div>
          </div>
        </section>

        <!-- Slips -->
        <section v-if="o.payment_method === 'promptpay'" class="card card-pad">
          <h2 class="admin-section-title">{{ $t('admin.orders.slips') }}</h2>
          <p v-if="!o.slips.length" class="muted text-sm">{{ $t('admin.orders.noSlips') }}</p>
          <ul v-else class="slips">
            <li v-for="s in o.slips" :key="s.id" class="slip">
              <button
                type="button"
                class="slip-thumb"
                :aria-label="$t('admin.orders.viewSlip')"
                @click="openViewer(s)"
              >
                <img
                  :src="adminApi.slipFileUrl(s.id)"
                  :alt="$t('admin.orders.slipImage', { id: s.id })"
                  loading="lazy"
                />
              </button>
              <div class="slip-info">
                <StatusBadge :status="s.status" kind="slip" />
                <span class="sub">{{
                  $t('admin.orders.uploadedAt', { date: dateTime(s.uploaded_at) })
                }}</span>
                <span v-if="s.reviewed_at" class="sub">
                  {{ $t('admin.orders.reviewedAt', { date: dateTime(s.reviewed_at) }) }}
                </span>
                <p v-if="s.note" class="note">{{ s.note }}</p>
                <div v-if="s.status === 'pending'" class="row">
                  <BaseButton size="sm" @click="openSlip(s, 'approve')">
                    <CheckCircle2 :size="16" aria-hidden="true" />{{ $t('admin.orders.approve') }}
                  </BaseButton>
                  <BaseButton size="sm" variant="secondary" @click="openSlip(s, 'reject')">
                    <XCircle :size="16" aria-hidden="true" />{{ $t('admin.orders.reject') }}
                  </BaseButton>
                </div>
              </div>
            </li>
          </ul>
        </section>

        <!-- History -->
        <section class="card card-pad">
          <h2 class="admin-section-title">{{ $t('admin.orders.history') }}</h2>
          <ol class="timeline">
            <li v-for="(h, i) in timeline" :key="i">
              <span class="dot" :class="{ first: i === 0 }" aria-hidden="true" />
              <div>
                <StatusBadge :status="h.status" />
                <span class="sub when">{{ dateTime(h.created_at) }}</span>
                <p v-if="h.note" class="note">{{ h.note }}</p>
              </div>
            </li>
          </ol>
        </section>
      </div>

      <div class="stack">
        <!-- Status actions -->
        <section class="card card-pad">
          <h2 class="admin-section-title">{{ $t('admin.orders.statusActions') }}</h2>
          <p v-if="statusHint" class="hint">
            <Info :size="16" aria-hidden="true" />{{ statusHint }}
          </p>
          <div class="actions">
            <BaseButton
              v-if="nextAction"
              block
              :disabled="nextAction.disabled"
              @click="openStatus(nextAction.to)"
            >
              <PackageCheck v-if="nextAction.to === 'packing'" :size="18" aria-hidden="true" />
              <Truck v-else-if="nextAction.to === 'shipped'" :size="18" aria-hidden="true" />
              <CheckCircle2 v-else :size="18" aria-hidden="true" />
              {{ nextAction.label }}
            </BaseButton>
            <p v-if="nextAction?.hint" class="warn text-xs">{{ nextAction.hint }}</p>
            <BaseButton
              v-if="canCancel"
              block
              variant="secondary"
              class="cancel"
              @click="openStatus('cancelled')"
            >
              <XCircle :size="18" aria-hidden="true" />{{ $t('admin.orders.cancel') }}
            </BaseButton>
          </div>
          <dl class="admin-dl dates">
            <template v-if="o.payment_due_at && o.status === 'pending_payment'">
              <dt>{{ $t('admin.orders.paymentDue') }}</dt>
              <dd>{{ dateTime(o.payment_due_at) }}</dd>
            </template>
            <template v-if="o.paid_at"
              ><dt>{{ $t('admin.orders.paidAt') }}</dt>
              <dd>{{ dateTime(o.paid_at) }}</dd></template
            >
            <template v-if="o.shipped_at">
              <dt>{{ $t('admin.orders.shippedAt') }}</dt>
              <dd>{{ dateTime(o.shipped_at) }}</dd>
            </template>
            <template v-if="o.delivered_at">
              <dt>{{ $t('admin.orders.deliveredAt') }}</dt>
              <dd>{{ dateTime(o.delivered_at) }}</dd>
            </template>
            <template v-if="o.cancelled_at">
              <dt>{{ $t('admin.orders.cancelledAt') }}</dt>
              <dd>
                {{ dateTime(o.cancelled_at)
                }}<template v-if="o.cancel_reason"> — {{ o.cancel_reason }}</template>
              </dd>
            </template>
          </dl>
        </section>

        <!-- Tracking -->
        <section v-if="showTracking" class="card card-pad">
          <h2 class="admin-section-title">{{ $t('admin.orders.tracking') }}</h2>
          <form class="grid-form" novalidate @submit.prevent="saveTracking">
            <BaseInput
              v-model="tracking.values.courier"
              :label="$t('admin.orders.courier')"
              :placeholder="$t('admin.orders.courierPlaceholder')"
              :error="tracking.errors.value.courier"
              required
              :disabled="o.status === 'delivered'"
            />
            <BaseInput
              v-model="tracking.values.tracking_number"
              :label="$t('admin.orders.trackingNumber')"
              :error="tracking.errors.value.tracking_number"
              required
              autocomplete="off"
              :disabled="o.status === 'delivered'"
            />
            <BaseInput
              v-model="tracking.values.tracking_url"
              type="url"
              :label="$t('admin.orders.trackingUrl')"
              :error="tracking.errors.value.tracking_url"
              placeholder="https://"
              optional
              :disabled="o.status === 'delivered'"
            />
            <div class="row">
              <BaseButton
                v-if="o.status !== 'delivered'"
                type="submit"
                variant="secondary"
                :loading="tracking.submitting.value"
              >
                {{ $t('admin.orders.saveTracking') }}
              </BaseButton>
              <a
                v-if="o.tracking_url"
                :href="o.tracking_url"
                target="_blank"
                rel="noopener noreferrer"
                class="ext"
              >
                {{ $t('admin.orders.openTracking') }}<ExternalLink :size="14" aria-hidden="true" />
              </a>
            </div>
          </form>
        </section>

        <!-- Customer -->
        <template v-if="o">
          <section class="card card-pad">
            <h2 class="admin-section-title">{{ $t('admin.orders.customer') }}</h2>
            <dl class="admin-dl">
              <dt>{{ $t('admin.common.name') }}</dt>
              <dd>{{ o.customer.name || '—' }}</dd>
              <dt>{{ $t('admin.common.email') }}</dt>
              <dd>
                <a :href="`mailto:${o.email || o.customer.email}`">{{
                  o.email || o.customer.email
                }}</a>
              </dd>
              <dt>{{ $t('admin.common.phone') }}</dt>
              <dd>{{ o.customer.phone || '—' }}</dd>
            </dl>
          </section>

          <section class="card card-pad">
            <h2 class="admin-section-title">{{ $t('admin.orders.shippingAddress') }}</h2>
            <address class="addr">
              <strong>{{ o.shipping_address.recipient_name }}</strong
              ><br />
              <a :href="`tel:${o.shipping_address.phone}`">{{ o.shipping_address.phone }}</a
              ><br />
              {{ o.shipping_address.line1 }}<br />
              {{ o.shipping_address.subdistrict }}, {{ o.shipping_address.district }}<br />
              {{ o.shipping_address.province }} {{ o.shipping_address.postcode }}
            </address>
            <template v-if="o.note">
              <h3 class="note-title">{{ $t('admin.orders.note') }}</h3>
              <p class="note">{{ o.note }}</p>
            </template>
          </section>
        </template>
      </div>
    </div>

    <!-- Status confirm -->
    <BaseModal
      v-model:open="statusModal"
      size="sm"
      :title="
        pendingStatus === 'cancelled'
          ? $t('admin.orders.cancelTitle')
          : pendingStatus
            ? $t(`status.${pendingStatus}`)
            : ''
      "
    >
      <form id="status-form" class="stack" @submit.prevent="submitStatus">
        <p v-if="pendingStatus === 'cancelled'" class="muted">
          {{ $t('admin.orders.cancelText') }}
        </p>
        <BaseInput
          v-model="statusNote"
          multiline
          :rows="3"
          :label="
            pendingStatus === 'cancelled'
              ? $t('admin.orders.cancelReason')
              : $t('admin.orders.statusNote')
          "
          optional
          maxlength="500"
        />
        <p v-if="statusError" class="text-danger text-sm" role="alert">{{ statusError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="statusModal = false">{{
          $t('common.cancel')
        }}</BaseButton>
        <BaseButton
          type="submit"
          form="status-form"
          :variant="pendingStatus === 'cancelled' ? 'danger' : 'primary'"
          :loading="statusBusy"
        >
          {{ $t('common.confirm') }}
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Slip approve / reject -->
    <BaseModal
      v-model:open="slipModal"
      size="sm"
      :title="
        slipAction === 'approve' ? $t('admin.orders.approveTitle') : $t('admin.orders.rejectTitle')
      "
    >
      <form id="slip-form" class="stack" @submit.prevent="submitSlip">
        <img
          v-if="slipTarget"
          :src="adminApi.slipFileUrl(slipTarget.id)"
          :alt="$t('admin.orders.slipImage', { id: slipTarget.id })"
          class="slip-preview"
        />
        <BaseInput
          v-model="slipNote"
          multiline
          :rows="3"
          :label="$t('admin.orders.slipNote')"
          :hint="$t('admin.orders.slipNoteHint')"
          :required="slipAction === 'reject'"
          :optional="slipAction === 'approve'"
          :error="slipError"
          maxlength="500"
        />
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="slipModal = false">{{
          $t('common.cancel')
        }}</BaseButton>
        <BaseButton
          type="submit"
          form="slip-form"
          :variant="slipAction === 'approve' ? 'primary' : 'danger'"
          :loading="slipBusy"
        >
          {{ slipAction === 'approve' ? $t('admin.orders.approve') : $t('admin.orders.reject') }}
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Slip viewer -->
    <BaseModal
      v-model:open="viewerOpen"
      size="lg"
      :title="viewer ? $t('admin.orders.slipImage', { id: viewer.id }) : ''"
    >
      <div v-if="viewer" class="viewer">
        <img
          :src="adminApi.slipFileUrl(viewer.id)"
          :alt="$t('admin.orders.slipImage', { id: viewer.id })"
        />
        <a :href="adminApi.slipFileUrl(viewer.id)" target="_blank" rel="noopener" class="ext">
          {{ $t('admin.common.view') }}<ExternalLink :size="14" aria-hidden="true" />
        </a>
      </div>
      <template v-if="viewer?.status === 'pending'" #footer>
        <BaseButton variant="secondary" @click="openSlip(viewer, 'reject')">{{
          $t('admin.orders.reject')
        }}</BaseButton>
        <BaseButton @click="openSlip(viewer, 'approve')">{{
          $t('admin.orders.approve')
        }}</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<style scoped>
.meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--sp-2);
  margin-top: var(--sp-2);
}
.items {
  list-style: none;
  margin: 0;
  padding: 0;
}
.item {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-3) 0;
  border-bottom: 1px solid var(--c-border);
}
.thumb {
  width: 52px;
  flex: none;
}
.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.nm {
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.totals {
  margin-top: var(--sp-3);
  font-size: var(--fs-sm);
}
.discount {
  color: var(--c-success);
}
.slips {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-4);
}
.slip {
  display: flex;
  gap: var(--sp-4);
  align-items: flex-start;
}
.slip-thumb {
  flex: none;
  width: 96px;
  height: 128px;
  padding: 0;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-md);
  overflow: hidden;
  background: var(--c-surface-2);
  cursor: zoom-in;
}
.slip-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.slip-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  min-width: 0;
}
.note {
  margin: var(--sp-1) 0 0;
  font-size: var(--fs-sm);
  background: var(--c-surface-2);
  border-radius: var(--radius-sm);
  padding: var(--sp-2) var(--sp-3);
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}
.timeline li {
  display: flex;
  gap: var(--sp-3);
  position: relative;
  padding-bottom: var(--sp-4);
}
.timeline li:not(:last-child)::before {
  content: '';
  position: absolute;
  left: 5px;
  top: 14px;
  bottom: 0;
  width: 2px;
  background: var(--c-border);
}
.dot {
  flex: none;
  width: 12px;
  height: 12px;
  margin-top: 5px;
  border-radius: 50%;
  background: var(--c-border-strong);
}
.dot.first {
  background: var(--c-primary);
  box-shadow: 0 0 0 4px var(--c-primary-100);
}
.when {
  margin-left: var(--sp-2);
}
.hint {
  display: flex;
  gap: var(--sp-2);
  align-items: flex-start;
  font-size: var(--fs-sm);
  color: var(--c-text-2);
  background: var(--c-info-bg);
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  margin-bottom: var(--sp-3);
}
.hint svg {
  flex: none;
  margin-top: 3px;
  color: var(--c-info);
}
.actions {
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
.warn {
  margin: 0;
  color: var(--c-warning);
}
.cancel {
  color: var(--c-danger);
}
.dates {
  margin-top: var(--sp-4);
}
.dates:empty {
  display: none;
}
.addr {
  font-style: normal;
  font-size: var(--fs-sm);
  line-height: 1.7;
}
.note-title {
  font-size: var(--fs-sm);
  margin: var(--sp-4) 0 var(--sp-1);
}
.ext {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: var(--fs-sm);
}
.slip-preview {
  max-height: 260px;
  width: auto;
  margin-inline: auto;
  border-radius: var(--radius-md);
  border: 1px solid var(--c-border);
}
.viewer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--sp-3);
}
.viewer img {
  max-height: 70dvh;
  width: auto;
  border-radius: var(--radius-md);
}
</style>
