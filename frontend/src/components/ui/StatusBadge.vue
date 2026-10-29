<script setup lang="ts">
import { computed } from 'vue'
import BaseBadge from './BaseBadge.vue'
import type { OrderStatus, SlipStatus } from '@/types/api'

const props = defineProps<{ status: OrderStatus | SlipStatus; kind?: 'order' | 'slip' }>()

type Tone = 'neutral' | 'primary' | 'accent' | 'success' | 'warning' | 'danger' | 'info'

const TONES: Record<string, Tone> = {
  pending_payment: 'warning',
  payment_review: 'info',
  paid: 'primary',
  confirmed: 'primary',
  packing: 'accent',
  shipped: 'info',
  delivered: 'success',
  cancelled: 'neutral',
  expired: 'neutral',
  pending: 'warning',
  approved: 'success',
  rejected: 'danger',
}

const tone = computed<Tone>(() => TONES[props.status] ?? 'neutral')
const key = computed(() =>
  props.kind === 'slip' ? `slipStatus.${props.status}` : `status.${props.status}`,
)
</script>

<template>
  <BaseBadge :tone="tone" dot>{{ $t(key) }}</BaseBadge>
</template>
