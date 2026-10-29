<script setup lang="ts">
import { useId } from 'vue'

defineOptions({ inheritAttrs: false })
const model = defineModel<boolean>({ default: false })
defineProps<{ label?: string; hint?: string; toggle?: boolean }>()
const id = `cb-${useId()}`
</script>

<template>
  <label class="cb" :class="{ 'is-switch': toggle }" :for="id">
    <input
      :id="id"
      v-model="model"
      type="checkbox"
      :role="toggle ? 'switch' : undefined"
      v-bind="$attrs"
    />
    <span class="box" aria-hidden="true" />
    <span class="text">
      <span v-if="label || $slots.default" class="label"
        ><slot>{{ label }}</slot></span
      >
      <span v-if="hint" class="hint">{{ hint }}</span>
    </span>
  </label>
</template>

<style scoped>
.cb {
  display: inline-flex;
  align-items: flex-start;
  gap: var(--sp-2);
  cursor: pointer;
  position: relative;
  font-size: var(--fs-sm);
}
input {
  position: absolute;
  opacity: 0;
  width: 1px;
  height: 1px;
}
.box {
  flex: none;
  width: 20px;
  height: 20px;
  margin-top: 1px;
  border: 1.5px solid var(--c-border-strong);
  border-radius: 6px;
  background: var(--c-surface);
  display: grid;
  place-items: center;
  transition: all var(--dur-fast) var(--ease);
}
input:checked + .box {
  background: var(--c-primary);
  border-color: var(--c-primary);
}
input:checked + .box::after {
  content: '';
  width: 10px;
  height: 6px;
  border: 2px solid #fff;
  border-top: 0;
  border-right: 0;
  transform: rotate(-45deg) translate(1px, -1px);
}
input:focus-visible + .box {
  box-shadow: var(--ring);
}
input:disabled + .box {
  opacity: 0.5;
}
.is-switch .box {
  width: 38px;
  height: 22px;
  border-radius: 999px;
  background: var(--c-surface-3);
  border-color: transparent;
  position: relative;
}
.is-switch .box::before {
  content: '';
  position: absolute;
  left: 3px;
  top: 3px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  box-shadow: var(--shadow-sm);
  transition: transform var(--dur) var(--ease);
}
.is-switch input:checked + .box {
  background: var(--c-primary);
}
.is-switch input:checked + .box::after {
  display: none;
}
.is-switch input:checked + .box::before {
  transform: translateX(16px);
}
.text {
  display: flex;
  flex-direction: column;
}
.label {
  color: var(--c-text);
}
.hint {
  color: var(--c-muted);
  font-size: var(--fs-xs);
}
</style>
