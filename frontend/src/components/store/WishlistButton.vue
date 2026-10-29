<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Heart } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useWishlistStore } from '@/stores/wishlist'
import { useUiStore } from '@/stores/ui'
import { useErrorMessage } from '@/composables/useErrorMessage'

const props = withDefaults(defineProps<{ productId: number; variant?: 'float' | 'outline' }>(), {
  variant: 'float',
})
const emit = defineEmits<{ changed: [boolean] }>()

const auth = useAuthStore()
const wishlist = useWishlistStore()
const ui = useUiStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const { errorText } = useErrorMessage()
const busy = ref(false)

const on = computed(() => wishlist.has(props.productId))

async function toggle() {
  if (!auth.isLoggedIn) {
    void router.push({ name: 'login', query: { redirect: route.fullPath } })
    return
  }
  if (busy.value) return
  busy.value = true
  try {
    const now = await wishlist.toggle(props.productId)
    ui.success(now ? t('product.wishlistAdded') : t('product.wishlistRemoved'))
    emit('changed', now)
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <button
    type="button"
    class="wish"
    :class="[`v-${variant}`, { on }]"
    :aria-pressed="on"
    :aria-label="on ? $t('product.removeWishlist') : $t('product.addWishlist')"
    :disabled="busy"
    @click.prevent.stop="toggle"
  >
    <Heart :size="variant === 'float' ? 18 : 20" aria-hidden="true" />
  </button>
</template>

<style scoped>
.wish {
  display: grid;
  place-items: center;
  border: 0;
  cursor: pointer;
  color: var(--c-text-2);
  transition:
    transform var(--dur-fast) var(--ease),
    color var(--dur-fast);
}
.wish:active {
  transform: scale(0.9);
}
.v-float {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgb(255 255 255 / 0.92);
  box-shadow: var(--shadow-sm);
}
.v-outline {
  width: 50px;
  height: 50px;
  border-radius: var(--radius-md);
  border: 1px solid var(--c-border-strong);
  background: var(--c-surface);
}
.v-outline:hover {
  border-color: #e11d48;
}
.wish:hover {
  color: #e11d48;
}
.on {
  color: #e11d48;
}
.on :deep(svg) {
  fill: currentColor;
}
.wish:disabled {
  opacity: 0.7;
}
</style>
