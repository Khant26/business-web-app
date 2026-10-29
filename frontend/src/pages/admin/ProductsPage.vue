<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Plus, Search, Star } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { adminApi } from '@/api/admin'
import { useAdminList } from '@/components/admin/useAdminList'
import { useAsync } from '@/composables/useAsync'
import { useLocalized } from '@/composables/useLocalized'
import { formatBaht } from '@/utils/money'
import type { ProductStatus } from '@/types/api'
import type { Column } from '@/components/admin/types'
import type { SelectOption } from '@/types/ui'

const { t } = useI18n()
const { date } = useLocalized()

const categories = useAsync(() => adminApi.categories())
onMounted(() => void categories.run())

const { filters, list, hasFilters, clear } = useAdminList(
  { q: '', status: '' as ProductStatus | '', category: '' },
  (page, page_size, f) =>
    adminApi.products({
      page,
      page_size,
      q: f.q.trim() || undefined,
      status: f.status || undefined,
      category: f.category ? Number(f.category) : undefined,
    }),
  {
    debounce: ['q'],
    sanitize: (f) => {
      if (!['draft', 'active', 'archived'].includes(f.status)) f.status = ''
      if (f.category && !/^\d+$/.test(f.category)) f.category = ''
    },
  },
)

const statusOptions = computed<SelectOption<ProductStatus | ''>[]>(() => [
  { value: '', label: t('admin.products.allStatuses') },
  { value: 'active', label: t('admin.products.status.active') },
  { value: 'draft', label: t('admin.products.status.draft') },
  { value: 'archived', label: t('admin.products.status.archived') },
])
const categoryOptions = computed<SelectOption<string>[]>(() => [
  { value: '', label: t('admin.products.allCategories') },
  ...(categories.data.value ?? []).map((c) => ({ value: String(c.id), label: c.name_en })),
])

const columns = computed<Column[]>(() => [
  { key: 'name_en', label: t('admin.common.name'), primary: true },
  { key: 'category_name_en', label: t('admin.products.category') },
  { key: 'status', label: t('admin.products.statusLabel') },
  { key: 'min_price_satang', label: t('admin.products.price'), align: 'right' },
  { key: 'total_stock', label: t('admin.products.stock'), align: 'right' },
  { key: 'updated_at', label: t('admin.common.updated_at') },
])

const statusTone = (s: ProductStatus) =>
  s === 'active' ? 'success' : s === 'draft' ? 'warning' : 'neutral'
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.products')">
      <BaseButton :to="{ name: 'admin-product-new' }">
        <Plus :size="18" aria-hidden="true" />{{ $t('admin.products.new') }}
      </BaseButton>
    </AdminPageHeader>

    <div class="admin-toolbar">
      <div class="grow admin-search">
        <Search :size="16" aria-hidden="true" />
        <label for="products-q" class="sr-only">{{ $t('admin.common.search') }}</label>
        <input
          id="products-q"
          v-model="filters.q"
          type="search"
          :placeholder="$t('admin.products.searchPlaceholder')"
        />
      </div>
      <BaseSelect
        v-model="filters.status"
        class="fixed"
        :label="$t('admin.products.statusLabel')"
        :options="statusOptions"
      />
      <BaseSelect
        v-model="filters.category"
        class="fixed"
        :label="$t('admin.products.category')"
        :options="categoryOptions"
      />
      <BaseButton v-if="hasFilters" variant="ghost" @click="clear">{{
        $t('admin.common.clearFilters')
      }}</BaseButton>
    </div>

    <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
    <template v-else>
      <DataTable
        :columns="columns"
        :rows="list.items.value"
        :row-key="(r) => r.id"
        :loading="list.loading.value && !list.data.value"
        :caption="$t('admin.nav.products')"
      >
        <template #cell-name_en="{ row }">
          <span class="admin-thumb-cell">
            <ImagePlaceholder class="thumb" :src="row.image_url" :name="row.name_en" small />
            <span class="txt">
              <RouterLink :to="{ name: 'admin-product-edit', params: { id: row.id } }" class="nm">
                {{ row.name_en }}
              </RouterLink>
              <span class="sub">
                <Star
                  v-if="row.featured"
                  :size="12"
                  class="star"
                  :aria-label="$t('admin.products.featured')"
                />
                {{ $t('admin.products.variantsN', { n: row.variant_count }, row.variant_count) }}
                · <span class="admin-mono">{{ row.slug }}</span>
              </span>
            </span>
          </span>
        </template>
        <template #cell-category_name_en="{ row }">{{ row.category_name_en || '—' }}</template>
        <template #cell-status="{ row }">
          <BaseBadge :tone="statusTone(row.status)" size="sm">{{
            $t(`admin.products.status.${row.status}`)
          }}</BaseBadge>
        </template>
        <template #cell-min_price_satang="{ row }">
          <span class="tabular">{{ formatBaht(row.min_price_satang) }}</span>
        </template>
        <template #cell-total_stock="{ row }">
          <span class="tabular" :class="{ 'text-danger': row.total_stock === 0 }">{{
            row.total_stock
          }}</span>
        </template>
        <template #cell-updated_at="{ row }"
          ><span class="nowrap">{{ date(row.updated_at) }}</span></template
        >
        <template #empty>
          <EmptyState
            :title="hasFilters ? $t('admin.common.noResults') : $t('admin.products.empty')"
          >
            <BaseButton v-if="hasFilters" variant="secondary" @click="clear">
              {{ $t('admin.common.clearFilters') }}
            </BaseButton>
            <BaseButton v-else :to="{ name: 'admin-product-new' }">{{
              $t('admin.products.new')
            }}</BaseButton>
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
.nm {
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
  overflow: hidden;
  text-overflow: ellipsis;
}
.star {
  color: var(--c-star);
  fill: currentColor;
  vertical-align: -1px;
}
</style>
