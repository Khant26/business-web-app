<script setup lang="ts" generic="R">
/**
 * Responsive table: a regular table on ≥768px, stacked cards on small screens
 * (each cell shows its column label). Cells render via `#cell-<key>` slots.
 */
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import type { Column } from './types'

withDefaults(
  defineProps<{
    columns: Column[]
    rows: R[]
    rowKey: (row: R) => string | number
    loading?: boolean
    skeletonRows?: number
    caption?: string
    rowClass?: (row: R) => string | Record<string, boolean> | undefined
  }>(),
  { skeletonRows: 6 },
)

function cellValue(row: R, key: string): unknown {
  return (row as Record<string, unknown>)[key]
}
</script>

<template>
  <div class="dt card">
    <table class="table">
      <caption v-if="caption" class="sr-only">
        {{
          caption
        }}
      </caption>
      <thead>
        <tr>
          <th
            v-for="c in columns"
            :key="c.key"
            scope="col"
            :class="{ num: c.align === 'right', center: c.align === 'center' }"
            :style="c.width ? { width: c.width } : undefined"
          >
            {{ c.label }}
          </th>
        </tr>
      </thead>
      <tbody v-if="loading">
        <tr v-for="i in skeletonRows" :key="i" aria-hidden="true">
          <td v-for="c in columns" :key="c.key" :data-label="c.label">
            <SkeletonBlock height="14px" :width="c.primary ? '70%' : '50%'" />
          </td>
        </tr>
      </tbody>
      <tbody v-else>
        <template v-for="row in rows" :key="rowKey(row)">
          <tr :class="rowClass?.(row)">
            <td
              v-for="c in columns"
              :key="c.key"
              :data-label="c.label"
              :class="{
                num: c.align === 'right',
                center: c.align === 'center',
                primary: c.primary,
              }"
            >
              <slot :name="`cell-${c.key}`" :row="row" :value="cellValue(row, c.key)">
                {{ cellValue(row, c.key) }}
              </slot>
            </td>
          </tr>
          <slot name="after-row" :row="row" :colspan="columns.length" />
        </template>
      </tbody>
    </table>
    <slot v-if="!loading && rows.length === 0" name="empty" />
  </div>
</template>

<style scoped>
.dt {
  overflow: hidden;
}
.table th.center,
.table td.center {
  text-align: center;
}
@media (min-width: 768px) {
  .dt {
    overflow-x: auto;
  }
}
@media (max-width: 767px) {
  .dt {
    background: transparent;
    border: 0;
    box-shadow: none;
    overflow: visible;
  }
  .table thead {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
  }
  .table,
  .table tbody,
  .table tr,
  .table td {
    display: block;
    width: 100%;
  }
  .table tbody {
    display: flex;
    flex-direction: column;
    gap: var(--sp-3);
  }
  .table tr {
    background: var(--c-surface);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-lg);
    padding: var(--sp-2) var(--sp-4);
    box-shadow: var(--shadow-xs);
  }
  .table tbody tr:hover {
    background: var(--c-surface);
  }
  .table td {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--sp-3);
    padding: var(--sp-2) 0;
    border-bottom: 1px dashed var(--c-border);
    text-align: right;
    min-height: 40px;
  }
  .table td:last-child {
    border-bottom: 0;
  }
  .table td::before {
    content: attr(data-label);
    font-size: var(--fs-xs);
    color: var(--c-muted);
    font-weight: var(--fw-semibold);
    text-align: left;
    flex: none;
    max-width: 45%;
  }
  .table td.primary {
    justify-content: flex-start;
    text-align: left;
    font-size: var(--fs-md);
  }
  .table td.primary::before {
    display: none;
  }
  .table td.num {
    text-align: right;
  }
  .table td:empty {
    display: none;
  }
}
</style>
