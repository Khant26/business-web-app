<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { AlertTriangle, Banknote, Package, Receipt, ShoppingBag } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import StatCard from '@/components/admin/StatCard.vue'
import SalesChart from '@/components/admin/SalesChart.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { adminApi } from '@/api/admin'
import { useAsync } from '@/composables/useAsync'
import { useLocalized } from '@/composables/useLocalized'
import { useAuthStore } from '@/stores/auth'
import { formatBaht } from '@/utils/money'

const auth = useAuthStore()
const { dateTime } = useLocalized()
const dash = useAsync(() => adminApi.dashboard())
onMounted(() => void dash.run())
</script>

<template>
  <div>
    <AdminPageHeader
      :title="$t('admin.nav.dashboard')"
      :subtitle="
        $t('admin.dashboard.greeting', { name: auth.user?.name || auth.user?.email || '' })
      "
    />

    <ErrorState v-if="dash.error.value" :error="dash.error.value" @retry="dash.retry" />
    <template v-else>
      <section class="stats" :aria-busy="dash.loading.value">
        <StatCard
          :label="$t('admin.dashboard.todaySales')"
          :value="formatBaht(dash.data.value?.today_sales_satang ?? 0)"
          :icon="Banknote"
          tone="success"
          :loading="dash.loading.value"
        />
        <StatCard
          :label="$t('admin.dashboard.todayOrders')"
          :value="dash.data.value?.today_orders ?? 0"
          :icon="ShoppingBag"
          :loading="dash.loading.value"
          :to="{ name: 'admin-orders' }"
        />
        <StatCard
          :label="$t('admin.dashboard.pendingSlips')"
          :value="dash.data.value?.pending_slips ?? 0"
          :icon="Receipt"
          tone="info"
          :loading="dash.loading.value"
          :to="{ name: 'admin-orders', query: { status: 'payment_review' } }"
        />
        <StatCard
          :label="$t('admin.dashboard.toPack')"
          :value="dash.data.value?.to_pack ?? 0"
          :icon="Package"
          tone="accent"
          :loading="dash.loading.value"
          :to="{ name: 'admin-orders', query: { status: 'paid' } }"
        />
        <StatCard
          :label="$t('admin.dashboard.lowStock')"
          :value="dash.data.value?.low_stock_count ?? 0"
          :icon="AlertTriangle"
          :tone="(dash.data.value?.low_stock_count ?? 0) > 0 ? 'danger' : 'success'"
          :loading="dash.loading.value"
          :to="{ name: 'admin-inventory', query: { low: '1' } }"
        />
      </section>

      <section class="card card-pad chart-card">
        <h2 class="admin-section-title">{{ $t('admin.dashboard.sales7d') }}</h2>
        <SkeletonBlock v-if="dash.loading.value" height="220px" radius="var(--radius-md)" />
        <SalesChart v-else :data="dash.data.value?.sales_7d ?? []" />
      </section>

      <div class="cols">
        <section class="card card-pad">
          <h2 class="admin-section-title">
            {{ $t('admin.dashboard.recentOrders') }}
            <RouterLink :to="{ name: 'admin-orders' }" class="more">{{
              $t('admin.dashboard.viewAll')
            }}</RouterLink>
          </h2>
          <SkeletonBlock v-if="dash.loading.value" :lines="5" height="36px" />
          <EmptyState
            v-else-if="!dash.data.value?.recent_orders.length"
            compact
            :title="$t('admin.dashboard.noRecent')"
          />
          <ul v-else class="list">
            <li v-for="o in dash.data.value.recent_orders" :key="o.code">
              <RouterLink :to="{ name: 'admin-order', params: { code: o.code } }" class="row">
                <span class="main">
                  <span class="code admin-mono">{{ o.code }}</span>
                  <span class="sub"
                    >{{ o.recipient_name || o.customer_email }} · {{ dateTime(o.created_at) }}</span
                  >
                </span>
                <span class="right">
                  <span class="amt tabular">{{ formatBaht(o.total_satang) }}</span>
                  <StatusBadge :status="o.status" />
                </span>
              </RouterLink>
            </li>
          </ul>
        </section>

        <section class="card card-pad">
          <h2 class="admin-section-title">
            {{ $t('admin.dashboard.lowStockList') }}
            <RouterLink :to="{ name: 'admin-inventory', query: { low: '1' } }" class="more">
              {{ $t('admin.dashboard.viewAll') }}
            </RouterLink>
          </h2>
          <SkeletonBlock v-if="dash.loading.value" :lines="5" height="36px" />
          <EmptyState
            v-else-if="!dash.data.value?.low_stock.length"
            compact
            :title="$t('admin.dashboard.noLowStock')"
          />
          <ul v-else class="list">
            <li v-for="r in dash.data.value.low_stock" :key="r.variant_id">
              <RouterLink
                :to="{ name: 'admin-product-edit', params: { id: r.product_id } }"
                class="row"
              >
                <span class="admin-thumb-cell">
                  <ImagePlaceholder
                    class="thumb"
                    :src="r.image_url"
                    :name="r.product_name_en"
                    small
                  />
                  <span class="txt">
                    <span class="nm">{{ r.product_name_en }}</span>
                    <span class="sub">{{ r.variant_name_en }} · {{ r.sku }}</span>
                  </span>
                </span>
                <span class="stock" :class="{ zero: r.stock === 0 }">
                  {{ $t('admin.dashboard.left', { n: r.stock }) }}
                </span>
              </RouterLink>
            </li>
          </ul>
        </section>
      </div>
    </template>
  </div>
</template>

<style scoped>
.stats {
  display: grid;
  gap: var(--sp-3);
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-bottom: var(--sp-5);
}
.stats > :first-child {
  grid-column: 1 / -1;
}
@media (min-width: 768px) {
  .stats {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .stats > :first-child {
    grid-column: auto;
  }
}
@media (min-width: 1280px) {
  .stats {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}
.chart-card {
  margin-bottom: var(--sp-5);
}
.cols {
  display: grid;
  gap: var(--sp-5);
}
@media (min-width: 1024px) {
  .cols {
    grid-template-columns: 1fr 1fr;
  }
}
.more {
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.list li + li {
  border-top: 1px solid var(--c-border);
}
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--sp-3);
  padding: var(--sp-3) var(--sp-1);
  color: var(--c-text);
  border-radius: var(--radius-sm);
}
.row:hover {
  background: var(--c-surface-2);
  text-decoration: none;
}
.main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.code {
  font-weight: var(--fw-semibold);
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nm {
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  flex: none;
}
.amt {
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
}
.stock {
  flex: none;
  font-size: var(--fs-xs);
  font-weight: var(--fw-bold);
  color: var(--c-warning);
  background: var(--c-warning-bg);
  padding: 2px 8px;
  border-radius: var(--radius-full);
}
.stock.zero {
  color: var(--c-danger);
  background: var(--c-danger-bg);
}
</style>
