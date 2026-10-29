<script setup lang="ts">
import { computed } from 'vue'
import { useNow } from '@/composables/useNow'
import { msUntil, splitDuration } from '@/utils/date'

const props = withDefaults(defineProps<{ endsAt: string; tone?: 'light' | 'dark' }>(), {
  tone: 'dark',
})
const now = useNow(1000)
const left = computed(() => splitDuration(msUntil(props.endsAt, now.value)))
const pad = (n: number) => String(n).padStart(2, '0')
</script>

<template>
  <div class="cd" :class="`t-${tone}`" role="timer" :aria-label="$t('catalog.countdownLabel')">
    <template v-if="left.total > 0">
      <span v-if="left.days > 0" class="u"
        ><b>{{ left.days }}</b
        >{{ $t('catalog.days') }}</span
      >
      <span class="u"
        ><b>{{ pad(left.hours) }}</b
        >{{ $t('catalog.hours') }}</span
      >
      <span class="u"
        ><b>{{ pad(left.minutes) }}</b
        >{{ $t('catalog.minutes') }}</span
      >
      <span class="u"
        ><b>{{ pad(left.seconds) }}</b
        >{{ $t('catalog.seconds') }}</span
      >
    </template>
    <span v-else class="ended">{{ $t('catalog.eventEnded') }}</span>
  </div>
</template>

<style scoped>
.cd {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  font-variant-numeric: tabular-nums;
}
.u {
  display: inline-flex;
  align-items: baseline;
  gap: 2px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: var(--fs-xs);
  font-weight: var(--fw-medium);
}
.u b {
  font-size: var(--fs-md);
  font-weight: var(--fw-bold);
}
.t-dark .u {
  background: var(--c-text);
  color: #fff;
}
.t-light .u {
  background: rgb(255 255 255 / 0.18);
  color: #fff;
  backdrop-filter: blur(4px);
}
.ended {
  font-size: var(--fs-sm);
  font-weight: var(--fw-semibold);
}
</style>
