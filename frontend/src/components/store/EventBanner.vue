<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowRight, Flame } from 'lucide-vue-next'
import EventCountdown from './EventCountdown.vue'
import { useLocalized } from '@/composables/useLocalized'
import { hueFrom } from '@/utils/misc'
import type { ShopEvent } from '@/types/api'

const props = withDefaults(defineProps<{ event: ShopEvent; hero?: boolean; link?: boolean }>(), {
  link: true,
})
const { loc } = useLocalized()
const title = computed(() => loc(props.event, 'title'))
const desc = computed(() => loc(props.event, 'description'))
const hue = computed(() => hueFrom(props.event.slug))
</script>

<template>
  <section class="eb" :class="{ hero }" :style="{ '--h': hue }">
    <img
      v-if="event.banner_url"
      :src="event.banner_url"
      alt=""
      class="bg"
      :loading="hero ? 'eager' : 'lazy'"
    />
    <div v-else class="bg pattern" aria-hidden="true" />
    <div class="shade" aria-hidden="true" />
    <div class="content">
      <span class="kicker"><Flame :size="14" aria-hidden="true" />{{ $t('nav.events') }}</span>
      <component :is="hero ? 'h2' : 'h1'" class="title">{{ title }}</component>
      <p v-if="desc" class="desc">{{ desc }}</p>
      <div class="row-b">
        <EventCountdown :ends-at="event.ends_at" tone="light" />
        <RouterLink v-if="link" :to="{ name: 'event', params: { slug: event.slug } }" class="cta">
          {{ $t('home.heroCta') }}<ArrowRight :size="16" aria-hidden="true" />
        </RouterLink>
      </div>
    </div>
  </section>
</template>

<style scoped>
.eb {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-xl);
  min-height: 220px;
  display: flex;
  align-items: flex-end;
  color: #fff;
  isolation: isolate;
}
@media (min-width: 768px) {
  .eb {
    min-height: 300px;
  }
  .eb.hero {
    min-height: 360px;
  }
}
.bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: -2;
}
.pattern {
  background:
    radial-gradient(circle at 85% 15%, hsl(calc(var(--h) + 30) 95% 65% / 0.9) 0%, transparent 40%),
    radial-gradient(circle at 10% 90%, hsl(25 95% 55% / 0.85) 0%, transparent 45%),
    linear-gradient(
      135deg,
      var(--c-primary-800) 0%,
      var(--c-primary) 55%,
      hsl(var(--h) 70% 45%) 100%
    );
}
.pattern::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgb(255 255 255 / 0.14) 1px, transparent 1px);
  background-size: 18px 18px;
}
.shade {
  position: absolute;
  inset: 0;
  z-index: -1;
  background: linear-gradient(
    to top,
    rgb(15 17 30 / 0.65) 0%,
    rgb(15 17 30 / 0.1) 60%,
    transparent 100%
  );
}
.content {
  padding: var(--sp-5);
  width: 100%;
  max-width: 680px;
}
@media (min-width: 768px) {
  .content {
    padding: var(--sp-10);
  }
}
.kicker {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-xs);
  font-weight: var(--fw-bold);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  background: var(--c-accent-strong);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  margin-bottom: var(--sp-3);
}
.title {
  margin: 0 0 var(--sp-2);
  font-size: var(--fs-2xl);
  font-weight: 800;
  text-shadow: 0 2px 12px rgb(0 0 0 / 0.2);
}
@media (min-width: 768px) {
  .title {
    font-size: 2.6rem;
  }
}
.desc {
  margin: 0 0 var(--sp-4);
  color: rgb(255 255 255 / 0.9);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.row-b {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--sp-3);
}
.cta {
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
  height: 42px;
  padding: 0 var(--sp-5);
  border-radius: var(--radius-full);
  background: #fff;
  color: var(--c-primary-hover);
  font-weight: var(--fw-bold);
  font-size: var(--fs-sm);
}
.cta:hover {
  text-decoration: none;
  background: var(--c-primary-50);
}
</style>
