<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import { intField, optionalText, requiredText, slugField } from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import FileDrop from '@/components/ui/FileDrop.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import { adminApi } from '@/api/admin'
import { isApiError } from '@/api/client'
import { useAsync } from '@/composables/useAsync'
import { useForm } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import { slugify } from '@/utils/misc'
import type { Category } from '@/types/api'
import type { Column } from '@/components/admin/types'

const { t } = useI18n()
const ui = useUiStore()
const { errorText } = useErrorMessage()

const cats = useAsync(() => adminApi.categories())
onMounted(() => void cats.run())
const rows = computed(() =>
  [...(cats.data.value ?? [])].sort((a, b) => a.sort_order - b.sort_order || a.id - b.id),
)

const columns = computed<Column[]>(() => [
  { key: 'name_en', label: t('admin.common.name'), primary: true },
  { key: 'slug', label: t('admin.common.slug') },
  { key: 'product_count', label: t('admin.nav.products'), align: 'right' },
  { key: 'sort_order', label: t('admin.common.sortOrder'), align: 'right' },
  { key: 'is_active', label: t('admin.common.status') },
  { key: 'actions', label: t('admin.common.actions'), align: 'right' },
])

function replaceRow(c: Category) {
  const list = cats.data.value ?? []
  cats.set(
    list.some((x) => x.id === c.id) ? list.map((x) => (x.id === c.id ? c : x)) : [...list, c],
  )
}

// ---------- Form ----------
const schema = z.object({
  name_en: requiredText(120),
  name_my: optionalText(120),
  slug: slugField(),
  sort_order: intField({ min: -100000, max: 100000 }),
  is_active: z.boolean(),
})
const blank = { name_en: '', name_my: '', slug: '', sort_order: '0', is_active: true }
const form = useForm(schema, { ...blank })
const open = ref(false)
const editing = ref<Category | null>(null)
const slugTouched = ref(false)
const formError = ref('')
const uploading = ref(false)

watch(
  () => form.values.name_en,
  (v) => {
    if (!slugTouched.value) form.values.slug = slugify(v)
  },
)

function openNew() {
  editing.value = null
  slugTouched.value = false
  formError.value = ''
  const next = rows.value.reduce((m, c) => Math.max(m, c.sort_order), -1) + 1
  form.reset({ ...blank, sort_order: String(next) })
  open.value = true
}

function openEdit(c: Category) {
  editing.value = c
  slugTouched.value = true
  formError.value = ''
  form.reset({
    name_en: c.name_en,
    name_my: c.name_my,
    slug: c.slug,
    sort_order: String(c.sort_order),
    is_active: c.is_active,
  })
  open.value = true
}

async function save() {
  const data = form.validate()
  if (!data) return
  form.submitting.value = true
  formError.value = ''
  try {
    if (editing.value) {
      const c = await adminApi.updateCategory(editing.value.id, data)
      replaceRow(c)
      editing.value = c
      open.value = false
      ui.success(t('admin.common.saved'))
    } else {
      const c = await adminApi.createCategory(data)
      replaceRow(c)
      // Keep the dialog open in edit mode so an image can be added right away.
      editing.value = c
      slugTouched.value = true
      ui.success(t('admin.common.created'))
    }
  } catch (e) {
    if (!form.applyApiError(e)) formError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}

async function uploadImage(file: File) {
  if (!editing.value) return
  uploading.value = true
  try {
    const c = await adminApi.uploadCategoryImage(editing.value.id, file)
    replaceRow(c)
    editing.value = c
    ui.success(t('admin.common.imageUploaded'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    uploading.value = false
  }
}

// ---------- Delete ----------
const toDelete = ref<Category | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)
function askDelete(c: Category) {
  toDelete.value = c
  confirmOpen.value = true
}
async function remove() {
  const c = toDelete.value
  if (!c) return
  deleting.value = true
  try {
    await adminApi.deleteCategory(c.id)
    cats.set((cats.data.value ?? []).filter((x) => x.id !== c.id))
    confirmOpen.value = false
    ui.success(t('admin.common.deleted'))
  } catch (e) {
    ui.error(isApiError(e) && e.code === 'in_use' ? t('admin.categories.inUse') : errorText(e))
    confirmOpen.value = false
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.categories')">
      <BaseButton @click="openNew"
        ><Plus :size="18" aria-hidden="true" />{{ $t('admin.categories.new') }}</BaseButton
      >
    </AdminPageHeader>

    <ErrorState v-if="cats.error.value" :error="cats.error.value" @retry="cats.retry" />
    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :row-key="(c) => c.id"
      :loading="cats.loading.value && !cats.data.value"
      :caption="$t('admin.nav.categories')"
    >
      <template #cell-name_en="{ row }">
        <span class="admin-thumb-cell">
          <ImagePlaceholder class="thumb" :src="row.image_url" :name="row.name_en" small />
          <span class="txt">
            <button type="button" class="linkish" @click="openEdit(row)">{{ row.name_en }}</button>
            <span v-if="row.name_my" class="sub" lang="my">{{ row.name_my }}</span>
          </span>
        </span>
      </template>
      <template #cell-slug="{ row }"
        ><span class="admin-mono">{{ row.slug }}</span></template
      >
      <template #cell-product_count="{ row }"
        ><span class="tabular">{{ row.product_count }}</span></template
      >
      <template #cell-sort_order="{ row }"
        ><span class="tabular">{{ row.sort_order }}</span></template
      >
      <template #cell-is_active="{ row }">
        <BaseBadge :tone="row.is_active ? 'success' : 'neutral'" size="sm">
          {{ row.is_active ? $t('admin.common.active') : $t('admin.common.inactive') }}
        </BaseBadge>
      </template>
      <template #cell-actions="{ row }">
        <span class="admin-row-actions">
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.edit')} ${row.name_en}`"
            @click="openEdit(row)"
          >
            <Pencil :size="16" aria-hidden="true" />
          </BaseButton>
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.delete')} ${row.name_en}`"
            @click="askDelete(row)"
          >
            <Trash2 :size="16" aria-hidden="true" />
          </BaseButton>
        </span>
      </template>
      <template #empty>
        <EmptyState :title="$t('admin.categories.empty')">
          <BaseButton @click="openNew">{{ $t('admin.categories.new') }}</BaseButton>
        </EmptyState>
      </template>
    </DataTable>

    <BaseModal
      v-model:open="open"
      size="lg"
      :title="editing ? $t('admin.categories.edit') : $t('admin.categories.new')"
    >
      <form id="cat-form" class="grid-form cols-2" novalidate @submit.prevent="save">
        <BaseInput
          v-model="form.values.name_en"
          :label="$t('admin.common.nameEn')"
          :error="form.errors.value.name_en"
          required
        />
        <BaseInput
          v-model="form.values.name_my"
          lang="my"
          :label="$t('admin.common.nameMy')"
          :error="form.errors.value.name_my"
          optional
        />
        <BaseInput
          v-model="form.values.slug"
          :label="$t('admin.common.slug')"
          :hint="$t('admin.common.slugHint')"
          :error="form.errors.value.slug"
          required
          autocomplete="off"
          @input="slugTouched = true"
        />
        <BaseInput
          v-model="form.values.sort_order"
          type="number"
          inputmode="numeric"
          :label="$t('admin.common.sortOrder')"
          :error="form.errors.value.sort_order"
        />
        <div class="span-2">
          <BaseCheckbox v-model="form.values.is_active" toggle :label="$t('admin.common.active')" />
        </div>
        <div class="span-2 image-row">
          <p class="field-label">{{ $t('admin.common.image') }}</p>
          <template v-if="editing">
            <div class="img-wrap">
              <ImagePlaceholder
                class="cur"
                :src="editing.image_url"
                :name="editing.name_en"
                ratio="4 / 3"
              />
              <FileDrop
                :label="
                  editing.image_url
                    ? $t('admin.common.replaceImage')
                    : $t('admin.common.uploadImage')
                "
                :hint="$t('admin.common.imageHint')"
                :preview="false"
                :disabled="uploading"
                @select="uploadImage"
              />
            </div>
          </template>
          <p v-else class="muted text-sm">{{ $t('admin.categories.imageAfterSave') }}</p>
        </div>
        <p v-if="formError" class="span-2 text-danger text-sm" role="alert">{{ formError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="open = false">{{ $t('common.close') }}</BaseButton>
        <BaseButton type="submit" form="cat-form" :loading="form.submitting.value">{{
          $t('common.save')
        }}</BaseButton>
      </template>
    </BaseModal>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('admin.common.confirmDelete')"
      :message="toDelete ? `${toDelete.name_en} — ${$t('admin.common.confirmDeleteText')}` : ''"
      :confirm-label="$t('admin.common.delete')"
      danger
      :loading="deleting"
      @confirm="remove"
    />
  </div>
</template>

<style scoped>
.linkish {
  border: 0;
  background: none;
  padding: 0;
  color: var(--c-primary);
  font-weight: var(--fw-semibold);
  cursor: pointer;
  text-align: left;
  font-size: var(--fs-sm);
}
.linkish:hover {
  text-decoration: underline;
}
.field-label {
  margin: 0 0 var(--sp-1);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
}
.img-wrap {
  display: grid;
  gap: var(--sp-3);
  grid-template-columns: 1fr;
}
@media (min-width: 560px) {
  .img-wrap {
    grid-template-columns: 180px 1fr;
    align-items: center;
  }
}
.cur {
  border: 1px solid var(--c-border);
}
</style>
