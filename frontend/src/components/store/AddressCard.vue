<script setup lang="ts">
import { MapPin, Phone } from 'lucide-vue-next'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import type { ShippingAddress } from '@/types/api'

defineProps<{
  address: ShippingAddress & { label?: string; is_default?: boolean }
  plain?: boolean
}>()
</script>

<template>
  <div class="ac" :class="{ plain }">
    <div class="top">
      <strong class="name">{{ address.recipient_name }}</strong>
      <span v-if="address.label" class="label">{{ address.label }}</span>
      <BaseBadge v-if="address.is_default" tone="primary" size="sm">{{
        $t('address.default')
      }}</BaseBadge>
    </div>
    <p class="line">
      <MapPin :size="14" aria-hidden="true" />
      <span>
        {{ address.line1 }}, {{ address.subdistrict }}, {{ address.district }},
        {{ address.province }} {{ address.postcode }}
      </span>
    </p>
    <p class="line">
      <Phone :size="14" aria-hidden="true" /><span class="tabular">{{ address.phone }}</span>
    </p>
    <slot />
  </div>
</template>

<style scoped>
.ac {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  font-size: var(--fs-sm);
}
.top {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  flex-wrap: wrap;
}
.name {
  font-size: var(--fs-md);
}
.label {
  color: var(--c-muted);
  font-size: var(--fs-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: var(--fw-semibold);
}
.line {
  margin: 0;
  display: flex;
  gap: var(--sp-2);
  color: var(--c-text-2);
  align-items: flex-start;
}
.line :deep(svg) {
  flex: none;
  margin-top: 4px;
  color: var(--c-muted);
}
</style>
