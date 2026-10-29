<script setup lang="ts">
import BaseModal from './BaseModal.vue'
import BaseButton from './BaseButton.vue'

const open = defineModel<boolean>('open', { default: false })
withDefaults(
  defineProps<{
    title: string
    message?: string
    confirmLabel?: string
    danger?: boolean
    loading?: boolean
  }>(),
  {},
)
const emit = defineEmits<{ confirm: [] }>()
</script>

<template>
  <BaseModal v-model:open="open" :title="title" size="sm">
    <p v-if="message" class="msg">{{ message }}</p>
    <slot />
    <template #footer>
      <BaseButton variant="secondary" @click="open = false">{{ $t('common.cancel') }}</BaseButton>
      <BaseButton
        :variant="danger ? 'danger' : 'primary'"
        :loading="loading"
        @click="emit('confirm')"
      >
        {{ confirmLabel ?? $t('common.confirm') }}
      </BaseButton>
    </template>
  </BaseModal>
</template>

<style scoped>
.msg {
  margin: 0;
  color: var(--c-text-2);
}
</style>
