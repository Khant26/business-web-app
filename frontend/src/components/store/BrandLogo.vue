<script setup lang="ts">
defineProps<{ light?: boolean }>()
</script>

<template>
  <span class="logo" :class="{ light }">
    <svg class="mark" viewBox="0 0 64 64" width="34" height="34" aria-hidden="true">
      <rect width="64" height="64" rx="16" fill="currentColor" />
      <path d="M20 24h24l-2.5 22a4 4 0 0 1-4 3.5h-11a4 4 0 0 1-4-3.5z" fill="#fff" />
      <path
        d="M26 24v-3a6 6 0 0 1 12 0v3"
        stroke="#f59e0b"
        stroke-width="3.5"
        fill="none"
        stroke-linecap="round"
      />
    </svg>
    <span class="word">Sawyun<span class="dot">.</span></span>
  </span>
</template>

<style scoped>
.logo {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--c-primary);
}
.word {
  font-family: 'Inter Variable', system-ui, sans-serif;
  font-weight: 800;
  font-size: 1.3rem;
  letter-spacing: -0.03em;
  color: var(--c-text);
  line-height: 1;
}
.dot {
  color: var(--c-accent);
}
.light {
  color: #fff;
}
.light .mark rect {
  fill: var(--c-primary);
}
.light .word {
  color: #fff;
}
</style>
