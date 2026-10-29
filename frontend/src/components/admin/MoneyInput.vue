<script setup lang="ts">
/** Baht text input bound to an integer satang model (null when empty and `nullable`). */
import { ref, watch } from 'vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import { bahtToSatang, satangToBaht } from '@/utils/money'

const model = defineModel<number | null>({ default: null })
const props = defineProps<{
  label?: string
  error?: string
  hint?: string
  required?: boolean
  optional?: boolean
  nullable?: boolean
  id?: string
}>()

function fmt(v: number | null | undefined): string {
  if (v === null || v === undefined || Number.isNaN(v)) return ''
  const b = satangToBaht(v)
  return Number.isInteger(b) ? String(b) : b.toFixed(2)
}

const text = ref(fmt(model.value))

watch(model, (v) => {
  // Only rewrite the text when the model changed from outside.
  const current = text.value.trim() === '' ? null : bahtToSatang(text.value)
  if (current !== v && !(Number.isNaN(current) && Number.isNaN(v))) text.value = fmt(v)
})

watch(text, (s) => {
  const trimmed = String(s ?? '').trim()
  if (trimmed === '') {
    model.value = props.nullable ? null : NaN
    return
  }
  model.value = bahtToSatang(trimmed)
})
</script>

<template>
  <BaseInput
    :id="id"
    v-model="text"
    :label="label"
    :error="error"
    :hint="hint"
    :required="required"
    :optional="optional"
    prefix="฿"
    inputmode="decimal"
    autocomplete="off"
    placeholder="0.00"
  />
</template>
