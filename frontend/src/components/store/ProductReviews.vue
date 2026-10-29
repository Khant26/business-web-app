<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { MessageSquareText, PenLine } from 'lucide-vue-next'
import RatingStars from '@/components/ui/RatingStars.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { useAsync } from '@/composables/useAsync'
import { useForm, msg } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useLocalized } from '@/composables/useLocalized'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { isApiError } from '@/api/client'
import { catalogApi } from '@/api/catalog'

const props = defineProps<{ slug: string; ratingAvg: number; ratingCount: number }>()

const PAGE_SIZE = 5
const auth = useAuthStore()
const ui = useUiStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { errorText } = useErrorMessage()
const { date } = useLocalized()

const page = ref(1)
const reviews = useAsync((slug: string, p: number) => catalogApi.reviews(slug, p, PAGE_SIZE))
watch(
  () => [props.slug, page.value] as const,
  ([s, p]) => void reviews.run(s, p),
  { immediate: true },
)
const pages = computed(() => Math.max(1, Math.ceil((reviews.data.value?.total ?? 0) / PAGE_SIZE)))

const schema = z.object({
  rating: z.number().int().min(1, msg('validation.rating')).max(5, msg('validation.rating')),
  comment: z
    .string()
    .trim()
    .max(1000, msg('validation.maxLength', { max: 1000 })),
})
const form = useForm(schema, { rating: 0, comment: '' })
const writing = ref(route.hash === '#write-review')
const blocked = ref<string | null>(null)
const submitted = ref(false)

function startWriting() {
  if (!auth.isLoggedIn) {
    void router.push({ name: 'login', query: { redirect: `${route.path}#write-review` } })
    return
  }
  writing.value = true
}

async function submit() {
  const data = form.validate()
  if (!data) return
  form.submitting.value = true
  blocked.value = null
  try {
    await catalogApi.createReview(props.slug, data)
    submitted.value = true
    writing.value = false
    form.reset()
    ui.success(t('product.reviewSubmitted'))
  } catch (e) {
    if (isApiError(e) && (e.code === 'not_purchased' || e.code === 'already_reviewed')) {
      blocked.value = errorText(e)
    } else if (!form.applyApiError(e)) {
      ui.error(errorText(e))
    }
  } finally {
    form.submitting.value = false
  }
}
</script>

<template>
  <section id="reviews" class="reviews" aria-labelledby="h-reviews">
    <div class="head">
      <div>
        <h2 id="h-reviews">{{ $t('product.reviews') }}</h2>
        <div class="summary">
          <span class="avg">{{ ratingCount ? ratingAvg.toFixed(1) : '–' }}</span>
          <div>
            <RatingStars :value="ratingAvg" :size="18" />
            <p class="muted text-sm">{{ $t('product.reviewsCount', ratingCount) }}</p>
          </div>
        </div>
      </div>
      <BaseButton v-if="!writing && !submitted" variant="secondary" @click="startWriting">
        <PenLine :size="16" aria-hidden="true" />{{ $t('product.writeReview') }}
      </BaseButton>
    </div>

    <p v-if="submitted" class="note ok" role="status">{{ $t('product.reviewSubmitted') }}</p>

    <form
      v-if="writing"
      id="write-review"
      class="card card-pad form"
      novalidate
      @submit.prevent="submit"
    >
      <p v-if="blocked" class="note warn" role="alert">{{ blocked }}</p>
      <div>
        <RatingStars v-model:value="form.values.rating" editable :size="20" />
        <p v-if="form.errors.value.rating" class="field-err" role="alert">
          {{ form.errors.value.rating }}
        </p>
      </div>
      <BaseInput
        v-model="form.values.comment"
        multiline
        :rows="4"
        maxlength="1000"
        :label="$t('product.yourComment')"
        :placeholder="$t('product.commentPlaceholder')"
        :error="form.errors.value.comment"
        :hint="`${form.values.comment.length}/1000`"
      />
      <div class="row">
        <BaseButton type="submit" :loading="form.submitting.value">{{
          $t('product.submitReview')
        }}</BaseButton>
        <BaseButton variant="ghost" @click="writing = false">{{ $t('common.cancel') }}</BaseButton>
      </div>
    </form>

    <ErrorState
      v-if="reviews.error.value"
      compact
      :error="reviews.error.value"
      @retry="reviews.retry"
    />
    <ul v-else-if="reviews.loading.value && !reviews.data.value" class="list">
      <li v-for="i in 2" :key="i" class="item"><SkeletonBlock :lines="3" /></li>
    </ul>
    <EmptyState
      v-else-if="!reviews.data.value?.items.length"
      compact
      :icon="MessageSquareText"
      :title="$t('product.noReviews')"
    />
    <ul v-else class="list">
      <li v-for="r in reviews.data.value.items" :key="r.id" class="item">
        <div class="item-head">
          <span class="avatar" aria-hidden="true">{{
            (r.user_name || '?').charAt(0).toUpperCase()
          }}</span>
          <div>
            <strong>{{ r.user_name || '—' }}</strong>
            <div class="row-s">
              <RatingStars :value="r.rating" :size="13" />
              <time class="muted text-xs" :datetime="r.created_at">{{ date(r.created_at) }}</time>
            </div>
          </div>
        </div>
        <p v-if="r.comment" class="comment">{{ r.comment }}</p>
      </li>
    </ul>
    <BasePagination v-model:page="page" :pages="pages" />
  </section>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--sp-4);
  flex-wrap: wrap;
  margin-bottom: var(--sp-4);
}
.summary {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
}
.summary p {
  margin: 0;
}
.avg {
  font-size: 2.5rem;
  font-weight: 800;
  line-height: 1;
  font-variant-numeric: tabular-nums;
}
.form {
  display: flex;
  flex-direction: column;
  gap: var(--sp-4);
  margin-bottom: var(--sp-5);
}
.field-err {
  color: var(--c-danger);
  font-size: var(--fs-xs);
  margin: 4px 0 0;
}
.note {
  margin: 0 0 var(--sp-4);
  padding: var(--sp-3) var(--sp-4);
  border-radius: var(--radius-md);
  font-size: var(--fs-sm);
}
.form .note {
  margin: 0;
}
.ok {
  background: var(--c-success-bg);
  color: var(--c-success);
}
.warn {
  background: var(--c-warning-bg);
  color: var(--c-warning);
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.item {
  padding: var(--sp-4) 0;
  border-bottom: 1px solid var(--c-border);
}
.item-head {
  display: flex;
  gap: var(--sp-3);
  align-items: center;
}
.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--c-surface-3);
  display: grid;
  place-items: center;
  font-weight: var(--fw-bold);
  color: var(--c-text-2);
  flex: none;
}
.row-s {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
}
.comment {
  margin: var(--sp-2) 0 0;
  white-space: pre-line;
  color: var(--c-text-2);
}
</style>
