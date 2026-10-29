<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import RatingStars from '@/components/ui/RatingStars.vue'
import PriceTag from './PriceTag.vue'
import WishlistButton from './WishlistButton.vue'
import { useLocalized } from '@/composables/useLocalized'
import { percentOff } from '@/utils/money'
import type { ProductCard } from '@/types/api'

const props = withDefaults(defineProps<{ product: ProductCard; wishlist?: boolean }>(), {
  wishlist: true,
})
defineEmits<{ wishlistChanged: [boolean] }>()

const { loc } = useLocalized()
const name = computed(() => loc(props.product, 'name'))
const off = computed(() => percentOff(props.product.price_satang, props.product.compare_at_satang))
const to = computed(() => ({ name: 'product', params: { slug: props.product.slug } }))
</script>

<template>
  <article class="pc" :class="{ oos: !product.in_stock }">
    <div class="media">
      <RouterLink :to="to" tabindex="-1" aria-hidden="true">
        <ImagePlaceholder
          :src="product.image_url"
          :name="name"
          :alt="name"
          ratio="1 / 1"
          rounded="0"
        />
      </RouterLink>
      <div class="badges">
        <span v-if="off > 0" class="b sale">{{ $t('product.off', { n: off }) }}</span>
        <span v-if="!product.in_stock" class="b oos-b">{{ $t('product.outOfStock') }}</span>
      </div>
      <WishlistButton
        v-if="wishlist"
        class="wish"
        :product-id="product.id"
        @changed="$emit('wishlistChanged', $event)"
      />
    </div>
    <div class="body">
      <p v-if="product.category" class="cat">{{ loc(product.category, 'name') }}</p>
      <h3 class="name">
        <RouterLink :to="to" class="link">{{ name }}</RouterLink>
      </h3>
      <div class="meta">
        <template v-if="product.rating_count > 0">
          <RatingStars :value="product.rating_avg" :size="13" />
          <span class="muted">({{ product.rating_count }})</span>
        </template>
        <span v-if="product.sold_count > 0" class="muted sold">{{
          $t('product.sold', { n: product.sold_count })
        }}</span>
      </div>
      <PriceTag
        class="price"
        :price="product.price_satang"
        :compare-at="product.compare_at_satang"
      />
    </div>
  </article>
</template>

<style scoped>
.pc {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition:
    box-shadow var(--dur) var(--ease),
    transform var(--dur) var(--ease),
    border-color var(--dur) var(--ease);
  height: 100%;
}
.pc:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--c-primary-200);
}
@media (hover: hover) {
  .pc:hover {
    transform: translateY(-2px);
  }
}
.pc:focus-within {
  border-color: var(--c-primary-300);
}
.media {
  position: relative;
}
.oos .media :deep(.ph) {
  opacity: 0.6;
}
.badges {
  position: absolute;
  top: var(--sp-2);
  left: var(--sp-2);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}
.b {
  font-size: 11px;
  font-weight: var(--fw-bold);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  line-height: 1.5;
}
.sale {
  background: var(--c-accent-strong);
  color: #fff;
}
.oos-b {
  background: rgb(22 24 39 / 0.78);
  color: #fff;
}
.wish {
  position: absolute;
  top: var(--sp-2);
  right: var(--sp-2);
}
.body {
  padding: var(--sp-3);
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}
@media (min-width: 640px) {
  .body {
    padding: var(--sp-3) var(--sp-4) var(--sp-4);
  }
}
.cat {
  margin: 0;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--c-muted);
  font-weight: var(--fw-semibold);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.name {
  margin: 0;
  font-size: var(--fs-sm);
  font-weight: var(--fw-semibold);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 2.8em;
}
.link {
  color: var(--c-text);
}
.link::after {
  content: '';
  position: absolute;
  inset: 0;
  z-index: 0;
}
.link:hover {
  text-decoration: none;
  color: var(--c-primary);
}
.wish {
  z-index: 1;
}
.meta {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: var(--fs-xs);
  min-height: 18px;
  flex-wrap: wrap;
}
.sold {
  margin-left: auto;
}
.price {
  margin-top: auto;
  padding-top: var(--sp-1);
}
</style>
