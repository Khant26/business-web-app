<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import EventBanner from './EventBanner.vue'
import type { ShopEvent } from '@/types/api'

const props = defineProps<{ events: ShopEvent[] }>()
const index = ref(0)
const paused = ref(false)
const count = computed(() => props.events.length)
let timer: ReturnType<typeof setInterval> | undefined

const reduceMotion =
  typeof window !== 'undefined' && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

function go(i: number) {
  if (count.value === 0) return
  index.value = (i + count.value) % count.value
}

onMounted(() => {
  if (reduceMotion) return
  timer = setInterval(() => {
    if (!paused.value && count.value > 1 && !document.hidden) go(index.value + 1)
  }, 6000)
})
onBeforeUnmount(() => timer && clearInterval(timer))

let startX = 0
function onTouchStart(e: TouchEvent) {
  startX = e.touches[0]?.clientX ?? 0
}
function onTouchEnd(e: TouchEvent) {
  const dx = (e.changedTouches[0]?.clientX ?? 0) - startX
  if (Math.abs(dx) > 50) go(index.value + (dx < 0 ? 1 : -1))
}
</script>

<template>
  <div
    class="carousel"
    role="region"
    aria-roledescription="carousel"
    :aria-label="$t('nav.events')"
    @mouseenter="paused = true"
    @mouseleave="paused = false"
    @focusin="paused = true"
    @focusout="paused = false"
    @touchstart.passive="onTouchStart"
    @touchend.passive="onTouchEnd"
  >
    <div class="track" :style="{ transform: `translateX(-${index * 100}%)` }">
      <div
        v-for="(ev, i) in events"
        :key="ev.id"
        class="slide"
        role="group"
        aria-roledescription="slide"
        :aria-label="`${i + 1} / ${count}`"
        :aria-hidden="i !== index"
        :inert="i !== index"
      >
        <EventBanner :event="ev" hero />
      </div>
    </div>
    <template v-if="count > 1">
      <button
        type="button"
        class="nav prev"
        :aria-label="$t('home.prevSlide')"
        @click="go(index - 1)"
      >
        <ChevronLeft :size="20" aria-hidden="true" />
      </button>
      <button
        type="button"
        class="nav next"
        :aria-label="$t('home.nextSlide')"
        @click="go(index + 1)"
      >
        <ChevronRight :size="20" aria-hidden="true" />
      </button>
      <div class="dots">
        <button
          v-for="(ev, i) in events"
          :key="ev.id"
          type="button"
          class="dot"
          :class="{ on: i === index }"
          :aria-label="$t('home.goToSlide', { n: i + 1 })"
          :aria-current="i === index || undefined"
          @click="go(i)"
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.carousel {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-xl);
}
.track {
  display: flex;
  transition: transform var(--dur-slow) var(--ease);
}
.slide {
  flex: 0 0 100%;
  min-width: 0;
}
.nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 0;
  background: rgb(255 255 255 / 0.9);
  color: var(--c-text);
  display: none;
  place-items: center;
  cursor: pointer;
  box-shadow: var(--shadow-md);
}
.nav:hover {
  background: #fff;
}
.prev {
  left: var(--sp-3);
}
.next {
  right: var(--sp-3);
}
@media (min-width: 768px) {
  .nav {
    display: grid;
  }
}
.dots {
  position: absolute;
  right: var(--sp-5);
  top: var(--sp-5);
  display: flex;
  gap: 6px;
}
.dot {
  width: 8px;
  height: 8px;
  padding: 0;
  border-radius: 999px;
  border: 0;
  background: rgb(255 255 255 / 0.5);
  cursor: pointer;
  transition: width var(--dur) var(--ease);
}
.dot.on {
  width: 22px;
  background: #fff;
}
</style>
