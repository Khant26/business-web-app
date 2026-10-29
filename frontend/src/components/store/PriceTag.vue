<script setup lang="ts">
import { computed } from 'vue'
import { formatBaht, percentOff } from '@/utils/money'

const props = withDefaults(
  defineProps<{
    price: number
    compareAt?: number | null
    size?: 'sm' | 'md' | 'lg'
    showBadge?: boolean
  }>(),
  { size: 'md', compareAt: null },
)
const off = computed(() => percentOff(props.price, props.compareAt))
</script>

<template>
  <span class="price" :class="`s-${size}`">
    <span class="now" :class="{ sale: off > 0 }">{{ formatBaht(price) }}</span>
    <template v-if="off > 0">
      <s class="was"
        ><span class="sr-only">{{ $t('product.sale') }}: </span>{{ formatBaht(compareAt) }}</s
      >
      <span v-if="showBadge" class="off">{{ $t('product.off', { n: off }) }}</span>
    </template>
  </span>
</template>

<style scoped>
.price {
  display: inline-flex;
  align-items: baseline;
  flex-wrap: wrap;
  column-gap: var(--sp-2);
  row-gap: 2px;
  font-variant-numeric: tabular-nums;
}
.now {
  font-weight: var(--fw-bold);
  color: var(--c-text);
}
.now.sale {
  color: var(--c-accent-strong);
}
.was {
  color: var(--c-muted);
  font-size: 0.85em;
}
.off {
  align-self: center;
  font-size: 11px;
  font-weight: var(--fw-bold);
  color: #fff;
  background: var(--c-accent-strong);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
}
.s-sm .now {
  font-size: var(--fs-sm);
}
.s-md .now {
  font-size: var(--fs-md);
}
.s-lg .now {
  font-size: var(--fs-2xl);
}
.s-lg .was {
  font-size: var(--fs-md);
}
.s-lg .off {
  font-size: var(--fs-sm);
  padding: 2px 8px;
}
</style>
