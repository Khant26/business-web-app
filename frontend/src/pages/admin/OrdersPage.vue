<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Receipt, Search } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import BaseTabs from '@/components/ui/BaseTabs.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { adminApi } from '@/api/admin'
import { useAdminList } from '@/components/admin/useAdminList'
import { useLocalized } from '@/composables/useLocalized'
import { formatBaht } from '@/utils/money'
import { ORDER_STATUSES, type OrderStatus, type PaymentMethod } from '@/types/api'
import type { Column } from '@/components/admin/types'
import type { SelectOption, TabItem } from '@/types/ui'

type StatusTab = OrderStatus | 'all'

const { t } = useI18n()
const { dateTime } = useLocalized()

const {
  filters,
  list,
  hasFilters,
  clear: clearFilters,
} = useAdminList(
  {
    status: 'all' as StatusTab,
    q: '',
    payment_method: '' as PaymentMethod | '',
    from: '',
    to: '',
  },
  (page, page_size, f) =>
    adminApi.orders({
      page,
      page_size,
      status: f.status === 'all' ? undefined : f.status,
      q: f.q.trim() || undefined,
      payment_method: f.payment_method || undefined,
      from: f.from || undefined,
      to: f.to || undefined,
    }),
  {
    debounce: ['q'],
    sanitize: (f) => {
      if (f.status !== 'all' && !ORDER_STATUSES.includes(f.status)) f.status = 'all'
      if (f.payment_method !== 'promptpay' && f.payment_method !== 'cod') f.payment_method = ''
    },
  },
)

const tabs = computed<TabItem<StatusTab>[]>(() => [
  { value: 'all', label: t('common.all') },
  ...ORDER_STATUSES.map((s) => ({ value: s, label: t(`status.${s}`) })),
])
const methodOptions = computed<SelectOption<PaymentMethod | ''>[]>(() => [
  { value: '', label: t('admin.orders.allMethods') },
  { value: 'promptpay', label: t('payment.promptpay') },
  { value: 'cod', label: t('payment.cod') },
])

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.orders.code'), primary: true },
  { key: 'customer', label: t('admin.orders.customer') },
  { key: 'created_at', label: t('admin.orders.placed') },
  { key: 'payment_method', label: t('admin.orders.paymentMethod') },
  { key: 'item_count', label: t('admin.orders.items'), align: 'right' },
  { key: 'total_satang', label: t('admin.orders.total'), align: 'right' },
  { key: 'status', label: t('admin.common.status') },
])
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.orders')" />

    <BaseTabs v-model="filters.status" :tabs="tabs" :label="$t('admin.common.status')" />

    <div class="admin-toolbar">
      <div class="grow admin-search">
        <Search :size="16" aria-hidden="true" />
        <label for="orders-q" class="sr-only">{{ $t('admin.common.search') }}</label>
        <input
          id="orders-q"
          v-model="filters.q"
          type="search"
          :placeholder="$t('admin.orders.searchPlaceholder')"
        />
      </div>
      <BaseSelect
        v-model="filters.payment_method"
        class="fixed"
        :label="$t('admin.orders.paymentMethod')"
        :options="methodOptions"
      />
      <BaseInput
        v-model="filters.from"
        class="fixed"
        type="date"
        :label="$t('admin.orders.from')"
      />
      <BaseInput v-model="filters.to" class="fixed" type="date" :label="$t('admin.orders.to')" />
      <BaseButton v-if="hasFilters" variant="ghost" @click="clearFilters">
        {{ $t('admin.common.clearFilters') }}
      </BaseButton>
    </div>

    <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
    <template v-else>
      <DataTable
        :columns="columns"
        :rows="list.items.value"
        :row-key="(r) => r.code"
        :loading="list.loading.value && !list.data.value"
        :caption="$t('admin.nav.orders')"
      >
        <template #cell-code="{ row }">
          <span class="code-cell">
            <RouterLink
              :to="{ name: 'admin-order', params: { code: row.code } }"
              class="admin-mono code"
            >
              {{ row.code }}
            </RouterLink>
            <span v-if="row.has_pending_slip" class="slip" :title="$t('admin.orders.pendingSlip')">
              <Receipt :size="14" aria-hidden="true" />{{ $t('admin.orders.pendingSlip') }}
            </span>
          </span>
        </template>
        <template #cell-customer="{ row }">
          <span class="cust">
            <span>{{ row.recipient_name }}</span>
            <span class="muted text-xs">{{ row.customer_email }}</span>
          </span>
        </template>
        <template #cell-created_at="{ row }">
          <span class="nowrap">{{ dateTime(row.created_at) }}</span>
        </template>
        <template #cell-payment_method="{ row }">{{
          $t(`payment.${row.payment_method}`)
        }}</template>
        <template #cell-total_satang="{ row }">
          <strong class="tabular">{{ formatBaht(row.total_satang) }}</strong>
        </template>
        <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
        <template #empty>
          <EmptyState
            :title="$t('admin.orders.empty')"
            :text="hasFilters ? $t('admin.common.noResults') : undefined"
          >
            <BaseButton v-if="hasFilters" variant="secondary" @click="clearFilters">
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
</template>

<style scoped>
.code-cell {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}
.code {
  font-weight: var(--fw-semibold);
}
.slip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: var(--fw-semibold);
  color: var(--c-info);
  background: var(--c-info-bg);
  padding: 1px 8px;
  border-radius: var(--radius-full);
}
.cust {
  display: inline-flex;
  flex-direction: column;
  min-width: 0;
}
</style>
