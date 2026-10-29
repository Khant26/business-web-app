<script setup lang="ts">
import { RouterLink, type RouteLocationRaw } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'

defineProps<{ title: string; subtitle?: string; back?: RouteLocationRaw }>()
</script>

<template>
  <div class="aph">
    <div class="titles">
      <RouterLink v-if="back" :to="back" class="back">
        <ArrowLeft :size="16" aria-hidden="true" />{{ $t('admin.common.back') }}
      </RouterLink>
      <h1 class="title">{{ title }}</h1>
      <p v-if="subtitle" class="sub">{{ subtitle }}</p>
      <slot name="meta" />
    </div>
    <div v-if="$slots.default" class="actions"><slot /></div>
  </div>
</template>

<style scoped>
.aph {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--sp-3);
  margin-bottom: var(--sp-5);
}
.titles {
  min-width: 0;
}
.back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: var(--fs-sm);
  color: var(--c-muted);
  margin-bottom: var(--sp-1);
}
.title {
  margin: 0;
  font-size: var(--fs-xl);
  overflow-wrap: anywhere;
}
@media (min-width: 768px) {
  .title {
    font-size: var(--fs-2xl);
  }
}
.sub {
  margin: var(--sp-1) 0 0;
  color: var(--c-muted);
  font-size: var(--fs-sm);
}
.actions {
  display: flex;
  gap: var(--sp-2);
  flex-wrap: wrap;
}
</style>
