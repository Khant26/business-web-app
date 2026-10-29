<script setup lang="ts">
/** Image picker with preview, type/size validation (jpg/png/webp ≤ 5 MB). */
import { onBeforeUnmount, ref, useId } from 'vue'
import { ImagePlus } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
  defineProps<{
    label?: string
    hint?: string
    maxBytes?: number
    preview?: boolean
    disabled?: boolean
  }>(),
  { maxBytes: 5 * 1024 * 1024, preview: true },
)
const emit = defineEmits<{ select: [File] }>()
const { t } = useI18n()

const id = `fd-${useId()}`
const dragging = ref(false)
const error = ref('')
const previewUrl = ref<string | null>(null)
const ACCEPT = ['image/jpeg', 'image/png', 'image/webp']

function handle(file: File | undefined | null) {
  error.value = ''
  if (!file) return
  if (!ACCEPT.includes(file.type)) {
    error.value = t('validation.fileType')
    return
  }
  if (file.size > props.maxBytes) {
    error.value = t('validation.fileSize')
    return
  }
  if (props.preview) {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = URL.createObjectURL(file)
  }
  emit('select', file)
}

function onChange(e: Event) {
  const input = e.target as HTMLInputElement
  handle(input.files?.[0])
  input.value = ''
}
function onDrop(e: DragEvent) {
  dragging.value = false
  if (props.disabled) return
  handle(e.dataTransfer?.files?.[0])
}
function clearPreview() {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = null
}
defineExpose({ clearPreview })
onBeforeUnmount(clearPreview)
</script>

<template>
  <div class="fd-wrap">
    <label
      :for="id"
      class="drop"
      :class="{ dragging, disabled, 'has-preview': !!previewUrl }"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop.prevent="onDrop"
    >
      <img v-if="previewUrl" :src="previewUrl" alt="" class="pv" />
      <span v-else class="ph">
        <ImagePlus :size="28" aria-hidden="true" />
        <span class="lbl">{{ label ?? $t('common.upload') }}</span>
        <span v-if="hint" class="hint">{{ hint }}</span>
      </span>
      <input
        :id="id"
        type="file"
        accept="image/jpeg,image/png,image/webp"
        class="sr-only"
        :disabled="disabled"
        @change="onChange"
      />
    </label>
    <p v-if="error" class="err" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.drop {
  display: grid;
  place-items: center;
  min-height: 140px;
  border: 2px dashed var(--c-border-strong);
  border-radius: var(--radius-lg);
  background: var(--c-surface-2);
  cursor: pointer;
  text-align: center;
  padding: var(--sp-4);
  transition:
    border-color var(--dur-fast),
    background var(--dur-fast);
  overflow: hidden;
}
.drop:hover,
.drop.dragging {
  border-color: var(--c-primary-500);
  background: var(--c-primary-50);
}
.drop:focus-within {
  box-shadow: var(--ring);
}
.drop.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.has-preview {
  padding: var(--sp-2);
  background: var(--c-surface);
}
.ph {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--sp-1);
  color: var(--c-muted);
}
.lbl {
  font-weight: var(--fw-semibold);
  color: var(--c-primary);
}
.hint {
  font-size: var(--fs-xs);
}
.pv {
  max-height: 320px;
  width: auto;
  border-radius: var(--radius-md);
}
.err {
  margin: var(--sp-2) 0 0;
  color: var(--c-danger);
  font-size: var(--fs-xs);
}
</style>
