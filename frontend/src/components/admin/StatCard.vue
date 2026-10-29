<script setup lang="ts">
import type { Component } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'

withDefaults(
  defineProps<{
    label: string
    value: string | number
    icon: Component
    tone?: 'primary' | 'accent' | 'success' | 'warning' | 'danger' | 'info'
    to?: RouteLocationRaw
    loading?: boolean
  }>(),
  { tone: 'primary' },
)
</script>

<template>
  <component :is="to ? RouterLink : 'div'" :to="to" class="stat" :class="`t-${tone}`">
    <span class="ic" aria-hidden="true"><component :is="icon" :size="20" /></span>
    <span class="txt">
      <span class="label">{{ label }}</span>
      <span v-if="loading" class="value sk" aria-hidden="true" />
      <span v-else class="value">{{ value }}</span>
    </span>
  </component>
</template>

<style scoped>
.stat {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-4);
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  color: var(--c-text);
  min-width: 0;
  transition:
    border-color var(--dur-fast),
    box-shadow var(--dur-fast);
}
a.stat:hover {
  text-decoration: none;
  border-color: var(--c-primary-200);
  box-shadow: var(--shadow-md);
}
.ic {
  flex: none;
  width: 42px;
  height: 42px;
  border-radius: var(--radius-md);
  display: grid;
  place-items: center;
}
.t-primary .ic {
  background: var(--c-primary-100);
  color: var(--c-primary);
}
.t-accent .ic {
  background: var(--c-accent-100);
  color: var(--c-accent-strong);
}
.t-success .ic {
  background: var(--c-success-bg);
  color: var(--c-success);
}
.t-warning .ic {
  background: var(--c-warning-bg);
  color: var(--c-warning);
}
.t-danger .ic {
  background: var(--c-danger-bg);
  color: var(--c-danger);
}
.t-info .ic {
  background: var(--c-info-bg);
  color: var(--c-info);
}
.txt {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.label {
  font-size: var(--fs-xs);
  color: var(--c-muted);
  font-weight: var(--fw-medium);
}
.value {
  font-size: var(--fs-xl);
  font-weight: var(--fw-bold);
  font-variant-numeric: tabular-nums;
  line-height: 1.3;
  overflow-wrap: anywhere;
}
.sk {
  display: block;
  width: 80px;
  height: 24px;
  margin-top: 4px;
  border-radius: var(--radius-sm);
  background: var(--c-surface-3);
}
</style>
