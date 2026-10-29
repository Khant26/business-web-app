<script setup lang="ts">
import { ref, toRef, useId } from 'vue'
import { X } from 'lucide-vue-next'
import { useFocusTrap } from '@/composables/useFocusTrap'

const open = defineModel<boolean>('open', { default: false })
const props = withDefaults(
  defineProps<{ title?: string; size?: 'sm' | 'md' | 'lg' | 'xl'; persistent?: boolean }>(),
  { size: 'md' },
)
const emit = defineEmits<{ close: [] }>()

const panel = ref<HTMLElement | null>(null)
const titleId = `m-${useId()}`

function close() {
  open.value = false
  emit('close')
}
function onBackdrop() {
  if (!props.persistent) close()
}

useFocusTrap(panel, toRef(open), close)
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="open" class="overlay" @mousedown.self="onBackdrop">
        <div
          ref="panel"
          class="panel"
          :class="`size-${size}`"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="title ? titleId : undefined"
          tabindex="-1"
        >
          <header v-if="title || $slots.header" class="head">
            <slot name="header">
              <h2 :id="titleId" class="title">{{ title }}</h2>
            </slot>
            <button type="button" class="x" :aria-label="$t('common.close')" @click="close">
              <X :size="20" aria-hidden="true" />
            </button>
          </header>
          <div class="body"><slot /></div>
          <footer v-if="$slots.footer" class="foot"><slot name="footer" /></footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  background: rgb(15 17 30 / 0.5);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 0;
}
@media (min-width: 640px) {
  .overlay {
    align-items: center;
    padding: var(--sp-6);
  }
}
.panel {
  width: 100%;
  max-height: 92dvh;
  display: flex;
  flex-direction: column;
  background: var(--c-surface);
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  box-shadow: var(--shadow-lg);
  outline: none;
}
@media (min-width: 640px) {
  .panel {
    border-radius: var(--radius-xl);
  }
  .size-sm {
    max-width: 420px;
  }
  .size-md {
    max-width: 560px;
  }
  .size-lg {
    max-width: 760px;
  }
  .size-xl {
    max-width: 1000px;
  }
}
.head {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-4) var(--sp-5);
  border-bottom: 1px solid var(--c-border);
}
.title {
  margin: 0;
  font-size: var(--fs-lg);
  flex: 1;
}
.x {
  margin-left: auto;
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border: 0;
  background: transparent;
  border-radius: var(--radius-md);
  color: var(--c-muted);
  cursor: pointer;
}
.x:hover {
  background: var(--c-surface-2);
  color: var(--c-text);
}
.body {
  padding: var(--sp-5);
  overflow-y: auto;
}
.foot {
  display: flex;
  justify-content: flex-end;
  gap: var(--sp-2);
  flex-wrap: wrap;
  padding: var(--sp-4) var(--sp-5);
  border-top: 1px solid var(--c-border);
  padding-bottom: max(var(--sp-4), env(safe-area-inset-bottom));
}
.modal-enter-active,
.modal-leave-active {
  transition: opacity var(--dur) var(--ease);
}
.modal-enter-active .panel,
.modal-leave-active .panel {
  transition: transform var(--dur) var(--ease);
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .panel,
.modal-leave-to .panel {
  transform: translateY(24px);
}
</style>
