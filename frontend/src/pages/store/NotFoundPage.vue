<script setup lang="ts">
import { Compass } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
</script>

<template>
  <div class="container page nf">
    <div class="code" aria-hidden="true">404</div>
    <Compass :size="40" class="ic" aria-hidden="true" />
    <h1>{{ $t('notFound.title') }}</h1>
    <p class="muted">{{ $t('notFound.text') }}</p>
    <div class="row">
      <BaseButton :to="{ name: 'home' }">{{ $t('notFound.home') }}</BaseButton>
      <BaseButton :to="{ name: 'products' }" variant="secondary">{{
        $t('home.browse')
      }}</BaseButton>
    </div>
  </div>
</template>

<style scoped>
.nf {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding-block: var(--sp-16);
  position: relative;
}
.code {
  font-size: clamp(6rem, 30vw, 12rem);
  font-weight: 900;
  line-height: 1;
  color: var(--c-primary-100);
  letter-spacing: -0.05em;
  user-select: none;
}
.ic {
  color: var(--c-primary);
  margin: calc(var(--sp-8) * -1) 0 var(--sp-4);
}
.row {
  justify-content: center;
  margin-top: var(--sp-4);
}
</style>
