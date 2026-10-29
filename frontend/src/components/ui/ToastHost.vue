<script setup lang="ts">
import { CheckCircle2, AlertCircle, Info, AlertTriangle, X } from 'lucide-vue-next'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const ICONS = { success: CheckCircle2, error: AlertCircle, info: Info, warning: AlertTriangle }
</script>

<template>
  <div class="toasts" aria-live="polite" aria-relevant="additions">
    <TransitionGroup name="toast">
      <div
        v-for="t in ui.toasts"
        :key="t.id"
        class="toast"
        :class="`k-${t.kind}`"
        :role="t.kind === 'error' ? 'alert' : 'status'"
      >
        <component :is="ICONS[t.kind]" :size="20" class="ic" aria-hidden="true" />
        <span class="msg">{{ t.message }}</span>
        <button type="button" class="x" :aria-label="$t('common.close')" @click="ui.dismiss(t.id)">
          <X :size="16" aria-hidden="true" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toasts {
  position: fixed;
  z-index: var(--z-toast);
  left: var(--sp-3);
  right: var(--sp-3);
  bottom: calc(var(--tabbar-h) + var(--sp-3) + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
  pointer-events: none;
  align-items: center;
}
@media (min-width: 900px) {
  .toasts {
    left: auto;
    right: var(--sp-6);
    bottom: var(--sp-6);
    align-items: flex-end;
  }
}
.toast {
  pointer-events: auto;
  display: flex;
  align-items: flex-start;
  gap: var(--sp-3);
  width: 100%;
  max-width: 420px;
  padding: var(--sp-3) var(--sp-3) var(--sp-3) var(--sp-4);
  background: var(--c-text);
  color: #fff;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  font-size: var(--fs-sm);
}
.ic {
  flex: none;
  margin-top: 1px;
}
.k-success .ic {
  color: #4ade80;
}
.k-error .ic {
  color: #f87171;
}
.k-info .ic {
  color: #93c5fd;
}
.k-warning .ic {
  color: #fbbf24;
}
.msg {
  flex: 1;
  padding-top: 1px;
}
.x {
  flex: none;
  border: 0;
  background: transparent;
  color: rgb(255 255 255 / 0.7);
  cursor: pointer;
  width: 28px;
  height: 28px;
  display: grid;
  place-items: center;
  border-radius: 6px;
}
.x:hover {
  background: rgb(255 255 255 / 0.12);
  color: #fff;
}
.toast-enter-active,
.toast-leave-active {
  transition: all var(--dur) var(--ease);
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(12px);
}
</style>
