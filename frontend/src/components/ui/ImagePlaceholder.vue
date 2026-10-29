<script setup lang="ts">
/**
 * Shows `src` when present (and loads), otherwise a tinted block with the
 * name's initial and a soft icon — products in seed data have no images.
 */
import { computed, ref, watch } from 'vue'
import { ShoppingBag } from 'lucide-vue-next'
import { hueFrom, initial } from '@/utils/misc'

const props = withDefaults(
  defineProps<{
    src?: string | null
    name?: string
    alt?: string
    ratio?: string
    rounded?: string
    small?: boolean
    eager?: boolean
  }>(),
  { ratio: '1 / 1', rounded: 'var(--radius-md)' },
)

const failed = ref(false)
watch(
  () => props.src,
  () => (failed.value = false),
)

const showImg = computed(() => !!props.src && !failed.value)
const hue = computed(() => hueFrom(props.name ?? ''))
const letter = computed(() => initial(props.name))
</script>

<template>
  <div class="ph" :style="{ aspectRatio: ratio, borderRadius: rounded, '--h': hue }">
    <img
      v-if="showImg"
      :src="src ?? ''"
      :alt="alt ?? name ?? ''"
      :loading="eager ? 'eager' : 'lazy'"
      decoding="async"
      @error="failed = true"
    />
    <div
      v-else
      class="fallback"
      :class="{ small }"
      role="img"
      :aria-label="alt ?? name ?? $t('product.noImage')"
    >
      <span class="letter" aria-hidden="true">{{ letter }}</span>
      <ShoppingBag v-if="!small" class="bag" :size="22" aria-hidden="true" />
    </div>
  </div>
</template>

<style scoped>
.ph {
  position: relative;
  width: 100%;
  overflow: hidden;
  background: var(--c-surface-2);
}
img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.fallback {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background:
    radial-gradient(circle at 25% 20%, hsl(var(--h) 80% 96%) 0%, transparent 55%),
    linear-gradient(135deg, hsl(var(--h) 60% 92%) 0%, hsl(calc(var(--h) + 40) 55% 86%) 100%);
  color: hsl(var(--h) 45% 32%);
}
.letter {
  font-size: clamp(1.5rem, 9cqi, 3.5rem);
  font-weight: var(--fw-bold);
  letter-spacing: -0.02em;
  opacity: 0.85;
  line-height: 1;
}
.ph {
  container-type: inline-size;
}
.small .letter {
  font-size: clamp(0.9rem, 32cqi, 1.6rem);
}
.bag {
  position: absolute;
  right: 10%;
  bottom: 10%;
  opacity: 0.35;
}
</style>
