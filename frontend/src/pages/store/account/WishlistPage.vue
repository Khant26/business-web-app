<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Heart } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import ProductGrid from '@/components/store/ProductGrid.vue'
import { accountApi } from '@/api/account'
import { useAsync } from '@/composables/useAsync'

const state = useAsync(() => accountApi.wishlist())
onMounted(() => void state.run())
const products = computed(() => state.data.value ?? [])

function onChanged(productId: number, on: boolean) {
  if (!on) state.set(products.value.filter((p) => p.id !== productId))
}
</script>

<template>
  <div>
    <h1 class="h">{{ $t('nav.wishlist') }}</h1>
    <ProductGrid
      v-if="state.loading.value && !state.data.value"
      :products="[]"
      loading
      dense
      :skeletons="6"
    />
    <ErrorState v-else-if="state.error.value" :error="state.error.value" @retry="state.retry" />
    <EmptyState
      v-else-if="products.length === 0"
      :icon="Heart"
      :title="$t('account.wishlist.empty')"
      :text="$t('account.wishlist.emptyText')"
    >
      <BaseButton :to="{ name: 'products' }">{{ $t('home.browse') }}</BaseButton>
    </EmptyState>
    <ProductGrid v-else :products="products" dense @wishlist-changed="onChanged" />
  </div>
</template>

<style scoped>
.h {
  font-size: var(--fs-xl);
}
</style>
