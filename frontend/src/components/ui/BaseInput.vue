<script setup lang="ts">
import FormField from './FormField.vue'

defineOptions({ inheritAttrs: false })

const model = defineModel<string | number | null>({ default: '' })

defineProps<{
  label?: string
  error?: string
  hint?: string
  required?: boolean
  optional?: boolean
  id?: string
  multiline?: boolean
  rows?: number
  prefix?: string
}>()

defineEmits<{ blur: [FocusEvent] }>()
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
    <div class="control" :class="{ invalid, 'has-prefix': !!prefix }">
      <span v-if="prefix" class="prefix" aria-hidden="true">{{ prefix }}</span>
      <textarea
        v-if="multiline"
        :id="fid"
        v-model="model"
        :rows="rows ?? 4"
        :aria-invalid="invalid || undefined"
        :aria-describedby="describedby"
        :required="required"
        v-bind="$attrs"
        @blur="$emit('blur', $event)"
      />
      <input
        v-else
        :id="fid"
        v-model="model"
        :aria-invalid="invalid || undefined"
        :aria-describedby="describedby"
        :required="required"
        v-bind="$attrs"
        @blur="$emit('blur', $event)"
      />
    </div>
  </FormField>
</template>

<style scoped>
.control {
  position: relative;
  display: flex;
  align-items: stretch;
}
.prefix {
  display: flex;
  align-items: center;
  padding: 0 var(--sp-3);
  background: var(--c-surface-2);
  border: 1px solid var(--c-border-strong);
  border-right: 0;
  border-radius: var(--radius-md) 0 0 var(--radius-md);
  color: var(--c-muted);
  font-size: var(--fs-sm);
}
input,
textarea {
  width: 100%;
  min-height: 44px;
  padding: var(--sp-2) var(--sp-3);
  background: var(--c-surface);
  border: 1px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  font-size: 16px; /* prevents iOS zoom */
  transition:
    border-color var(--dur-fast) var(--ease),
    box-shadow var(--dur-fast) var(--ease);
}
@media (min-width: 768px) {
  input,
  textarea {
    font-size: var(--fs-sm);
  }
}
.has-prefix input {
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
}
textarea {
  resize: vertical;
  min-height: 88px;
}
input:hover,
textarea:hover {
  border-color: var(--c-muted);
}
input:focus,
textarea:focus {
  outline: none;
  border-color: var(--c-primary-600);
  box-shadow: var(--ring);
}
input:disabled,
input[readonly],
textarea:disabled {
  background: var(--c-surface-2);
  color: var(--c-muted);
}
.invalid input,
.invalid textarea {
  border-color: var(--c-danger);
}
.invalid input:focus,
.invalid textarea:focus {
  box-shadow: 0 0 0 3px rgb(185 28 28 / 0.2);
}
</style>
