<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import ProductBrowser from '@/components/store/ProductBrowser.vue'

const route = useRoute()
const q = computed(() => (typeof route.query.q === 'string' ? route.query.q.trim() : ''))
</script>

<template>
  <div class="container page">
    <h1 class="page-title">
      {{ q ? $t('catalog.searchFor', { q }) : $t('catalog.allProducts') }}
    </h1>
    <ProductBrowser />
  </div>
</template>
