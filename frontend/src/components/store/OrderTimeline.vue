<script setup lang="ts">
import { computed } from 'vue'
import { Check, X } from 'lucide-vue-next'
import { useLocalized } from '@/composables/useLocalized'
import type { OrderHistoryEntry, OrderStatus, PaymentMethod } from '@/types/api'

const props = defineProps<{
  status: OrderStatus
  paymentMethod: PaymentMethod
  history: OrderHistoryEntry[]
}>()

const { dateTime } = useLocalized()

const FLOWS: Record<PaymentMethod, OrderStatus[]> = {
  promptpay: ['pending_payment', 'payment_review', 'paid', 'packing', 'shipped', 'delivered'],
  cod: ['confirmed', 'packing', 'shipped', 'delivered'],
}

interface Step {
  status: OrderStatus
  state: 'done' | 'current' | 'upcoming' | 'stopped'
  at: string | null
  note: string
}

const terminal = computed(() => props.status === 'cancelled' || props.status === 'expired')

const steps = computed<Step[]>(() => {
  const flow = FLOWS[props.paymentMethod]
  const latest = (s: OrderStatus) => {
    for (let i = props.history.length - 1; i >= 0; i--) {
      const h = props.history[i]
      if (h && h.status === s) return h
    }
    return null
  }

  if (terminal.value) {
    // Show what actually happened, ending with the cancel/expire step.
    const seen = new Set<OrderStatus>()
    const out: Step[] = []
    for (const h of props.history) {
      if (seen.has(h.status)) continue
      seen.add(h.status)
      out.push({
        status: h.status,
        state: h.status === props.status ? 'stopped' : 'done',
        at: h.created_at,
        note: h.note,
      })
    }
    if (!seen.has(props.status))
      out.push({ status: props.status, state: 'stopped', at: null, note: '' })
    return out
  }

  const currentIdx = Math.max(0, flow.indexOf(props.status))
  return flow.map((s, i) => {
    const h = latest(s)
    return {
      status: s,
      state: i < currentIdx ? 'done' : i === currentIdx ? 'current' : 'upcoming',
      at: i <= currentIdx ? (h?.created_at ?? null) : null,
      note: i <= currentIdx ? (h?.note ?? '') : '',
    }
  })
})
</script>

<template>
  <ol class="tl">
    <li
      v-for="s in steps"
      :key="s.status"
      class="step"
      :class="`st-${s.state}`"
      :aria-current="s.state === 'current' ? 'step' : undefined"
    >
      <span class="dot" aria-hidden="true">
        <Check
          v-if="s.state === 'done' || (s.state === 'current' && s.status === 'delivered')"
          :size="14"
        />
        <X v-else-if="s.state === 'stopped'" :size="14" />
      </span>
      <div class="txt">
        <span class="name">{{ $t(`status.${s.status}`) }}</span>
        <span v-if="s.at" class="at">{{ dateTime(s.at) }}</span>
        <span v-if="s.note" class="note">{{ s.note }}</span>
      </div>
    </li>
  </ol>
</template>

<style scoped>
.tl {
  list-style: none;
  margin: 0;
  padding: 0;
}
.step {
  position: relative;
  display: flex;
  gap: var(--sp-3);
  padding-bottom: var(--sp-4);
}
.step:last-child {
  padding-bottom: 0;
}
.step:not(:last-child)::before {
  content: '';
  position: absolute;
  left: 11px;
  top: 26px;
  bottom: 2px;
  width: 2px;
  background: var(--c-border);
}
.st-done:not(:last-child)::before {
  background: var(--c-primary-300);
}
.dot {
  flex: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  border: 2px solid var(--c-border-strong);
  background: var(--c-surface);
  color: #fff;
}
.st-done .dot {
  background: var(--c-primary);
  border-color: var(--c-primary);
}
.st-current .dot {
  border-color: var(--c-primary);
  background: var(--c-primary);
  box-shadow: 0 0 0 4px var(--c-primary-100);
}
.st-stopped .dot {
  background: var(--c-danger);
  border-color: var(--c-danger);
}
.txt {
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding-top: 1px;
}
.name {
  font-weight: var(--fw-semibold);
  font-size: var(--fs-sm);
}
.st-upcoming .name {
  color: var(--c-muted);
  font-weight: var(--fw-medium);
}
.st-current .name {
  color: var(--c-primary);
}
.st-stopped .name {
  color: var(--c-danger);
}
.at {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.note {
  font-size: var(--fs-xs);
  color: var(--c-text-2);
  margin-top: 2px;
}
</style>
