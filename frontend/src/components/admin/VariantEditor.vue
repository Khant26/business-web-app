<script setup lang="ts">
/** Variants table + add/edit modal for one product. */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import MoneyInput from './MoneyInput.vue'
import DataTable from './DataTable.vue'
import {
  intField,
  optionalSatangField,
  optionalText,
  requiredText,
  satangField,
} from './formFields'
import { adminApi } from '@/api/admin'
import { useForm, msg } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import { formatBaht } from '@/utils/money'
import type { AdminVariant } from '@/types/api'
import type { Column } from './types'

const variants = defineModel<AdminVariant[]>({ required: true })
const props = defineProps<{ productId: number }>()
const emit = defineEmits<{ changed: [] }>()

const { t } = useI18n()
const ui = useUiStore()
const { errorText } = useErrorMessage()

const schema = z
  .object({
    sku: requiredText(64),
    name_en: requiredText(120),
    name_my: optionalText(120),
    price_satang: satangField(),
    compare_at_satang: optionalSatangField(),
    low_stock_threshold: intField({ min: 0, max: 100000 }),
    sort_order: intField({ min: -100000, max: 100000 }),
    stock: intField({ min: 0, max: 1000000 }),
    is_active: z.boolean(),
  })
  .refine((v) => v.compare_at_satang === null || v.compare_at_satang > v.price_satang, {
    path: ['compare_at_satang'],
    message: msg('admin.products.comparePriceError'),
  })

const form = useForm(schema, {
  sku: '',
  name_en: '',
  name_my: '',
  price_satang: NaN,
  compare_at_satang: null,
  low_stock_threshold: '5',
  sort_order: '0',
  stock: '0',
  is_active: true,
})

const open = ref(false)
const editing = ref<AdminVariant | null>(null)
const formError = ref('')

function openNew() {
  editing.value = null
  formError.value = ''
  const nextSort = variants.value.reduce((m, v) => Math.max(m, v.sort_order), -1) + 1
  form.reset({
    sku: '',
    name_en: '',
    name_my: '',
    price_satang: NaN,
    compare_at_satang: null,
    low_stock_threshold: '5',
    sort_order: String(nextSort),
    stock: '0',
    is_active: true,
  })
  open.value = true
}

function openEdit(v: AdminVariant) {
  editing.value = v
  formError.value = ''
  form.reset({
    sku: v.sku,
    name_en: v.name_en,
    name_my: v.name_my,
    price_satang: v.price_satang,
    compare_at_satang: v.compare_at_satang,
    low_stock_threshold: String(v.low_stock_threshold),
    sort_order: String(v.sort_order),
    stock: String(v.stock),
    is_active: v.is_active,
  })
  open.value = true
}

function sortList(list: AdminVariant[]) {
  return [...list].sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
}

async function save() {
  const data = form.validate()
  if (!data) return
  form.submitting.value = true
  formError.value = ''
  const body = {
    sku: data.sku,
    name_en: data.name_en,
    name_my: data.name_my,
    price_satang: data.price_satang,
    compare_at_satang: data.compare_at_satang,
    low_stock_threshold: data.low_stock_threshold,
    is_active: data.is_active,
    sort_order: data.sort_order,
  }
  try {
    if (editing.value) {
      const updated = await adminApi.updateVariant(editing.value.id, body)
      variants.value = sortList(variants.value.map((v) => (v.id === updated.id ? updated : v)))
    } else {
      const created = await adminApi.createVariant(props.productId, { ...body, stock: data.stock })
      variants.value = sortList([...variants.value, created])
    }
    open.value = false
    ui.success(t('admin.products.variantSaved'))
    emit('changed')
  } catch (e) {
    if (!form.applyApiError(e)) formError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}

const toRemove = ref<AdminVariant | null>(null)
const confirmOpen = ref(false)
const removing = ref(false)

function askRemove(v: AdminVariant) {
  toRemove.value = v
  confirmOpen.value = true
}

async function remove() {
  const v = toRemove.value
  if (!v) return
  removing.value = true
  try {
    await adminApi.deleteVariant(v.id)
    // Backend deactivates variants with history; reload so the list reflects that.
    const fresh = await adminApi.product(props.productId)
    variants.value = sortList(fresh.variants)
    confirmOpen.value = false
    ui.success(t('admin.products.variantRemoved'))
    emit('changed')
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    removing.value = false
  }
}

const columns = computed<Column[]>(() => [
  { key: 'name_en', label: t('admin.products.variants'), primary: true },
  { key: 'sku', label: t('admin.products.sku') },
  { key: 'price_satang', label: t('admin.products.price'), align: 'right' },
  { key: 'stock', label: t('admin.products.stock'), align: 'right' },
  { key: 'is_active', label: t('admin.common.status') },
  { key: 'actions', label: t('admin.common.actions'), align: 'right' },
])
</script>

<template>
  <div class="ve">
    <DataTable v-if="variants.length" :columns="columns" :rows="variants" :row-key="(v) => v.id">
      <template #cell-name_en="{ row }">
        <span class="txt">
          <span class="nm">{{ row.name_en }}</span>
          <span v-if="row.name_my" class="sub" lang="my">{{ row.name_my }}</span>
        </span>
      </template>
      <template #cell-sku="{ row }"
        ><span class="admin-mono">{{ row.sku }}</span></template
      >
      <template #cell-price_satang="{ row }">
        <span class="price tabular">
          {{ formatBaht(row.price_satang) }}
          <s v-if="row.compare_at_satang" class="muted text-xs">{{
            formatBaht(row.compare_at_satang)
          }}</s>
        </span>
      </template>
      <template #cell-stock="{ row }">
        <span class="tabular" :class="{ 'text-danger': row.stock <= row.low_stock_threshold }">{{
          row.stock
        }}</span>
      </template>
      <template #cell-is_active="{ row }">
        <BaseBadge :tone="row.is_active ? 'success' : 'neutral'" size="sm">
          {{ row.is_active ? $t('admin.common.active') : $t('admin.products.inactive') }}
        </BaseBadge>
      </template>
      <template #cell-actions="{ row }">
        <span class="admin-row-actions">
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.edit')} ${row.sku}`"
            @click="openEdit(row)"
          >
            <Pencil :size="16" aria-hidden="true" />
          </BaseButton>
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.products.deleteVariant')} ${row.sku}`"
            @click="askRemove(row)"
          >
            <Trash2 :size="16" aria-hidden="true" />
          </BaseButton>
        </span>
      </template>
    </DataTable>
    <p v-else class="none">{{ $t('admin.products.noVariants') }}</p>

    <BaseButton variant="secondary" class="add" @click="openNew">
      <Plus :size="18" aria-hidden="true" />{{ $t('admin.products.addVariant') }}
    </BaseButton>

    <BaseModal
      v-model:open="open"
      size="lg"
      :title="editing ? $t('admin.products.editVariant') : $t('admin.products.newVariant')"
    >
      <form id="variant-form" class="grid-form cols-2" novalidate @submit.prevent="save">
        <BaseInput
          v-model="form.values.sku"
          :label="$t('admin.products.sku')"
          :error="form.errors.value.sku"
          required
          autocomplete="off"
        />
        <BaseInput
          v-model="form.values.sort_order"
          type="number"
          inputmode="numeric"
          :label="$t('admin.common.sortOrder')"
          :error="form.errors.value.sort_order"
        />
        <BaseInput
          v-model="form.values.name_en"
          :label="$t('admin.products.variantNameEn')"
          :hint="$t('admin.products.variantNameHint')"
          :error="form.errors.value.name_en"
          required
        />
        <BaseInput
          v-model="form.values.name_my"
          lang="my"
          :label="$t('admin.products.variantNameMy')"
          :error="form.errors.value.name_my"
          optional
        />
        <MoneyInput
          v-model="form.values.price_satang"
          :label="$t('admin.products.priceBaht')"
          :error="form.errors.value.price_satang"
          required
        />
        <MoneyInput
          v-model="form.values.compare_at_satang"
          nullable
          :label="$t('admin.products.compareAtBaht')"
          :hint="$t('admin.products.compareAtHint')"
          :error="form.errors.value.compare_at_satang"
          optional
        />
        <BaseInput
          v-model="form.values.low_stock_threshold"
          type="number"
          inputmode="numeric"
          min="0"
          :label="$t('admin.products.lowStockThreshold')"
          :error="form.errors.value.low_stock_threshold"
        />
        <BaseInput
          v-if="!editing"
          v-model="form.values.stock"
          type="number"
          inputmode="numeric"
          min="0"
          :label="$t('admin.products.initialStock')"
          :error="form.errors.value.stock"
        />
        <p v-else class="muted text-sm stock-hint">{{ $t('admin.products.stockAdjustHint') }}</p>
        <div class="span-2">
          <BaseCheckbox
            v-model="form.values.is_active"
            toggle
            :label="$t('admin.products.variantActive')"
          />
        </div>
        <p v-if="formError" class="span-2 text-danger text-sm" role="alert">{{ formError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="open = false">{{ $t('common.cancel') }}</BaseButton>
        <BaseButton type="submit" form="variant-form" :loading="form.submitting.value">
          {{ $t('common.save') }}
        </BaseButton>
      </template>
    </BaseModal>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('admin.products.deleteVariantTitle')"
      :message="$t('admin.products.deleteVariantText')"
      :confirm-label="$t('admin.products.deleteVariant')"
      danger
      :loading="removing"
      @confirm="remove"
    />
  </div>
</template>

<style scoped>
.txt {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.nm {
  font-weight: var(--fw-semibold);
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.price {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-end;
}
.none {
  color: var(--c-muted);
  font-size: var(--fs-sm);
  background: var(--c-surface-2);
  border-radius: var(--radius-md);
  padding: var(--sp-4);
  margin: 0;
}
.add {
  margin-top: var(--sp-3);
}
.stock-hint {
  align-self: end;
  margin: 0 0 var(--sp-2);
}
</style>
