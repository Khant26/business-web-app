<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronDown, ChevronRight } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import { useAdminList } from '@/components/admin/useAdminList'
import BasePagination from '@/components/ui/BasePagination.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { adminApi } from '@/api/admin'
import { useLocalized } from '@/composables/useLocalized'
import type { AuditLog } from '@/types/api'
import type { Column } from '@/components/admin/types'

const { t } = useI18n()
const { dateTime } = useLocalized()

const { list } = useAdminList({}, (page, page_size) => adminApi.auditLogs({ page, page_size }), {
  pageSize: 50,
})

const expanded = reactive(new Set<number>())
function toggle(id: number) {
  if (expanded.has(id)) expanded.delete(id)
  else expanded.add(id)
}

function hasDetails(l: AuditLog) {
  const d = l.details
  if (d === null || d === undefined || d === '') return false
  if (typeof d === 'object') return Object.keys(d as object).length > 0
  return true
}

function pretty(d: unknown): string {
  if (typeof d === 'string') {
    try {
      return JSON.stringify(JSON.parse(d), null, 2)
    } catch {
      return d
    }
  }
  return JSON.stringify(d, null, 2)
}

const columns = computed<Column[]>(() => [
  { key: 'created_at', label: t('admin.audit.time') },
  { key: 'user_email', label: t('admin.audit.user'), primary: true },
  { key: 'action', label: t('admin.audit.action') },
  { key: 'entity', label: t('admin.audit.entity') },
  { key: 'ip', label: t('admin.audit.ip') },
  { key: 'details', label: t('admin.audit.details'), align: 'right' },
])
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.audit')" />

    <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
    <template v-else>
      <DataTable
        :columns="columns"
        :rows="list.items.value"
        :row-key="(l) => l.id"
        :loading="list.loading.value && !list.data.value"
        :caption="$t('admin.nav.audit')"
      >
        <template #cell-created_at="{ row }"
          ><span class="nowrap text-sm">{{ dateTime(row.created_at) }}</span></template
        >
        <template #cell-user_email="{ row }">
          <span class="text-sm">{{ row.user_email || $t('admin.audit.system') }}</span>
        </template>
        <template #cell-action="{ row }"
          ><code class="admin-mono act">{{ row.action }}</code></template
        >
        <template #cell-entity="{ row }">
          <span class="text-sm"
            >{{ row.entity
            }}<span v-if="row.entity_id" class="muted"> #{{ row.entity_id }}</span></span
          >
        </template>
        <template #cell-ip="{ row }"
          ><span class="admin-mono muted text-xs">{{ row.ip || '—' }}</span></template
        >
        <template #cell-details="{ row }">
          <button
            v-if="hasDetails(row)"
            type="button"
            class="toggle"
            :aria-expanded="expanded.has(row.id)"
            :aria-controls="`audit-${row.id}`"
            @click="toggle(row.id)"
          >
            <ChevronDown v-if="expanded.has(row.id)" :size="16" aria-hidden="true" />
            <ChevronRight v-else :size="16" aria-hidden="true" />
            {{
              expanded.has(row.id) ? $t('admin.audit.hideDetails') : $t('admin.audit.showDetails')
            }}
          </button>
          <span v-else class="muted">—</span>
        </template>
        <template #after-row="{ row, colspan }">
          <tr v-if="expanded.has(row.id)" :id="`audit-${row.id}`" class="details-row">
            <td :colspan="colspan">
              <pre class="json">{{ pretty(row.details) }}</pre>
            </td>
          </tr>
        </template>
        <template #empty>
          <EmptyState :title="$t('admin.audit.empty')" />
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
.act {
  background: var(--c-surface-2);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  white-space: nowrap;
}
.toggle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 0;
  background: none;
  color: var(--c-primary);
  font-size: var(--fs-xs);
  font-weight: var(--fw-semibold);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  white-space: nowrap;
}
.toggle:hover {
  background: var(--c-primary-50);
}
.details-row td {
  background: var(--c-surface-2);
}
.details-row td::before {
  display: none;
}
.json {
  margin: 0;
  max-height: 360px;
  overflow: auto;
  font-size: var(--fs-xs);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  text-align: left;
  width: 100%;
}
</style>
