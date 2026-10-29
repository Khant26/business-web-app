<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Archive, ExternalLink, Save } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import ImageUploader from '@/components/admin/ImageUploader.vue'
import VariantEditor from '@/components/admin/VariantEditor.vue'
import { optionalText, requiredText, slugField } from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseTabs from '@/components/ui/BaseTabs.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import { adminApi } from '@/api/admin'
import { useAsync } from '@/composables/useAsync'
import { useForm } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { usePageTitle } from '@/composables/usePageTitle'
import { useUiStore } from '@/stores/ui'
import { slugify } from '@/utils/misc'
import type { AdminProduct, AdminVariant, ProductImage, ProductStatus } from '@/types/api'
import type { SelectOption, TabItem } from '@/types/ui'

const props = defineProps<{ id?: number }>()
const router = useRouter()
const { t } = useI18n()
const ui = useUiStore()
const { errorText } = useErrorMessage()

const isNew = computed(() => props.id === undefined)

const product = useAsync((id: number) => adminApi.product(id))
const categories = useAsync(() => adminApi.categories())
const events = useAsync(() => adminApi.events())

const schema = z.object({
  name_en: requiredText(200),
  name_my: optionalText(200),
  description_en: optionalText(5000),
  description_my: optionalText(5000),
  slug: slugField(),
  category_id: z.number().nullable(),
  status: z.enum(['draft', 'active', 'archived']),
  featured: z.boolean(),
  event_ids: z.array(z.number()),
})

const form = useForm(schema, {
  name_en: '',
  name_my: '',
  description_en: '',
  description_my: '',
  slug: '',
  category_id: null,
  status: 'draft',
  featured: false,
  event_ids: [],
})

const images = ref<ProductImage[]>([])
const variants = ref<AdminVariant[]>([])
const slugTouched = ref(false)
const langTab = ref<'en' | 'my'>('en')
const formError = ref('')

function fill(p: AdminProduct) {
  form.reset({
    name_en: p.name_en,
    name_my: p.name_my,
    description_en: p.description_en,
    description_my: p.description_my,
    slug: p.slug,
    category_id: p.category_id,
    status: p.status,
    featured: p.featured,
    event_ids: [...p.event_ids],
  })
  images.value = [...p.images].sort((a, b) => a.sort_order - b.sort_order)
  variants.value = [...p.variants].sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
  slugTouched.value = true
}

onMounted(async () => {
  void categories.run()
  void events.run()
  if (props.id !== undefined) {
    const p = await product.run(props.id)
    if (p) fill(p)
  }
})

usePageTitle(() => (isNew.value ? t('admin.products.new') : product.data.value?.name_en))

watch(
  () => form.values.name_en,
  (name) => {
    if (!slugTouched.value) form.values.slug = slugify(name)
  },
)

const tabs = computed<TabItem<'en' | 'my'>[]>(() => [
  { value: 'en', label: t('admin.common.english') },
  { value: 'my', label: t('admin.common.myanmar') },
])

const categoryOptions = computed<SelectOption<number | null>[]>(() => [
  { value: null, label: t('admin.products.noCategory') },
  ...(categories.data.value ?? []).map((c) => ({
    value: c.id,
    label: c.is_active ? c.name_en : `${c.name_en} (${t('admin.common.inactive')})`,
  })),
])
const statusOptions = computed<SelectOption<ProductStatus>[]>(() => [
  { value: 'draft', label: t('admin.products.status.draft') },
  { value: 'active', label: t('admin.products.status.active') },
  { value: 'archived', label: t('admin.products.status.archived') },
])

function toggleEvent(id: number, on: boolean) {
  const set = new Set(form.values.event_ids)
  if (on) set.add(id)
  else set.delete(id)
  form.values.event_ids = [...set]
}

async function save() {
  formError.value = ''
  const data = form.validate()
  if (!data) {
    const errs = form.errors.value
    if (!errs.name_en && !errs.description_en && (errs.name_my || errs.description_my))
      langTab.value = 'my'
    else if (errs.name_en || errs.description_en) langTab.value = 'en'
    return
  }
  form.submitting.value = true
  try {
    if (props.id === undefined) {
      const created = await adminApi.createProduct(data)
      ui.success(t('admin.products.productCreated'))
      void router.replace({ name: 'admin-product-edit', params: { id: created.id } })
    } else {
      const updated = await adminApi.updateProduct(props.id, data)
      product.set(updated)
      fill(updated)
      ui.success(t('admin.products.productSaved'))
    }
  } catch (e) {
    if (!form.applyApiError(e)) formError.value = errorText(e)
    else formError.value = t('errors.validation')
  } finally {
    form.submitting.value = false
  }
}

const archiveOpen = ref(false)
const archiving = ref(false)
async function archive() {
  if (props.id === undefined) return
  archiving.value = true
  try {
    await adminApi.archiveProduct(props.id)
    ui.success(t('admin.products.archived'))
    archiveOpen.value = false
    void router.push({ name: 'admin-products' })
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    archiving.value = false
  }
}

const loading = computed(() => !isNew.value && !product.data.value && !product.error.value)
</script>

<template>
  <div>
    <AdminPageHeader
      :title="
        isNew ? $t('admin.products.new') : product.data.value?.name_en || $t('admin.products.edit')
      "
      :back="{ name: 'admin-products' }"
    >
      <BaseButton
        v-if="!isNew && product.data.value?.status === 'active'"
        variant="secondary"
        :href="`/products/${product.data.value.slug}`"
        target="_blank"
        rel="noopener"
      >
        <ExternalLink :size="16" aria-hidden="true" />{{ $t('admin.products.viewInStore') }}
      </BaseButton>
      <BaseButton
        v-if="!isNew && product.data.value && product.data.value.status !== 'archived'"
        variant="secondary"
        @click="archiveOpen = true"
      >
        <Archive :size="16" aria-hidden="true" />{{ $t('admin.products.archive') }}
      </BaseButton>
    </AdminPageHeader>

    <ErrorState
      v-if="product.error.value"
      :error="product.error.value"
      @retry="id !== undefined && product.run(id).then((p) => p && fill(p))"
    />

    <div v-else-if="loading" class="admin-cols">
      <div class="card card-pad"><SkeletonBlock :lines="8" height="22px" /></div>
      <div class="card card-pad"><SkeletonBlock :lines="4" height="22px" /></div>
    </div>

    <template v-else>
      <form class="admin-cols" novalidate @submit.prevent="save">
        <section class="card card-pad stack">
          <h2 class="admin-section-title">{{ $t('admin.products.basics') }}</h2>
          <BaseTabs v-model="langTab" :tabs="tabs" :label="$t('common.language')" />
          <div v-show="langTab === 'en'" class="stack">
            <BaseInput
              v-model="form.values.name_en"
              :label="$t('admin.common.nameEn')"
              :error="form.errors.value.name_en"
              required
              maxlength="200"
            />
            <BaseInput
              v-model="form.values.description_en"
              multiline
              :rows="8"
              :label="$t('admin.common.descriptionEn')"
              :error="form.errors.value.description_en"
              optional
            />
          </div>
          <div v-show="langTab === 'my'" class="stack" lang="my">
            <BaseInput
              v-model="form.values.name_my"
              :label="$t('admin.common.nameMy')"
              :error="form.errors.value.name_my"
              optional
              maxlength="200"
            />
            <BaseInput
              v-model="form.values.description_my"
              multiline
              :rows="8"
              :label="$t('admin.common.descriptionMy')"
              :error="form.errors.value.description_my"
              optional
            />
          </div>
          <BaseInput
            v-model="form.values.slug"
            :label="$t('admin.common.slug')"
            :hint="$t('admin.common.slugHint')"
            :error="form.errors.value.slug"
            required
            autocomplete="off"
            prefix="/products/"
            @input="slugTouched = true"
          />
        </section>

        <aside class="stack">
          <section class="card card-pad stack">
            <BaseSelect
              v-model="form.values.status"
              :label="$t('admin.products.statusLabel')"
              :options="statusOptions"
              :error="form.errors.value.status"
            />
            <BaseSelect
              v-model="form.values.category_id"
              :label="$t('admin.products.category')"
              :options="categoryOptions"
              :error="form.errors.value.category_id"
            />
            <BaseCheckbox
              v-model="form.values.featured"
              toggle
              :label="$t('admin.products.featured')"
              :hint="$t('admin.products.featuredHint')"
            />
          </section>

          <section class="card card-pad">
            <h2 class="admin-section-title">{{ $t('admin.products.events') }}</h2>
            <p class="muted text-sm">{{ $t('admin.products.eventsHint') }}</p>
            <SkeletonBlock v-if="events.loading.value" :lines="3" height="20px" />
            <p v-else-if="!events.data.value?.length" class="muted text-sm">
              {{ $t('admin.products.noEvents') }}
            </p>
            <div v-else class="events">
              <BaseCheckbox
                v-for="ev in events.data.value"
                :key="ev.id"
                :model-value="form.values.event_ids.includes(ev.id)"
                :label="ev.title_en"
                @update:model-value="toggleEvent(ev.id, $event)"
              />
            </div>
          </section>

          <div class="save-bar">
            <p v-if="formError" class="text-danger text-sm" role="alert">{{ formError }}</p>
            <BaseButton type="submit" block size="lg" :loading="form.submitting.value">
              <Save :size="18" aria-hidden="true" />{{ $t('common.save') }}
            </BaseButton>
          </div>
        </aside>
      </form>

      <section class="card card-pad block">
        <h2 class="admin-section-title">{{ $t('admin.products.images') }}</h2>
        <p v-if="id === undefined" class="muted text-sm">
          {{ $t('admin.products.imagesSaveFirst') }}
        </p>
        <ImageUploader v-else v-model="images" :product-id="id" />
      </section>

      <section class="card card-pad block">
        <h2 class="admin-section-title">{{ $t('admin.products.variants') }}</h2>
        <p v-if="id === undefined" class="muted text-sm">
          {{ $t('admin.products.variantsSaveFirst') }}
        </p>
        <VariantEditor v-else v-model="variants" :product-id="id" />
      </section>
    </template>

    <ConfirmDialog
      v-model:open="archiveOpen"
      :title="$t('admin.products.archiveTitle')"
      :message="$t('admin.products.archiveText')"
      :confirm-label="$t('admin.products.archive')"
      danger
      :loading="archiving"
      @confirm="archive"
    />
  </div>
</template>

<style scoped>
.events {
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
  max-height: 260px;
  overflow-y: auto;
}
.save-bar {
  position: sticky;
  bottom: var(--sp-3);
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
.save-bar p {
  margin: 0;
  background: var(--c-danger-bg);
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
}
.block {
  margin-top: var(--sp-5);
}
</style>
