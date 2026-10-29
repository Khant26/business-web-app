<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

const model = defineModel<string>({ default: '' })
const props = withDefaults(
  defineProps<{ length?: number; disabled?: boolean; invalid?: boolean }>(),
  {
    length: 6,
  },
)
const emit = defineEmits<{ complete: [string] }>()

const inputs = ref<HTMLInputElement[]>([])
const digits = ref<string[]>(Array.from({ length: props.length }, (_, i) => model.value[i] ?? ''))

watch(model, (v) => {
  if (v !== digits.value.join('')) {
    digits.value = Array.from({ length: props.length }, (_, i) => v[i] ?? '')
  }
})

function sync() {
  const code = digits.value.join('')
  model.value = code
  if (code.length === props.length && /^\d+$/.test(code)) emit('complete', code)
}

function focusAt(i: number) {
  void nextTick(() => {
    const el = inputs.value[Math.max(0, Math.min(props.length - 1, i))]
    el?.focus()
    el?.select()
  })
}

function fill(from: number, text: string) {
  const chars = text
    .replace(/\D/g, '')
    .slice(0, props.length - from)
    .split('')
  chars.forEach((c, k) => (digits.value[from + k] = c))
  sync()
  focusAt(from + chars.length)
}

function onInput(i: number, e: Event) {
  const el = e.target as HTMLInputElement
  const v = el.value.replace(/\D/g, '')
  if (v.length > 1) {
    fill(i, v)
    return
  }
  digits.value[i] = v
  el.value = v
  sync()
  if (v) focusAt(i + 1)
}

function onKeydown(i: number, e: KeyboardEvent) {
  if (e.key === 'Backspace' && !digits.value[i] && i > 0) {
    e.preventDefault()
    digits.value[i - 1] = ''
    sync()
    focusAt(i - 1)
  } else if (e.key === 'ArrowLeft') {
    e.preventDefault()
    focusAt(i - 1)
  } else if (e.key === 'ArrowRight') {
    e.preventDefault()
    focusAt(i + 1)
  }
}

function onPaste(i: number, e: ClipboardEvent) {
  const text = e.clipboardData?.getData('text') ?? ''
  if (!text) return
  e.preventDefault()
  fill(i, text)
}

defineExpose({
  focus: () =>
    focusAt(digits.value.findIndex((d) => !d) === -1 ? 0 : digits.value.findIndex((d) => !d)),
})
</script>

<template>
  <div class="otp" :class="{ invalid }" role="group" :aria-label="$t('auth.code')">
    <input
      v-for="(d, i) in digits"
      :key="i"
      ref="inputs"
      :value="d"
      type="text"
      inputmode="numeric"
      pattern="[0-9]*"
      maxlength="6"
      :autocomplete="i === 0 ? 'one-time-code' : 'off'"
      :aria-label="$t('auth.digit', { n: i + 1 })"
      :aria-invalid="invalid || undefined"
      :disabled="disabled"
      @input="onInput(i, $event)"
      @keydown="onKeydown(i, $event)"
      @paste="onPaste(i, $event)"
      @focus="($event.target as HTMLInputElement).select()"
    />
  </div>
</template>

<style scoped>
.otp {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: var(--sp-2);
  max-width: 360px;
}
input {
  width: 100%;
  aspect-ratio: 5 / 6;
  max-height: 60px;
  text-align: center;
  font-size: 1.5rem;
  font-weight: var(--fw-bold);
  font-variant-numeric: tabular-nums;
  border: 1.5px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  background: var(--c-surface);
  caret-color: var(--c-primary);
  transition:
    border-color var(--dur-fast),
    box-shadow var(--dur-fast);
}
input:focus {
  outline: none;
  border-color: var(--c-primary-600);
  box-shadow: var(--ring);
}
.invalid input {
  border-color: var(--c-danger);
}
input:disabled {
  background: var(--c-surface-2);
}
</style>
