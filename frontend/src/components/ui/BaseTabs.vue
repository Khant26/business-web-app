<script setup lang="ts" generic="V extends string">
import type { TabItem } from '@/types/ui'

const model = defineModel<V>({ required: true })
defineProps<{ tabs: TabItem<V>[]; label?: string }>()

function onKey(e: KeyboardEvent, tabs: TabItem<V>[], idx: number) {
  let next = idx
  if (e.key === 'ArrowRight') next = (idx + 1) % tabs.length
  else if (e.key === 'ArrowLeft') next = (idx - 1 + tabs.length) % tabs.length
  else return
  e.preventDefault()
  const t = tabs[next]
  if (t) model.value = t.value
  const el = (e.currentTarget as HTMLElement).parentElement?.children[next] as
    HTMLElement | undefined
  el?.focus()
}
</script>

<template>
  <div class="tabs" role="tablist" :aria-label="label">
    <button
      v-for="(t, i) in tabs"
      :key="t.value"
      type="button"
      role="tab"
      class="tab"
      :class="{ active: model === t.value }"
      :aria-selected="model === t.value"
      :tabindex="model === t.value ? 0 : -1"
      @click="model = t.value"
      @keydown="onKey($event, tabs, i)"
    >
      {{ t.label }}
      <span v-if="t.count !== undefined" class="count">{{ t.count }}</span>
    </button>
  </div>
</template>

<style scoped>
.tabs {
  display: flex;
  gap: var(--sp-1);
  overflow-x: auto;
  scrollbar-width: none;
  border-bottom: 1px solid var(--c-border);
  margin-bottom: var(--sp-4);
}
.tabs::-webkit-scrollbar {
  display: none;
}
.tab {
  flex: none;
  border: 0;
  background: transparent;
  padding: var(--sp-3) var(--sp-3);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-muted);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.tab:hover {
  color: var(--c-text);
}
.tab.active {
  color: var(--c-primary);
  border-bottom-color: var(--c-primary);
  font-weight: var(--fw-semibold);
}
.tab:focus-visible {
  outline: none;
  box-shadow: inset var(--ring);
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
}
.count {
  background: var(--c-surface-3);
  border-radius: 999px;
  padding: 0 7px;
  font-size: 11px;
  color: var(--c-text-2);
}
.active .count {
  background: var(--c-primary-100);
  color: var(--c-primary);
}
</style>
