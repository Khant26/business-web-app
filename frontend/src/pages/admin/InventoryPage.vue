<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { History, Search, SlidersHorizontal } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import { useAdminList } from '@/components/admin/useAdminList'
import { intField, optionalText } from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseTabs from '@/components/ui/BaseTabs.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { adminApi } from '@/api/admin'
import { usePaged } from '@/composables/usePaged'
import { useForm, msg } from '@/composables/useForm'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import type { InventoryRow, StockReason } from '@/types/api'
import type { Column } from '@/components/admin/types'
import type { SelectOption, TabItem } from '@/types/ui'

const { t } = useI18n()
const { dateTime } = useLocalized()
const { errorText } = useErrorMessage()
const ui = useUiStore()

const tab = ref<'stock' | 'movements'>('stock')
const tabs = computed<TabItem<'stock' | 'movements'>[]>(() => [
  { value: 'stock', label: t('admin.inventory.stockTab') },
  { value: 'movements', label: t('admin.inventory.movementsTab') },
])

// ---------- Stock levels ----------
const { filters, list, hasFilters, clear } = useAdminList(
  { q: '', low: false },
  (page, page_size, f) =>
    adminApi.inventory({ page, page_size, q: f.q.trim() || undefined, low: f.low }),
  { debounce: ['q'], pageSize: 25 },
)

const columns = computed<Column[]>(() => [
  { key: 'product', label: t('admin.inventory.product'), primary: true },
  { key: 'sku', label: t('admin.products.sku') },
  { key: 'stock', label: t('admin.inventory.stock'), align: 'right' },
  { key: 'low_stock_threshold', label: t('admin.inventory.threshold'), align: 'right' },
  { key: 'actions', label: t('admin.common.actions'), align: 'right' },
])

// ---------- Movements ----------
const movementVariant = ref<InventoryRow | null>(null)
const movements = usePaged(
  (page, page_size) =>
    adminApi.movements({ page, page_size, variant_id: movementVariant.value?.variant_id }),
  { pageSize: 25, immediate: false },
)
let movementsLoaded = false
watch(tab, (v) => {
  if (v === 'movements' && !movementsLoaded) {
    movementsLoaded = true
    void movements.load()
  }
})
function showHistory(row: InventoryRow) {
  movementVariant.value = row
  movementsLoaded = true
  tab.value = 'movements'
  movements.reload()
}
function showAllMovements() {
  movementVariant.value = null
  movements.reload()
}

const movementColumns = computed<Column[]>(() => [
  { key: 'product', label: t('admin.inventory.product'), primary: true },
  { key: 'change', label: t('admin.inventory.change'), align: 'right' },
  { key: 'stock_after', label: t('admin.inventory.stockAfter'), align: 'right' },
  { key: 'reason', label: t('admin.inventory.reason') },
  { key: 'by', label: t('admin.inventory.by') },
  { key: 'created_at', label: t('admin.common.date') },
])

// ---------- Adjust dialog ----------
const adjustTarget = ref<InventoryRow | null>(null)
const adjustOpen = ref(false)
const adjustError = ref('')

const schema = z.object({
  change: intField({ min: -1000000, max: 1000000 }).refine(
    (v) => v !== 0,
    msg('admin.inventory.nonZero'),
  ),
  reason: z.enum(['restock', 'adjust', 'damage']),
  note: optionalText(500),
})
const form = useForm(schema, { change: '', reason: 'restock', note: '' })

const reasonOptions = computed<SelectOption<StockReason>[]>(() =>
  (['restock', 'adjust', 'damage'] as StockReason[]).map((r) => ({
    value: r,
    label: t(`admin.inventory.reasons.${r}`),
  })),
)

const previewStock = computed(() => {
  const n = Number(form.values.change)
  if (!adjustTarget.value || !Number.isFinite(n) || String(form.values.change).trim() === '')
    return null
  return adjustTarget.value.stock + Math.trunc(n)
})

function openAdjust(row: InventoryRow) {
  adjustTarget.value = row
  adjustError.value = ''
  form.reset({ change: '', reason: 'restock', note: '' })
  adjustOpen.value = true
}

watch(
  () => form.values.reason,
  (r) => {
    // Damage always removes stock: flip a positive number to negative.
    const n = Number(form.values.change)
    if (r === 'damage' && n > 0) form.values.change = String(-n)
  },
)

async function submitAdjust() {
  const row = adjustTarget.value
  const data = form.validate()
  if (!data || !row) return
  if (row.stock + data.change < 0) {
    form.errors.value = { ...form.errors.value, change: t('admin.inventory.negative') }
    return
  }
  form.submitting.value = true
  adjustError.value = ''
  try {
    const updated = await adminApi.adjustStock(row.variant_id, {
      change: data.change,
      reason: data.reason,
      note: data.note || undefined,
    })
    const d = list.data.value
    if (d) {
      list.data.value = {
        ...d,
        items: d.items.map((r) => (r.variant_id === updated.variant_id ? updated : r)),
      }
    }
    adjustOpen.value = false
    ui.success(t('admin.inventory.adjusted'))
    if (movementsLoaded) void movements.load()
  } catch (e) {
    if (!form.applyApiError(e)) adjustError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.inventory')" />
    <BaseTabs v-model="tab" :tabs="tabs" :label="$t('admin.nav.inventory')" />

    <!-- Stock levels -->
    <div v-show="tab === 'stock'">
      <div class="admin-toolbar">
        <div class="grow admin-search">
          <Search :size="16" aria-hidden="true" />
          <label for="inv-q" class="sr-only">{{ $t('admin.common.search') }}</label>
          <input
            id="inv-q"
            v-model="filters.q"
            type="search"
            :placeholder="$t('admin.inventory.searchPlaceholder')"
          />
        </div>
        <BaseCheckbox
          v-model="filters.low"
          toggle
          :label="$t('admin.inventory.lowOnly')"
          class="low-toggle"
        />
      </div>

      <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
      <template v-else>
        <DataTable
          :columns="columns"
          :rows="list.items.value"
          :row-key="(r) => r.variant_id"
          :loading="list.loading.value && !list.data.value"
          :row-class="(r) => ({ low: r.is_low })"
          :caption="$t('admin.inventory.stockTab')"
        >
          <template #cell-product="{ row }">
            <span class="admin-thumb-cell">
              <ImagePlaceholder
                class="thumb"
                :src="row.image_url"
                :name="row.product_name_en"
                small
              />
              <span class="txt">
                <RouterLink :to="{ name: 'admin-product-edit', params: { id: row.product_id } }">
                  {{ row.product_name_en }}
                </RouterLink>
                <span class="sub">{{ row.variant_name_en }}</span>
              </span>
            </span>
          </template>
          <template #cell-sku="{ row }"
            ><span class="admin-mono">{{ row.sku }}</span></template
          >
          <template #cell-stock="{ row }">
            <span class="stock-cell">
              <strong class="tabular">{{ row.stock }}</strong>
              <BaseBadge v-if="row.is_low" :tone="row.stock === 0 ? 'danger' : 'warning'" size="sm">
                {{ $t('admin.inventory.low') }}
              </BaseBadge>
            </span>
          </template>
          <template #cell-low_stock_threshold="{ row }">
            <span class="tabular muted">{{ row.low_stock_threshold }}</span>
          </template>
          <template #cell-actions="{ row }">
            <span class="admin-row-actions">
              <BaseButton size="sm" variant="secondary" @click="openAdjust(row)">
                <SlidersHorizontal :size="15" aria-hidden="true" />{{
                  $t('admin.inventory.adjust')
                }}
              </BaseButton>
              <BaseButton
                size="sm"
                variant="ghost"
                :aria-label="`${$t('admin.inventory.history')} ${row.sku}`"
                :title="$t('admin.inventory.history')"
                @click="showHistory(row)"
              >
                <History :size="16" aria-hidden="true" />
              </BaseButton>
            </span>
          </template>
          <template #empty>
            <EmptyState
              :title="hasFilters ? $t('admin.common.noResults') : $t('admin.inventory.empty')"
            >
              <BaseButton v-if="hasFilters" variant="secondary" @click="clear">
                {{ $t('admin.common.clearFilters') }}
              </BaseButton>
            </EmptyState>
          </template>
        </DataTable>
        <BasePagination
          :page="list.page.value"
          :pages="list.pages.value"
          @update:page="list.setPage"
        />
      </template>
    </div>

    <!-- Movements -->
    <div v-show="tab === 'movements'">
      <div v-if="movementVariant" class="filter-chip">
        <span>{{ $t('admin.inventory.movementsFor', { sku: movementVariant.sku }) }}</span>
        <BaseButton size="sm" variant="ghost" @click="showAllMovements">
          {{ $t('admin.inventory.showAllMovements') }}
        </BaseButton>
      </div>
      <ErrorState
        v-if="movements.error.value"
        :error="movements.error.value"
        @retry="movements.retry"
      />
      <template v-else>
        <DataTable
          :columns="movementColumns"
          :rows="movements.items.value"
          :row-key="(m) => m.id"
          :loading="movements.loading.value && !movements.data.value"
          :caption="$t('admin.inventory.movementsTab')"
        >
          <template #cell-product="{ row }">
            <span class="txt-col">
              <span>{{ row.product_name_en }}</span>
              <span class="sub admin-mono">{{ row.sku }}</span>
            </span>
          </template>
          <template #cell-change="{ row }">
            <strong class="tabular" :class="row.change > 0 ? 'text-success' : 'text-danger'">
              {{ row.change > 0 ? `+${row.change}` : row.change }}
            </strong>
          </template>
          <template #cell-stock_after="{ row }"
            ><span class="tabular">{{ row.stock_after }}</span></template
          >
          <template #cell-reason="{ row }">
            <span class="txt-col">
              <span>{{ $t(`admin.inventory.reasons.${row.reason}`) }}</span>
              <RouterLink
                v-if="row.order_code"
                :to="{ name: 'admin-order', params: { code: row.order_code } }"
                class="sub admin-mono"
              >
                {{ row.order_code }}
              </RouterLink>
              <span v-if="row.note" class="sub">{{ row.note }}</span>
            </span>
          </template>
          <template #cell-by="{ row }">
            <span class="text-sm">{{ row.user_email || '—' }}</span>
          </template>
          <template #cell-created_at="{ row }">
            <span class="nowrap text-sm">{{ dateTime(row.created_at) }}</span>
          </template>
          <template #empty>
            <EmptyState compact :title="$t('admin.inventory.noMovements')" />
          </template>
        </DataTable>
        <BasePagination
          :page="movements.page.value"
          :pages="movements.pages.value"
          @update:page="movements.setPage"
        />
      </template>
    </div>

    <BaseModal v-model:open="adjustOpen" size="sm" :title="$t('admin.inventory.adjustTitle')">
      <form
        v-if="adjustTarget"
        id="adjust-form"
        class="stack"
        novalidate
        @submit.prevent="submitAdjust"
      >
        <div class="target">
          <strong>{{ adjustTarget.product_name_en }}</strong>
          <span class="sub">{{ adjustTarget.variant_name_en }} · {{ adjustTarget.sku }}</span>
          <span class="sub">{{ $t('admin.inventory.current', { n: adjustTarget.stock }) }}</span>
        </div>
        <BaseSelect
          v-model="form.values.reason"
          :label="$t('admin.inventory.reason')"
          :options="reasonOptions"
        />
        <BaseInput
          v-model="form.values.change"
          type="number"
          inputmode="numeric"
          step="1"
          :label="$t('admin.inventory.change')"
          :hint="$t('admin.inventory.changeHint')"
          :error="form.errors.value.change"
          required
          data-autofocus
        />
        <p v-if="previewStock !== null" class="preview" :class="{ bad: previewStock < 0 }">
          {{
            previewStock < 0
              ? $t('admin.inventory.negative')
              : $t('admin.inventory.newStock', { n: previewStock })
          }}
        </p>
        <BaseInput
          v-model="form.values.note"
          multiline
          :rows="2"
          :label="$t('admin.inventory.note')"
          :error="form.errors.value.note"
          optional
        />
        <p v-if="adjustError" class="text-danger text-sm" role="alert">{{ adjustError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="adjustOpen = false">{{
          $t('common.cancel')
        }}</BaseButton>
        <BaseButton type="submit" form="adjust-form" :loading="form.submitting.value">{{
          $t('common.save')
        }}</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<style scoped>
.low-toggle {
  align-self: center;
}
.stock-cell {
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
}
:deep(tr.low) td {
  background: rgb(254 243 199 / 0.35);
}
.txt-col {
  display: inline-flex;
  flex-direction: column;
  min-width: 0;
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.filter-chip {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  flex-wrap: wrap;
  background: var(--c-primary-50);
  border: 1px solid var(--c-primary-200);
  border-radius: var(--radius-md);
  padding: var(--sp-2) var(--sp-3);
  margin-bottom: var(--sp-4);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
}
.target {
  display: flex;
  flex-direction: column;
  background: var(--c-surface-2);
  padding: var(--sp-3);
  border-radius: var(--radius-md);
}
.preview {
  margin: 0;
  font-weight: var(--fw-semibold);
  color: var(--c-primary);
}
.preview.bad {
  color: var(--c-danger);
}
</style>
