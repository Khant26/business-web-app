<script setup lang="ts">
import { ref, toRef, useId } from 'vue'
import { X } from 'lucide-vue-next'
import { useFocusTrap } from '@/composables/useFocusTrap'

const open = defineModel<boolean>('open', { default: false })
withDefaults(defineProps<{ title?: string; side?: 'left' | 'right' | 'bottom' }>(), {
  side: 'left',
})

const panel = ref<HTMLElement | null>(null)
const titleId = `d-${useId()}`
const close = () => (open.value = false)

useFocusTrap(panel, toRef(open), close)
</script>

<template>
  <Teleport to="body">
    <Transition :name="`drawer-${side}`">
      <div v-if="open" class="overlay" @mousedown.self="close">
        <aside
          ref="panel"
          class="panel"
          :class="`side-${side}`"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="title ? titleId : undefined"
          tabindex="-1"
        >
          <header class="head">
            <h2 v-if="title" :id="titleId" class="title">{{ title }}</h2>
            <slot name="header" />
            <button type="button" class="x" :aria-label="$t('common.close')" @click="close">
              <X :size="20" aria-hidden="true" />
            </button>
          </header>
          <div class="body"><slot /></div>
          <footer v-if="$slots.footer" class="foot"><slot name="footer" /></footer>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  z-index: var(--z-drawer);
  background: rgb(15 17 30 / 0.45);
}
.panel {
  position: absolute;
  background: var(--c-surface);
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
  outline: none;
}
.side-left,
.side-right {
  top: 0;
  bottom: 0;
  width: min(86vw, 340px);
}
.side-left {
  left: 0;
}
.side-right {
  right: 0;
}
.side-bottom {
  left: 0;
  right: 0;
  bottom: 0;
  max-height: 88dvh;
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
}
.head {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-4) var(--sp-4) var(--sp-3);
  border-bottom: 1px solid var(--c-border);
}
.title {
  margin: 0;
  font-size: var(--fs-lg);
}
.x {
  margin-left: auto;
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border: 0;
  background: transparent;
  border-radius: var(--radius-md);
  color: var(--c-muted);
  cursor: pointer;
}
.x:hover {
  background: var(--c-surface-2);
}
.body {
  flex: 1;
  overflow-y: auto;
  padding: var(--sp-4);
}
.foot {
  padding: var(--sp-3) var(--sp-4);
  padding-bottom: max(var(--sp-3), env(safe-area-inset-bottom));
  border-top: 1px solid var(--c-border);
  display: flex;
  gap: var(--sp-2);
}
[class*='drawer-'][class*='-enter-active'],
[class*='drawer-'][class*='-leave-active'] {
  transition: opacity var(--dur) var(--ease);
}
[class*='drawer-'][class*='-enter-active'] .panel,
[class*='drawer-'][class*='-leave-active'] .panel {
  transition: transform var(--dur-slow) var(--ease);
}
[class*='drawer-'][class*='-enter-from'],
[class*='drawer-'][class*='-leave-to'] {
  opacity: 0;
}
.drawer-left-enter-from .panel,
.drawer-left-leave-to .panel {
  transform: translateX(-100%);
}
.drawer-right-enter-from .panel,
.drawer-right-leave-to .panel {
  transform: translateX(100%);
}
.drawer-bottom-enter-from .panel,
.drawer-bottom-leave-to .panel {
  transform: translateY(100%);
}
</style>
