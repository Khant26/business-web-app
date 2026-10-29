<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import MoneyInput from '@/components/admin/MoneyInput.vue'
import {
  intField,
  optionalIntField,
  optionalSatangField,
  satangField,
} from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { adminApi } from '@/api/admin'
import { useAsync } from '@/composables/useAsync'
import { useForm, msg } from '@/composables/useForm'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import { fromLocalInput, toLocalInput } from '@/utils/date'
import { formatBaht } from '@/utils/money'
import type { Coupon, CouponInput, CouponType } from '@/types/api'
import type { Column } from '@/components/admin/types'
import type { SelectOption } from '@/types/ui'

const { t } = useI18n()
const ui = useUiStore()
const { date } = useLocalized()
const { errorText } = useErrorMessage()

const coupons = useAsync(() => adminApi.coupons())
onMounted(() => void coupons.run())
const rows = computed(() =>
  [...(coupons.data.value ?? [])].sort((a, b) => b.created_at.localeCompare(a.created_at)),
)

function discountLabel(c: Coupon) {
  return c.type === 'percent' ? `${c.value}%` : formatBaht(c.value)
}
function validity(c: Coupon) {
  if (!c.starts_at && !c.ends_at) return t('admin.coupons.always')
  return `${c.starts_at ? date(c.starts_at) : '…'} → ${c.ends_at ? date(c.ends_at) : '…'}`
}
function expired(c: Coupon) {
  return !!c.ends_at && new Date(c.ends_at).getTime() < Date.now()
}

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.coupons.code'), primary: true },
  { key: 'value', label: t('admin.coupons.value') },
  { key: 'used', label: t('admin.coupons.used'), align: 'right' },
  { key: 'validity', label: t('admin.coupons.validity') },
  { key: 'is_active', label: t('admin.common.status') },
  { key: 'actions', label: t('admin.common.actions'), align: 'right' },
])

function replaceRow(c: Coupon) {
  const list = coupons.data.value ?? []
  coupons.set(
    list.some((x) => x.id === c.id) ? list.map((x) => (x.id === c.id ? c : x)) : [...list, c],
  )
}

// ---------- Form ----------
const optionalDate = z
  .string()
  .refine((v) => v === '' || fromLocalInput(v) !== '', msg('validation.invalid'))

const schema = z
  .object({
    code: z
      .string()
      .trim()
      .toUpperCase()
      .min(1, msg('validation.required'))
      .max(40, msg('validation.maxLength', { max: 40 }))
      .regex(/^[A-Z0-9_-]+$/, msg('validation.invalid')),
    type: z.enum(['percent', 'fixed']),
    value_percent: z.union([z.string(), z.number()]),
    value_satang: z.number().nullable(),
    min_order_satang: satangField(),
    max_discount_satang: optionalSatangField(),
    usage_limit: optionalIntField({ min: 1, max: 10000000 }),
    per_user_limit: intField({ min: 1, max: 1000 }),
    starts_at: optionalDate,
    ends_at: optionalDate,
    is_active: z.boolean(),
  })
  .superRefine((v, ctx) => {
    if (v.type === 'percent') {
      const n = Number(v.value_percent)
      if (String(v.value_percent).trim() === '' || !Number.isInteger(n) || n < 1 || n > 100)
        ctx.addIssue({
          code: 'custom',
          path: ['value_percent'],
          message: msg('admin.validation.percentRange'),
        })
    } else if (
      v.value_satang === null ||
      !Number.isInteger(v.value_satang) ||
      v.value_satang <= 0
    ) {
      ctx.addIssue({ code: 'custom', path: ['value_satang'], message: msg('validation.positive') })
    }
    if (v.starts_at && v.ends_at && fromLocalInput(v.ends_at) <= fromLocalInput(v.starts_at))
      ctx.addIssue({
        code: 'custom',
        path: ['ends_at'],
        message: msg('admin.validation.endAfterStart'),
      })
  })

const blank = {
  code: '',
  type: 'percent' as CouponType,
  value_percent: '10',
  value_satang: null as number | null,
  min_order_satang: 0,
  max_discount_satang: null as number | null,
  usage_limit: '' as string | number | null,
  per_user_limit: '1',
  starts_at: '',
  ends_at: '',
  is_active: true,
}
const form = useForm(schema, { ...blank })
const open = ref(false)
const editing = ref<Coupon | null>(null)
const formError = ref('')

const typeOptions = computed<SelectOption<CouponType>[]>(() => [
  { value: 'percent', label: t('admin.coupons.percent') },
  { value: 'fixed', label: t('admin.coupons.fixed') },
])

function openNew() {
  editing.value = null
  formError.value = ''
  form.reset({ ...blank })
  open.value = true
}

function openEdit(c: Coupon) {
  editing.value = c
  formError.value = ''
  form.reset({
    code: c.code,
    type: c.type,
    value_percent: c.type === 'percent' ? String(c.value) : '10',
    value_satang: c.type === 'fixed' ? c.value : null,
    min_order_satang: c.min_order_satang,
    max_discount_satang: c.max_discount_satang,
    usage_limit: c.usage_limit === null ? '' : String(c.usage_limit),
    per_user_limit: String(c.per_user_limit),
    starts_at: toLocalInput(c.starts_at),
    ends_at: toLocalInput(c.ends_at),
    is_active: c.is_active,
  })
  open.value = true
}

async function save() {
  const d = form.validate()
  if (!d) return
  const body: CouponInput = {
    code: d.code,
    type: d.type,
    value: d.type === 'percent' ? Number(d.value_percent) : (d.value_satang ?? 0),
    min_order_satang: d.min_order_satang,
    max_discount_satang: d.type === 'percent' ? d.max_discount_satang : null,
    usage_limit: d.usage_limit,
    per_user_limit: d.per_user_limit,
    starts_at: d.starts_at ? fromLocalInput(d.starts_at) : null,
    ends_at: d.ends_at ? fromLocalInput(d.ends_at) : null,
    is_active: d.is_active,
  }
  form.submitting.value = true
  formError.value = ''
  try {
    const c = editing.value
      ? await adminApi.updateCoupon(editing.value.id, body)
      : await adminApi.createCoupon(body)
    replaceRow(c)
    open.value = false
    ui.success(editing.value ? t('admin.common.saved') : t('admin.common.created'))
  } catch (e) {
    if (form.applyApiError(e)) {
      // Map the API's `value` field onto whichever input is visible.
      const errs = { ...form.errors.value }
      if (errs.value) {
        errs[d.type === 'percent' ? 'value_percent' : 'value_satang'] = errs.value
        form.errors.value = errs
      }
    } else formError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}

// ---------- Delete ----------
const toDelete = ref<Coupon | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)
async function remove() {
  const c = toDelete.value
  if (!c) return
  deleting.value = true
  try {
    await adminApi.deleteCoupon(c.id)
    await coupons.run()
    confirmOpen.value = false
    ui.success(t('admin.common.deleted'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.coupons')">
      <BaseButton @click="openNew"
        ><Plus :size="18" aria-hidden="true" />{{ $t('admin.coupons.new') }}</BaseButton
      >
    </AdminPageHeader>

    <ErrorState v-if="coupons.error.value" :error="coupons.error.value" @retry="coupons.retry" />
    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :row-key="(c) => c.id"
      :loading="coupons.loading.value && !coupons.data.value"
      :caption="$t('admin.nav.coupons')"
    >
      <template #cell-code="{ row }">
        <button type="button" class="code" @click="openEdit(row)">{{ row.code }}</button>
      </template>
      <template #cell-value="{ row }">
        <span class="val">
          <strong>{{ discountLabel(row) }}</strong>
          <span v-if="row.min_order_satang > 0" class="sub">
            {{ $t('admin.coupons.minOrderShort', { amount: formatBaht(row.min_order_satang) }) }}
          </span>
        </span>
      </template>
      <template #cell-used="{ row }">
        <span class="tabular">{{ row.used_count }} / {{ row.usage_limit ?? '∞' }}</span>
      </template>
      <template #cell-validity="{ row }"
        ><span class="text-sm">{{ validity(row) }}</span></template
      >
      <template #cell-is_active="{ row }">
        <BaseBadge v-if="!row.is_active" size="sm">{{ $t('admin.common.inactive') }}</BaseBadge>
        <BaseBadge v-else-if="expired(row)" size="sm" tone="warning">{{
          $t('admin.events.ended')
        }}</BaseBadge>
        <BaseBadge v-else size="sm" tone="success">{{ $t('admin.common.active') }}</BaseBadge>
      </template>
      <template #cell-actions="{ row }">
        <span class="admin-row-actions">
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.edit')} ${row.code}`"
            @click="openEdit(row)"
          >
            <Pencil :size="16" aria-hidden="true" />
          </BaseButton>
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.delete')} ${row.code}`"
            @click="((toDelete = row), (confirmOpen = true))"
          >
            <Trash2 :size="16" aria-hidden="true" />
          </BaseButton>
        </span>
      </template>
      <template #empty>
        <EmptyState :title="$t('admin.coupons.empty')">
          <BaseButton @click="openNew">{{ $t('admin.coupons.new') }}</BaseButton>
        </EmptyState>
      </template>
    </DataTable>

    <BaseModal
      v-model:open="open"
      size="lg"
      :title="editing ? $t('admin.coupons.edit') : $t('admin.coupons.new')"
    >
      <form id="coupon-form" class="grid-form cols-2" novalidate @submit.prevent="save">
        <BaseInput
          v-model="form.values.code"
          :label="$t('admin.coupons.code')"
          :hint="$t('admin.coupons.codeHint')"
          :error="form.errors.value.code"
          required
          autocomplete="off"
          autocapitalize="characters"
          class="upper"
        />
        <BaseSelect
          v-model="form.values.type"
          :label="$t('admin.coupons.type')"
          :options="typeOptions"
        />
        <BaseInput
          v-if="form.values.type === 'percent'"
          v-model="form.values.value_percent"
          type="number"
          inputmode="numeric"
          min="1"
          max="100"
          :label="$t('admin.coupons.valuePercent')"
          :error="form.errors.value.value_percent"
          required
        />
        <MoneyInput
          v-else
          v-model="form.values.value_satang"
          nullable
          :label="$t('admin.coupons.valueBaht')"
          :error="form.errors.value.value_satang"
          required
        />
        <MoneyInput
          v-if="form.values.type === 'percent'"
          v-model="form.values.max_discount_satang"
          nullable
          :label="$t('admin.coupons.maxDiscount')"
          :hint="$t('admin.coupons.maxDiscountHint')"
          :error="form.errors.value.max_discount_satang"
          optional
        />
        <MoneyInput
          v-model="form.values.min_order_satang"
          :label="$t('admin.coupons.minOrder')"
          :error="form.errors.value.min_order_satang"
        />
        <BaseInput
          v-model="form.values.usage_limit"
          type="number"
          inputmode="numeric"
          min="1"
          :label="$t('admin.coupons.usageLimit')"
          :hint="$t('admin.coupons.usageLimitHint')"
          :error="form.errors.value.usage_limit"
          optional
        />
        <BaseInput
          v-model="form.values.per_user_limit"
          type="number"
          inputmode="numeric"
          min="1"
          :label="$t('admin.coupons.perUserLimit')"
          :error="form.errors.value.per_user_limit"
        />
        <BaseInput
          v-model="form.values.starts_at"
          type="datetime-local"
          :label="$t('admin.coupons.startsAt')"
          :error="form.errors.value.starts_at"
          optional
        />
        <BaseInput
          v-model="form.values.ends_at"
          type="datetime-local"
          :label="$t('admin.coupons.endsAt')"
          :hint="$t('admin.coupons.datesHint')"
          :error="form.errors.value.ends_at"
          optional
        />
        <div class="span-2">
          <BaseCheckbox v-model="form.values.is_active" toggle :label="$t('admin.common.active')" />
        </div>
        <p v-if="formError" class="span-2 text-danger text-sm" role="alert">{{ formError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="open = false">{{ $t('common.cancel') }}</BaseButton>
        <BaseButton type="submit" form="coupon-form" :loading="form.submitting.value">{{
          $t('common.save')
        }}</BaseButton>
      </template>
    </BaseModal>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('admin.common.confirmDelete')"
      :message="$t('admin.coupons.deleteText')"
      :confirm-label="$t('admin.common.delete')"
      danger
      :loading="deleting"
      @confirm="remove"
    />
  </div>
</template>

<style scoped>
.code {
  border: 1px dashed var(--c-primary-300);
  background: var(--c-primary-50);
  color: var(--c-primary-hover);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-weight: var(--fw-bold);
  padding: 2px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  letter-spacing: 0.04em;
}
.code:hover {
  background: var(--c-primary-100);
}
.val {
  display: inline-flex;
  flex-direction: column;
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.upper :deep(input) {
  text-transform: uppercase;
}
</style>
