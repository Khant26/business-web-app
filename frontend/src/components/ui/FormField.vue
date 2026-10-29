<script setup lang="ts">
/** Label + control slot + hint/error. The slot receives `id` and `describedby` for a11y wiring. */
import { computed, useId } from 'vue'

const props = defineProps<{
  label?: string
  error?: string
  hint?: string
  required?: boolean
  optional?: boolean
  id?: string
}>()

const autoId = useId()
const fieldId = computed(() => props.id ?? `f-${autoId}`)
const describedby = computed(
  () =>
    [props.error ? `${fieldId.value}-err` : '', props.hint ? `${fieldId.value}-hint` : '']
      .filter(Boolean)
      .join(' ') || undefined,
)
</script>

<template>
  <div class="field" :class="{ 'has-error': !!error }">
    <label v-if="label" :for="fieldId" class="field-label">
      {{ label }}
      <span v-if="required" class="req" aria-hidden="true">*</span>
      <span v-if="optional" class="opt">({{ $t('common.optional') }})</span>
    </label>
    <slot :id="fieldId" :describedby="describedby" :invalid="!!error" />
    <p v-if="hint && !error" :id="`${fieldId}-hint`" class="field-hint">{{ hint }}</p>
    <p v-if="error" :id="`${fieldId}-err`" class="field-error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: var(--sp-1);
  min-width: 0;
}
.field-label {
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
}
.req {
  color: var(--c-danger);
  margin-left: 2px;
}
.opt {
  color: var(--c-muted);
  font-weight: var(--fw-regular);
  margin-left: 4px;
  font-size: var(--fs-xs);
}
.field-hint {
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.field-error {
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--c-danger);
  font-weight: var(--fw-medium);
}
</style>
