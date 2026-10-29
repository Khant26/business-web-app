<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps<{ page: number; pages: number }>()
const emit = defineEmits<{ 'update:page': [number] }>()

type Item = number | 'gap-l' | 'gap-r'

const items = computed<Item[]>(() => {
  const { page, pages } = props
  if (pages <= 7) return Array.from({ length: pages }, (_, i) => i + 1)
  const out: Item[] = [1]
  const start = Math.max(2, page - 1)
  const end = Math.min(pages - 1, page + 1)
  if (start > 2) out.push('gap-l')
  for (let i = start; i <= end; i++) out.push(i)
  if (end < pages - 1) out.push('gap-r')
  out.push(pages)
  return out
})

function go(p: number) {
  if (p < 1 || p > props.pages || p === props.page) return
  emit('update:page', p)
}
</script>

<template>
  <nav v-if="pages > 1" class="pager" :aria-label="$t('common.pageOf', { page, pages })">
    <button
      type="button"
      class="pg nav"
      :disabled="page <= 1"
      :aria-label="$t('common.previous')"
      @click="go(page - 1)"
    >
      <ChevronLeft :size="18" aria-hidden="true" />
    </button>
    <template v-for="it in items" :key="it">
      <span v-if="typeof it !== 'number'" class="gap" aria-hidden="true">…</span>
      <button
        v-else
        type="button"
        class="pg"
        :class="{ active: it === page }"
        :aria-current="it === page ? 'page' : undefined"
        @click="go(it)"
      >
        {{ it }}
      </button>
    </template>
    <button
      type="button"
      class="pg nav"
      :disabled="page >= pages"
      :aria-label="$t('common.next')"
      @click="go(page + 1)"
    >
      <ChevronRight :size="18" aria-hidden="true" />
    </button>
  </nav>
</template>

<style scoped>
.pager {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 4px;
  margin-top: var(--sp-8);
  flex-wrap: wrap;
}
.pg {
  min-width: 40px;
  height: 40px;
  padding: 0 var(--sp-2);
  border-radius: var(--radius-md);
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  font-weight: var(--fw-medium);
  font-size: var(--fs-sm);
  cursor: pointer;
  display: grid;
  place-items: center;
  font-variant-numeric: tabular-nums;
}
.pg:hover:not(:disabled) {
  border-color: var(--c-primary-300);
  color: var(--c-primary);
}
.pg.active {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: #fff;
}
.pg:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.gap {
  padding: 0 4px;
  color: var(--c-muted);
}
</style>
