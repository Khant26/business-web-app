<script setup lang="ts" generic="T extends string | number | null">
import { ChevronDown } from 'lucide-vue-next'
import FormField from './FormField.vue'
import type { SelectOption } from '@/types/ui'

defineOptions({ inheritAttrs: false })

const model = defineModel<T>()

defineProps<{
  options: SelectOption<T>[]
  label?: string
  error?: string
  hint?: string
  required?: boolean
  optional?: boolean
  placeholder?: string
  id?: string
}>()
</script>

<template>
  <FormField
    :id="id"
    v-slot="{ id: fid, describedby, invalid }"
    :label="label"
    :error="error"
    :hint="hint"
    :required="required"
    :optional="optional"
  >
    <div class="select" :class="{ invalid }">
      <select
        :id="fid"
        v-model="model"
        :aria-invalid="invalid || undefined"
        :aria-describedby="describedby"
        :required="required"
        v-bind="$attrs"
      >
        <option v-if="placeholder !== undefined" :value="null" disabled>{{ placeholder }}</option>
        <option v-for="o in options" :key="String(o.value)" :value="o.value" :disabled="o.disabled">
          {{ o.label }}
        </option>
      </select>
      <ChevronDown class="chev" :size="16" aria-hidden="true" />
    </div>
  </FormField>
</template>

<style scoped>
.select {
  position: relative;
}
select {
  width: 100%;
  min-height: 44px;
  padding: var(--sp-2) var(--sp-8) var(--sp-2) var(--sp-3);
  appearance: none;
  background: var(--c-surface);
  border: 1px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  font-size: 16px;
  cursor: pointer;
}
@media (min-width: 768px) {
  select {
    font-size: var(--fs-sm);
  }
}
select:focus {
  outline: none;
  border-color: var(--c-primary-600);
  box-shadow: var(--ring);
}
select:disabled {
  background: var(--c-surface-2);
  color: var(--c-muted);
  cursor: not-allowed;
}
.invalid select {
  border-color: var(--c-danger);
}
.chev {
  position: absolute;
  right: var(--sp-3);
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: var(--c-muted);
}
</style>
