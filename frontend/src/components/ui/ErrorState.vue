<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, RotateCcw } from 'lucide-vue-next'
import BaseButton from './BaseButton.vue'
import { useErrorMessage } from '@/composables/useErrorMessage'

const props = defineProps<{ error: unknown; title?: string; compact?: boolean }>()
defineEmits<{ retry: [] }>()

const { errorText } = useErrorMessage()
const message = computed(() => errorText(props.error))
</script>

<template>
  <div class="err" :class="{ compact }" role="alert">
    <AlertTriangle class="icon" :size="compact ? 24 : 36" aria-hidden="true" />
    <h3 class="title">{{ title ?? $t('common.errorTitle') }}</h3>
    <p class="text">{{ message }}</p>
    <BaseButton variant="secondary" size="sm" @click="$emit('retry')">
      <RotateCcw :size="16" aria-hidden="true" />
      {{ $t('common.retry') }}
    </BaseButton>
  </div>
</template>

<style scoped>
.err {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--sp-2);
  padding: var(--sp-10) var(--sp-4);
}
.compact {
  padding: var(--sp-6) var(--sp-4);
}
.icon {
  color: var(--c-danger);
}
.title {
  margin: 0;
  font-size: var(--fs-lg);
}
.text {
  margin: 0 0 var(--sp-2);
  color: var(--c-muted);
  max-width: 48ch;
}
</style>
