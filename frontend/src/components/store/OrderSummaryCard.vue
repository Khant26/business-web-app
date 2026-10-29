<script setup lang="ts">
/** Money breakdown shared by checkout (Quote) and order detail (Order). */
import { formatBaht, formatBp } from '@/utils/money'

defineProps<{
  subtotal: number
  discount: number
  shipping: number
  vatRateBp: number
  vat: number
  total: number
  couponCode?: string
  zoneName?: string
  /** No province chosen yet, so the delivery fee is unknown. */
  shippingPending?: boolean
}>()
</script>

<template>
  <dl class="sum">
    <div class="sum-row">
      <dt>{{ $t('checkout.subtotal') }}</dt>
      <dd>{{ formatBaht(subtotal) }}</dd>
    </div>
    <div v-if="discount > 0" class="sum-row disc">
      <dt>
        {{
          couponCode
            ? $t('checkout.discountWithCode', { code: couponCode })
            : $t('checkout.discount')
        }}
      </dt>
      <dd>−{{ formatBaht(discount) }}</dd>
    </div>
    <div class="sum-row">
      <dt>
        {{ zoneName ? $t('checkout.deliveryZone', { zone: zoneName }) : $t('checkout.delivery') }}
      </dt>
      <dd>
        <template v-if="shippingPending">—</template>
        <span v-else-if="shipping === 0" class="free">{{ $t('common.free') }}</span>
        <template v-else>{{ formatBaht(shipping) }}</template>
      </dd>
    </div>
    <div class="sum-row">
      <dt>{{ $t('checkout.vat', { rate: formatBp(vatRateBp) }) }}</dt>
      <dd>{{ formatBaht(vat) }}</dd>
    </div>
    <div class="sum-row total">
      <dt>{{ $t('checkout.total') }}</dt>
      <dd>{{ formatBaht(total) }}</dd>
    </div>
  </dl>
</template>

<style scoped>
.sum {
  margin: 0;
  font-size: var(--fs-sm);
}
dt {
  color: var(--c-text-2);
}
dd {
  margin: 0;
  font-weight: var(--fw-medium);
  white-space: nowrap;
}
.total dt,
.total dd {
  color: var(--c-text);
  font-weight: var(--fw-bold);
}
.disc dd {
  color: var(--c-success);
}
.free {
  color: var(--c-success);
  font-weight: var(--fw-semibold);
}
</style>
