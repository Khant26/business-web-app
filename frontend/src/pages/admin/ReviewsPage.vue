<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Check, X } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { useAdminList } from '@/components/admin/useAdminList'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseTabs from '@/components/ui/BaseTabs.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import RatingStars from '@/components/ui/RatingStars.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import { adminApi } from '@/api/admin'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import type { AdminReview, ReviewStatus } from '@/types/api'
import type { TabItem } from '@/types/ui'

const { t } = useI18n()
const ui = useUiStore()
const { dateTime } = useLocalized()
const { errorText } = useErrorMessage()

const { filters, list } = useAdminList(
  { status: 'pending' as ReviewStatus },
  (page, page_size, f) => adminApi.reviews({ page, page_size, status: f.status }),
  {
    sanitize: (f) => {
      if (!['pending', 'approved', 'rejected'].includes(f.status)) f.status = 'pending'
    },
  },
)

const tabs = computed<TabItem<ReviewStatus>[]>(() => [
  { value: 'pending', label: t('admin.reviews.pending') },
  { value: 'approved', label: t('admin.reviews.approved') },
  { value: 'rejected', label: t('admin.reviews.rejected') },
])

const busy = ref<number | null>(null)
async function moderate(r: AdminReview, action: 'approve' | 'reject') {
  busy.value = r.id
  try {
    if (action === 'approve') await adminApi.approveReview(r.id)
    else await adminApi.rejectReview(r.id)
    ui.success(
      action === 'approve' ? t('admin.reviews.approvedMsg') : t('admin.reviews.rejectedMsg'),
    )
    // The review moves to another tab: drop it locally, then refresh counts/pages.
    const d = list.data.value
    if (d)
      list.data.value = { ...d, items: d.items.filter((x) => x.id !== r.id), total: d.total - 1 }
    if (list.items.value.length === 0 && list.page.value > 1) list.setPage(list.page.value - 1)
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    busy.value = null
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.reviews')" />
    <BaseTabs v-model="filters.status" :tabs="tabs" :label="$t('admin.common.status')" />

    <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
    <div v-else-if="list.loading.value && !list.data.value" class="cards">
      <div v-for="i in 4" :key="i" class="card card-pad">
        <SkeletonBlock :lines="3" height="16px" />
      </div>
    </div>
    <EmptyState v-else-if="list.items.value.length === 0" :title="$t('admin.reviews.empty')" />
    <template v-else>
      <ul class="cards">
        <li v-for="r in list.items.value" :key="r.id" class="card card-pad review">
          <div class="head">
            <div class="who">
              <RouterLink
                :to="{ name: 'admin-product-edit', params: { id: r.product_id } }"
                class="product"
              >
                {{ r.product_name_en }}
              </RouterLink>
              <span class="sub">{{ r.user_email }} · {{ dateTime(r.created_at) }}</span>
            </div>
            <RatingStars :value="r.rating" :size="16" />
          </div>
          <p class="comment" :class="{ empty: !r.comment }">
            {{ r.comment || $t('admin.reviews.noComment') }}
          </p>
          <div class="actions">
            <BaseButton
              v-if="r.status !== 'approved'"
              size="sm"
              :loading="busy === r.id"
              :disabled="busy !== null && busy !== r.id"
              @click="moderate(r, 'approve')"
            >
              <Check :size="16" aria-hidden="true" />{{ $t('admin.reviews.approve') }}
            </BaseButton>
            <BaseButton
              v-if="r.status !== 'rejected'"
              size="sm"
              variant="secondary"
              :disabled="busy !== null"
              @click="moderate(r, 'reject')"
            >
              <X :size="16" aria-hidden="true" />{{ $t('admin.reviews.reject') }}
            </BaseButton>
          </div>
        </li>
      </ul>
      <BasePagination
        :page="list.page.value"
        :pages="list.pages.value"
        @update:page="list.setPage"
      />
    </template>
  </div>
</template>

<style scoped>
.cards {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--sp-3);
}
@media (min-width: 1024px) {
  .cards {
    grid-template-columns: 1fr 1fr;
  }
}
.review {
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--sp-3);
  flex-wrap: wrap;
}
.who {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.product {
  font-weight: var(--fw-semibold);
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
  overflow-wrap: anywhere;
}
.comment {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-size: var(--fs-sm);
}
.comment.empty {
  color: var(--c-muted);
  font-style: italic;
}
.actions {
  display: flex;
  gap: var(--sp-2);
  margin-top: auto;
}
</style>
