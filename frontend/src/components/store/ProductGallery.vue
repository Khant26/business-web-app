<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import type { ProductImage } from '@/types/api'

const props = defineProps<{ images: ProductImage[]; name: string }>()
const sorted = computed(() => [...props.images].sort((a, b) => a.sort_order - b.sort_order))
const index = ref(0)
watch(
  () => props.images,
  () => (index.value = 0),
)
const current = computed(() => sorted.value[index.value])
const go = (i: number) => {
  const n = sorted.value.length
  if (n) index.value = (i + n) % n
}

let startX = 0
const onTouchStart = (e: TouchEvent) => (startX = e.touches[0]?.clientX ?? 0)
function onTouchEnd(e: TouchEvent) {
  const dx = (e.changedTouches[0]?.clientX ?? 0) - startX
  if (Math.abs(dx) > 40) go(index.value + (dx < 0 ? 1 : -1))
}
</script>

<template>
  <div class="gallery" :aria-label="$t('product.gallery')" role="group">
    <div class="main" @touchstart.passive="onTouchStart" @touchend.passive="onTouchEnd">
      <ImagePlaceholder
        :src="current?.url"
        :name="name"
        :alt="name"
        ratio="1 / 1"
        rounded="var(--radius-xl)"
        eager
      />
      <template v-if="sorted.length > 1">
        <button
          type="button"
          class="nav prev"
          :aria-label="$t('common.previous')"
          @click="go(index - 1)"
        >
          <ChevronLeft :size="20" aria-hidden="true" />
        </button>
        <button
          type="button"
          class="nav next"
          :aria-label="$t('common.next')"
          @click="go(index + 1)"
        >
          <ChevronRight :size="20" aria-hidden="true" />
        </button>
        <span class="pos" aria-hidden="true">{{ index + 1 }} / {{ sorted.length }}</span>
      </template>
    </div>
    <ul v-if="sorted.length > 1" class="thumbs">
      <li v-for="(img, i) in sorted" :key="img.id">
        <button
          type="button"
          class="thumb"
          :class="{ on: i === index }"
          :aria-label="$t('product.showImage', { n: i + 1 })"
          :aria-current="i === index || undefined"
          @click="index = i"
        >
          <ImagePlaceholder :src="img.url" :name="name" alt="" small />
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.main {
  position: relative;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-xl);
  background: var(--c-surface);
}
.nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 0;
  background: rgb(255 255 255 / 0.92);
  box-shadow: var(--shadow-md);
  display: grid;
  place-items: center;
  cursor: pointer;
}
.prev {
  left: var(--sp-3);
}
.next {
  right: var(--sp-3);
}
.pos {
  position: absolute;
  bottom: var(--sp-3);
  right: var(--sp-3);
  background: rgb(22 24 39 / 0.7);
  color: #fff;
  font-size: var(--fs-xs);
  padding: 2px 8px;
  border-radius: 999px;
}
.thumbs {
  list-style: none;
  margin: var(--sp-3) 0 0;
  padding: 0;
  display: flex;
  gap: var(--sp-2);
  overflow-x: auto;
}
.thumb {
  width: 68px;
  padding: 0;
  border: 2px solid transparent;
  border-radius: var(--radius-md);
  overflow: hidden;
  cursor: pointer;
  background: none;
}
.thumb.on {
  border-color: var(--c-primary);
}
</style>
