<script setup lang="ts">
import type { Component } from 'vue'
import { PackageOpen } from 'lucide-vue-next'

defineProps<{ title: string; text?: string; icon?: Component; compact?: boolean }>()
</script>

<template>
  <div class="empty" :class="{ compact }">
    <div class="icon" aria-hidden="true">
      <component :is="icon ?? PackageOpen" :size="compact ? 28 : 40" :stroke-width="1.5" />
    </div>
    <h3 class="title">{{ title }}</h3>
    <p v-if="text" class="text">{{ text }}</p>
    <div v-if="$slots.default" class="actions"><slot /></div>
  </div>
</template>

<style scoped>
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: var(--sp-12) var(--sp-4);
  gap: var(--sp-2);
}
.compact {
  padding: var(--sp-8) var(--sp-4);
}
.icon {
  display: grid;
  place-items: center;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--c-primary-50);
  color: var(--c-primary);
  margin-bottom: var(--sp-2);
}
.compact .icon {
  width: 56px;
  height: 56px;
}
.title {
  margin: 0;
  font-size: var(--fs-lg);
}
.text {
  margin: 0;
  color: var(--c-muted);
  max-width: 42ch;
}
.actions {
  margin-top: var(--sp-3);
  display: flex;
  gap: var(--sp-2);
  flex-wrap: wrap;
  justify-content: center;
}
</style>
