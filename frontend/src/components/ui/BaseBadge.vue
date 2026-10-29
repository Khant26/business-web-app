<script setup lang="ts">
withDefaults(
  defineProps<{
    tone?: 'neutral' | 'primary' | 'accent' | 'success' | 'warning' | 'danger' | 'info' | 'sale'
    size?: 'sm' | 'md'
    dot?: boolean
  }>(),
  { tone: 'neutral', size: 'md' },
)
</script>

<template>
  <span class="badge" :class="[`tone-${tone}`, `size-${size}`]">
    <span v-if="dot" class="dot" aria-hidden="true" />
    <slot />
  </span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: var(--radius-full);
  font-weight: var(--fw-semibold);
  white-space: nowrap;
  line-height: 1.3;
}
.size-sm {
  font-size: 11px;
  padding: 2px 8px;
}
.size-md {
  font-size: var(--fs-xs);
  padding: 3px 10px;
}
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}
.tone-neutral {
  background: var(--c-surface-3);
  color: var(--c-text-2);
}
.tone-primary {
  background: var(--c-primary-100);
  color: var(--c-primary-hover);
}
.tone-accent {
  background: var(--c-accent-100);
  color: var(--c-accent-strong);
}
.tone-sale {
  background: var(--c-accent-strong);
  color: #fff;
}
.tone-success {
  background: var(--c-success-bg);
  color: var(--c-success);
}
.tone-warning {
  background: var(--c-warning-bg);
  color: var(--c-warning);
}
.tone-danger {
  background: var(--c-danger-bg);
  color: var(--c-danger);
}
.tone-info {
  background: var(--c-info-bg);
  color: var(--c-info);
}
</style>
