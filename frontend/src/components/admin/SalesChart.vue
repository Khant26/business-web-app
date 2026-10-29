<script setup lang="ts">
/** 7-day sales bar chart in plain SVG with an accessible data table fallback. */
import { computed, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatBaht, formatBahtShort } from '@/utils/money'

const props = defineProps<{ data: { date: string; total_satang: number; orders: number }[] }>()
const { t, locale } = useI18n()
const titleId = `sc-${useId()}`

const W = 560
const H = 220
const PAD_L = 8
const PAD_R = 8
const PAD_T = 26
const PAD_B = 34

const max = computed(() => Math.max(1, ...props.data.map((d) => d.total_satang)))

function dayLabel(date: string) {
  const d = new Date(`${date}T00:00:00+07:00`)
  if (Number.isNaN(d.getTime())) return date
  return d.toLocaleDateString(locale.value === 'my' ? 'my-MM' : 'en-GB', {
    weekday: 'short',
    day: 'numeric',
    timeZone: 'Asia/Bangkok',
  })
}

const bars = computed(() => {
  const n = Math.max(1, props.data.length)
  const slot = (W - PAD_L - PAD_R) / n
  const bw = Math.min(48, slot * 0.6)
  const innerH = H - PAD_T - PAD_B
  return props.data.map((d, i) => {
    const h = d.total_satang > 0 ? Math.max(3, (d.total_satang / max.value) * innerH) : 2
    const x = PAD_L + slot * i + (slot - bw) / 2
    return {
      ...d,
      x,
      y: H - PAD_B - h,
      w: bw,
      h,
      cx: PAD_L + slot * i + slot / 2,
      label: dayLabel(d.date),
      last: i === props.data.length - 1,
    }
  })
})
</script>

<template>
  <figure class="chart">
    <svg
      :viewBox="`0 0 ${W} ${H}`"
      role="img"
      :aria-labelledby="titleId"
      preserveAspectRatio="none"
    >
      <title :id="titleId">{{ t('admin.dashboard.chartLabel') }}</title>
      <line :x1="0" :x2="W" :y1="H - PAD_B" :y2="H - PAD_B" class="axis" />
      <g v-for="b in bars" :key="b.date">
        <rect :x="b.x" :y="b.y" :width="b.w" :height="b.h" rx="6" :class="{ today: b.last }">
          <title>
            {{ b.label }}: {{ formatBaht(b.total_satang) }} ·
            {{ t('admin.dashboard.ordersN', { n: b.orders }) }}
          </title>
        </rect>
        <text v-if="b.total_satang > 0" :x="b.cx" :y="b.y - 8" class="val">
          {{ formatBahtShort(b.total_satang) }}
        </text>
        <text :x="b.cx" :y="H - 12" class="lbl">{{ b.label }}</text>
      </g>
    </svg>
    <table class="sr-only">
      <caption>
        {{
          t('admin.dashboard.sales7d')
        }}
      </caption>
      <thead>
        <tr>
          <th scope="col">{{ t('admin.dashboard.date') }}</th>
          <th scope="col">{{ t('admin.dashboard.sales') }}</th>
          <th scope="col">{{ t('admin.dashboard.orders') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in data" :key="d.date">
          <th scope="row">{{ d.date }}</th>
          <td>{{ formatBaht(d.total_satang) }}</td>
          <td>{{ d.orders }}</td>
        </tr>
      </tbody>
    </table>
  </figure>
</template>

<style scoped>
.chart {
  margin: 0;
  width: 100%;
}
svg {
  width: 100%;
  height: 220px;
  display: block;
  overflow: visible;
}
rect {
  fill: var(--c-primary-300);
  transition: fill var(--dur-fast);
}
rect.today {
  fill: var(--c-primary);
}
rect:hover {
  fill: var(--c-primary-600);
}
.axis {
  stroke: var(--c-border);
  stroke-width: 1;
}
text {
  font-size: 11px;
  text-anchor: middle;
  fill: var(--c-muted);
  font-family: var(--font-sans);
}
.val {
  fill: var(--c-text-2);
  font-weight: 600;
  font-size: 10.5px;
}
@media (max-width: 480px) {
  .val {
    display: none;
  }
}
</style>
