<script setup lang="ts">
import { computed, useId } from 'vue'
import { Star } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    value?: number
    size?: number
    editable?: boolean
    count?: number
    label?: string
  }>(),
  { value: 0, size: 16 },
)
const emit = defineEmits<{ 'update:value': [number] }>()
const name = `r-${useId()}`

const fills = computed(() =>
  [1, 2, 3, 4, 5].map((i) => Math.max(0, Math.min(1, props.value - (i - 1)))),
)
</script>

<template>
  <fieldset v-if="editable" class="stars edit">
    <legend class="sr-only">{{ label ?? $t('product.yourRating') }}</legend>
    <template v-for="i in 5" :key="i">
      <input
        :id="`${name}-${i}`"
        type="radio"
        :name="name"
        :value="i"
        :checked="value === i"
        class="sr-only"
        @change="emit('update:value', i)"
      />
      <label :for="`${name}-${i}`" :class="{ on: i <= value }" :title="String(i)">
        <Star :size="size + 12" aria-hidden="true" />
        <span class="sr-only">{{ i }}</span>
      </label>
    </template>
  </fieldset>
  <span
    v-else
    class="stars"
    role="img"
    :aria-label="$t('product.rating', { rating: value.toFixed(1) })"
  >
    <span
      v-for="(f, i) in fills"
      :key="i"
      class="star"
      :style="{ width: `${size}px`, height: `${size}px` }"
    >
      <Star :size="size" class="bg" aria-hidden="true" />
      <span class="fg" :style="{ width: `${f * 100}%` }">
        <Star :size="size" aria-hidden="true" />
      </span>
    </span>
    <span v-if="count !== undefined" class="count">({{ count }})</span>
  </span>
</template>

<style scoped>
.stars {
  display: inline-flex;
  align-items: center;
  gap: 1px;
}
.star {
  position: relative;
  display: inline-block;
}
.bg {
  color: var(--c-border-strong);
}
.fg {
  position: absolute;
  inset: 0 auto 0 0;
  overflow: hidden;
  color: var(--c-star);
}
.fg :deep(svg) {
  fill: currentColor;
}
.count {
  margin-left: 4px;
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.edit {
  border: 0;
  padding: 0;
  margin: 0;
  gap: 4px;
}
.edit label {
  cursor: pointer;
  color: var(--c-border-strong);
  border-radius: 6px;
  padding: 2px;
  display: grid;
}
.edit label.on {
  color: var(--c-star);
}
.edit label.on :deep(svg) {
  fill: currentColor;
}
.edit input:focus-visible + label {
  box-shadow: var(--ring);
}
.edit label:hover {
  transform: scale(1.08);
}
</style>
