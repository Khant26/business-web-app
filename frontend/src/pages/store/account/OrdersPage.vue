<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronRight, Package } from 'lucide-vue-next'
import BaseTabs from '@/components/ui/BaseTabs.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { usePaged } from '@/composables/usePaged'
import { useLocalized } from '@/composables/useLocalized'
import { ordersApi } from '@/api/orders'
import { formatBaht } from '@/utils/money'
import { ORDER_STATUSES, type OrderStatus } from '@/types/api'
import type { TabItem } from '@/types/ui'

type Filter = 'all' | OrderStatus

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { date } = useLocalized()

const isStatus = (v: unknown): v is OrderStatus =>
  typeof v === 'string' && (ORDER_STATUSES as readonly string[]).includes(v)

const filter = ref<Filter>(isStatus(route.query.status) ? route.query.status : 'all')

const tabs = computed<TabItem<Filter>[]>(() => [
  { value: 'all', label: t('account.orders.all') },
  ...ORDER_STATUSES.map((s) => ({ value: s as Filter, label: t(`status.${s}`) })),
])

const paged = usePaged(
  (page, page_size) =>
    ordersApi.list({
      status: filter.value === 'all' ? undefined : filter.value,
      page,
      page_size,
    }),
  { pageSize: 10, syncRoute: true },
)

watch(filter, (f) => {
  const q = { ...route.query }
  if (f === 'all') delete q.status
  else q.status = f
  delete q.page
  void router.replace({ query: q })
  paged.reload()
})
</script>

<template>
  <div>
    <h1 class="h">{{ $t('account.orders.title') }}</h1>
    <BaseTabs v-model="filter" :tabs="tabs" :label="$t('account.orders.statusFilter')" />

    <ul v-if="paged.loading.value && !paged.items.value.length" class="list" aria-busy="true">
      <li v-for="i in 4" :key="i" class="card sk">
        <SkeletonBlock width="64px" height="64px" radius="var(--radius-md)" />
        <div class="grow"><SkeletonBlock :lines="3" /></div>
      </li>
    </ul>
    <ErrorState v-else-if="paged.error.value" :error="paged.error.value" @retry="paged.retry" />
    <EmptyState
      v-else-if="paged.items.value.length === 0"
      :icon="Package"
      :title="filter === 'all' ? $t('account.orders.empty') : $t('account.orders.emptyStatus')"
      :text="filter === 'all' ? $t('account.orders.emptyText') : undefined"
    >
      <BaseButton v-if="filter === 'all'" :to="{ name: 'products' }">
        {{ $t('cart.continue') }}
      </BaseButton>
    </EmptyState>
    <template v-else>
      <ul class="list" :class="{ dim: paged.loading.value }">
        <li v-for="o in paged.items.value" :key="o.code">
          <RouterLink
            :to="{ name: 'account-order', params: { code: o.code } }"
            class="card order"
            :aria-label="$t('account.orders.viewOrder', { code: o.code })"
          >
            <div class="thumb">
              <ImagePlaceholder :src="o.first_item_image" :name="o.code" small />
            </div>
            <div class="info">
              <div class="top">
                <strong class="code">{{ o.code }}</strong>
                <StatusBadge :status="o.status" />
              </div>
              <span class="muted text-sm">{{
                $t('account.orders.placedOn', { date: date(o.created_at) })
              }}</span>
              <span class="muted text-sm">
                {{ $t('account.orders.itemCount', { n: o.item_count }, o.item_count) }} ·
                {{ $t(`payment.${o.payment_method}`) }}
              </span>
            </div>
            <div class="right">
              <strong class="tabular">{{ formatBaht(o.total_satang) }}</strong>
              <ChevronRight :size="18" class="chev" aria-hidden="true" />
            </div>
          </RouterLink>
        </li>
      </ul>
      <BasePagination
        :page="paged.page.value"
        :pages="paged.pages.value"
        @update:page="paged.setPage"
      />
    </template>
  </div>
</template>

<style scoped>
.h {
  font-size: var(--fs-xl);
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.sk {
  display: flex;
  gap: var(--sp-3);
  padding: var(--sp-4);
}
.grow {
  flex: 1;
}
.order {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-3);
  color: var(--c-text);
  text-decoration: none;
  transition:
    border-color var(--dur-fast),
    box-shadow var(--dur-fast);
}
@media (min-width: 640px) {
  .order {
    padding: var(--sp-4);
    gap: var(--sp-4);
  }
}
.order:hover {
  text-decoration: none;
  border-color: var(--c-primary-300);
  box-shadow: var(--shadow-md);
}
.thumb {
  flex: none;
  width: 60px;
}
.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.top {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  flex-wrap: wrap;
}
.code {
  font-variant-numeric: tabular-nums;
}
.right {
  display: flex;
  align-items: center;
  gap: var(--sp-1);
  white-space: nowrap;
}
.chev {
  color: var(--c-muted);
}
.dim {
  opacity: 0.6;
}
</style>
