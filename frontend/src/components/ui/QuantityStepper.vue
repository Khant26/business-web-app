<script setup lang="ts">
import { Minus, Plus } from 'lucide-vue-next'
import { clamp } from '@/utils/misc'

const model = defineModel<number>({ required: true })
const props = withDefaults(
  defineProps<{
    min?: number
    max?: number
    size?: 'sm' | 'md'
    disabled?: boolean
    label?: string
  }>(),
  { min: 1, max: 99, size: 'md' },
)

function set(v: number) {
  if (Number.isNaN(v)) return
  model.value = clamp(Math.floor(v), props.min, Math.max(props.min, props.max))
}

function onInput(e: Event) {
  const v = Number((e.target as HTMLInputElement).value)
  if (!Number.isNaN(v) && v > 0) set(v)
}
function onBlur(e: Event) {
  const el = e.target as HTMLInputElement
  const v = Number(el.value)
  set(Number.isNaN(v) || v < props.min ? props.min : v)
  el.value = String(model.value)
}
</script>

<template>
  <div
    class="stepper"
    :class="`s-${size}`"
    role="group"
    :aria-label="label ?? $t('product.quantity')"
  >
    <button
      type="button"
      class="b"
      :disabled="disabled || model <= min"
      :aria-label="$t('product.decrease')"
      @click="set(model - 1)"
    >
      <Minus :size="16" aria-hidden="true" />
    </button>
    <input
      type="number"
      inputmode="numeric"
      :min="min"
      :max="max"
      :value="model"
      :disabled="disabled"
      :aria-label="label ?? $t('product.quantity')"
      @input="onInput"
      @blur="onBlur"
    />
    <button
      type="button"
      class="b"
      :disabled="disabled || model >= max"
      :aria-label="$t('product.increase')"
      @click="set(model + 1)"
    >
      <Plus :size="16" aria-hidden="true" />
    </button>
  </div>
</template>

<style scoped>
.stepper {
  display: inline-flex;
  align-items: stretch;
  border: 1px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  background: var(--c-surface);
  overflow: hidden;
}
.b {
  border: 0;
  background: transparent;
  width: 40px;
  display: grid;
  place-items: center;
  cursor: pointer;
  color: var(--c-text-2);
}
.s-sm .b {
  width: 32px;
}
.b:hover:not(:disabled) {
  background: var(--c-surface-2);
  color: var(--c-primary);
}
.b:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.b:focus-visible {
  outline: none;
  box-shadow: inset var(--ring);
}
input {
  width: 44px;
  border: 0;
  border-inline: 1px solid var(--c-border);
  text-align: center;
  font-weight: var(--fw-semibold);
  font-variant-numeric: tabular-nums;
  height: 42px;
  font-size: 16px;
  -moz-appearance: textfield;
  appearance: textfield;
}
.s-sm input {
  height: 34px;
  width: 38px;
  font-size: var(--fs-sm);
}
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input:focus {
  outline: none;
  background: var(--c-primary-50);
}
</style>
