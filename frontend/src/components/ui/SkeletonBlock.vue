<script setup lang="ts">
withDefaults(
  defineProps<{
    width?: string
    height?: string
    radius?: string
    lines?: number
    circle?: boolean
  }>(),
  { width: '100%', height: '1em', radius: 'var(--radius-sm)', lines: 1 },
)
</script>

<template>
  <span class="sk-wrap" aria-hidden="true">
    <span
      v-for="i in lines"
      :key="i"
      class="sk"
      :style="{
        width: lines > 1 && i === lines ? '60%' : width,
        height,
        borderRadius: circle ? '50%' : radius,
      }"
    />
  </span>
</template>

<style scoped>
.sk-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.sk {
  display: block;
  background: linear-gradient(
    90deg,
    var(--c-surface-3) 0%,
    var(--c-surface-2) 50%,
    var(--c-surface-3) 100%
  );
  background-size: 200% 100%;
  animation: shimmer 1.3s ease-in-out infinite;
}
@keyframes shimmer {
  from {
    background-position: 200% 0;
  }
  to {
    background-position: -200% 0;
  }
}
</style>
