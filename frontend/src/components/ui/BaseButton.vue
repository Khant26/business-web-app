<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'accent' | 'link'
    size?: 'sm' | 'md' | 'lg'
    type?: 'button' | 'submit' | 'reset'
    to?: RouteLocationRaw
    href?: string
    loading?: boolean
    disabled?: boolean
    block?: boolean
    iconOnly?: boolean
  }>(),
  { variant: 'primary', size: 'md', type: 'button' },
)

const tag = computed(() => (props.to ? RouterLink : props.href ? 'a' : 'button'))
const isDisabled = computed(() => props.disabled || props.loading)
</script>

<template>
  <component
    :is="tag"
    :to="to"
    :href="href"
    :type="to || href ? undefined : type"
    :disabled="!to && !href ? isDisabled : undefined"
    :aria-disabled="isDisabled || undefined"
    :aria-busy="loading || undefined"
    class="btn"
    :class="[
      `btn-${variant}`,
      `btn-${size}`,
      { 'btn-block': block, 'btn-icon': iconOnly, 'is-loading': loading },
    ]"
  >
    <span v-if="loading" class="spinner" aria-hidden="true" />
    <slot />
  </component>
</template>

<style scoped>
.btn {
  --btn-bg: var(--c-primary);
  --btn-fg: var(--c-on-primary);
  --btn-border: transparent;
  --btn-hover: var(--c-primary-hover);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--sp-2);
  border: 1px solid var(--btn-border);
  background: var(--btn-bg);
  color: var(--btn-fg);
  font-weight: var(--fw-semibold);
  border-radius: var(--radius-md);
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
  line-height: 1.2;
  transition:
    background var(--dur-fast) var(--ease),
    border-color var(--dur-fast) var(--ease),
    box-shadow var(--dur-fast) var(--ease),
    transform var(--dur-fast) var(--ease);
  user-select: none;
}
.btn:hover {
  background: var(--btn-hover);
  text-decoration: none;
}
.btn:active:not(:disabled) {
  transform: translateY(1px);
}
.btn:focus-visible {
  outline: none;
  box-shadow: var(--ring);
}
.btn:disabled,
.btn[aria-disabled='true'] {
  opacity: 0.55;
  cursor: not-allowed;
  pointer-events: none;
}
.btn-sm {
  min-height: 34px;
  padding: 0 var(--sp-3);
  font-size: var(--fs-sm);
}
.btn-md {
  min-height: 42px;
  padding: 0 var(--sp-4);
  font-size: var(--fs-sm);
}
.btn-lg {
  min-height: 50px;
  padding: 0 var(--sp-6);
  font-size: var(--fs-md);
}
.btn-icon {
  padding: 0;
  aspect-ratio: 1;
}
.btn-block {
  width: 100%;
}
.btn-secondary {
  --btn-bg: var(--c-surface);
  --btn-fg: var(--c-text);
  --btn-border: var(--c-border-strong);
  --btn-hover: var(--c-surface-2);
}
.btn-ghost {
  --btn-bg: transparent;
  --btn-fg: var(--c-text-2);
  --btn-hover: var(--c-surface-2);
}
.btn-danger {
  --btn-bg: var(--c-danger);
  --btn-hover: #991b1b;
}
.btn-accent {
  --btn-bg: var(--c-accent-strong);
  --btn-hover: #9a3412;
}
.btn-link {
  --btn-bg: transparent;
  --btn-fg: var(--c-primary);
  --btn-hover: transparent;
  padding-inline: 0;
  min-height: auto;
}
.btn-link:hover {
  text-decoration: underline;
}
.spinner {
  width: 1em;
  height: 1em;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
