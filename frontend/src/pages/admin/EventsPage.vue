<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import ProductPicker from '@/components/admin/ProductPicker.vue'
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
import { useAsync } from '@/composables/useAsync'
import { useForm, msg } from '@/composables/useForm'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import { fromLocalInput, toLocalInput } from '@/utils/date'
import { slugify } from '@/utils/misc'
import type { AdminEvent } from '@/types/api'
import type { Column } from '@/components/admin/types'

const { t } = useI18n()
const ui = useUiStore()
const { dateTime } = useLocalized()
const { errorText } = useErrorMessage()

const events = useAsync(() => adminApi.events())
onMounted(() => void events.run())
const rows = computed(() =>
  [...(events.data.value ?? [])].sort(
    (a, b) => a.sort_order - b.sort_order || b.starts_at.localeCompare(a.starts_at),
  ),
)

type Phase = 'running' | 'scheduled' | 'ended'
function phase(e: AdminEvent): Phase {
  const now = Date.now()
  if (new Date(e.ends_at).getTime() < now) return 'ended'
  if (new Date(e.starts_at).getTime() > now) return 'scheduled'
  return 'running'
}
const phaseTone = { running: 'success', scheduled: 'info', ended: 'neutral' } as const

const columns = computed<Column[]>(() => [
  { key: 'title_en', label: t('admin.nav.events'), primary: true },
  { key: 'period', label: t('admin.events.period') },
  { key: 'status', label: t('admin.common.status') },
  { key: 'products', label: t('admin.events.products'), align: 'right' },
  { key: 'actions', label: t('admin.common.actions'), align: 'right' },
])

function replaceRow(e: AdminEvent) {
  const list = events.data.value ?? []
  events.set(
    list.some((x) => x.id === e.id) ? list.map((x) => (x.id === e.id ? e : x)) : [...list, e],
  )
}

// ---------- Form ----------
const isDate = (v: string) => v !== '' && fromLocalInput(v) !== ''
const schema = z
  .object({
    title_en: requiredText(160),
    title_my: optionalText(160),
    slug: slugField(),
    description_en: optionalText(3000),
    description_my: optionalText(3000),
    starts_at: z.string().refine(isDate, msg('validation.required')),
    ends_at: z.string().refine(isDate, msg('validation.required')),
    sort_order: intField({ min: -100000, max: 100000 }),
    is_active: z.boolean(),
    product_ids: z.array(z.number()),
  })
  .refine(
    (v) =>
      !isDate(v.starts_at) ||
      !isDate(v.ends_at) ||
      fromLocalInput(v.ends_at) > fromLocalInput(v.starts_at),
    {
      path: ['ends_at'],
      message: msg('admin.validation.endAfterStart'),
    },
  )

function blank() {
  const start = new Date()
  start.setMinutes(0, 0, 0)
  const end = new Date(start.getTime() + 7 * 86400000)
  return {
    title_en: '',
    title_my: '',
    slug: '',
    description_en: '',
    description_my: '',
    starts_at: toLocalInput(start.toISOString()),
    ends_at: toLocalInput(end.toISOString()),
    sort_order: '0',
    is_active: true,
    product_ids: [] as number[],
  }
}
const form = useForm(schema, blank())
const open = ref(false)
const editing = ref<AdminEvent | null>(null)
const slugTouched = ref(false)
const formError = ref('')
const uploading = ref(false)

watch(
  () => form.values.title_en,
  (v) => {
    if (!slugTouched.value) form.values.slug = slugify(v)
  },
)

function openNew() {
  editing.value = null
  slugTouched.value = false
  formError.value = ''
  form.reset(blank())
  open.value = true
}

function openEdit(e: AdminEvent) {
  editing.value = e
  slugTouched.value = true
  formError.value = ''
  form.reset({
    title_en: e.title_en,
    title_my: e.title_my,
    slug: e.slug,
    description_en: e.description_en,
    description_my: e.description_my,
    starts_at: toLocalInput(e.starts_at),
    ends_at: toLocalInput(e.ends_at),
    sort_order: String(e.sort_order),
    is_active: e.is_active,
    product_ids: [...e.product_ids],
  })
  open.value = true
}

async function save() {
  const data = form.validate()
  if (!data) return
  const body = {
    ...data,
    starts_at: fromLocalInput(data.starts_at),
    ends_at: fromLocalInput(data.ends_at),
  }
  form.submitting.value = true
  formError.value = ''
  try {
    if (editing.value) {
      const e = await adminApi.updateEvent(editing.value.id, body)
      replaceRow(e)
      open.value = false
      ui.success(t('admin.common.saved'))
    } else {
      const e = await adminApi.createEvent(body)
      replaceRow(e)
      editing.value = e
      slugTouched.value = true
      ui.success(t('admin.common.created'))
    }
  } catch (e) {
    if (!form.applyApiError(e)) formError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}

async function uploadBanner(file: File) {
  if (!editing.value) return
  uploading.value = true
  try {
    const e = await adminApi.uploadEventBanner(editing.value.id, file)
    replaceRow(e)
    editing.value = e
    ui.success(t('admin.common.imageUploaded'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    uploading.value = false
  }
}

// ---------- Delete ----------
const toDelete = ref<AdminEvent | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)
async function remove() {
  const e = toDelete.value
  if (!e) return
  deleting.value = true
  try {
    await adminApi.deleteEvent(e.id)
    events.set((events.data.value ?? []).filter((x) => x.id !== e.id))
    confirmOpen.value = false
    ui.success(t('admin.common.deleted'))
  } catch (err) {
    ui.error(errorText(err))
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.events')">
      <BaseButton @click="openNew"
        ><Plus :size="18" aria-hidden="true" />{{ $t('admin.events.new') }}</BaseButton
      >
    </AdminPageHeader>

    <ErrorState v-if="events.error.value" :error="events.error.value" @retry="events.retry" />
    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :row-key="(e) => e.id"
      :loading="events.loading.value && !events.data.value"
      :caption="$t('admin.nav.events')"
    >
      <template #cell-title_en="{ row }">
        <span class="admin-thumb-cell">
          <ImagePlaceholder
            class="banner"
            :src="row.banner_url"
            :name="row.title_en"
            ratio="16 / 6"
            small
          />
          <span class="txt">
            <button type="button" class="linkish" @click="openEdit(row)">{{ row.title_en }}</button>
            <span class="sub admin-mono">{{ row.slug }}</span>
          </span>
        </span>
      </template>
      <template #cell-period="{ row }">
        <span class="period text-sm">
          <span>{{ dateTime(row.starts_at) }}</span>
          <span class="muted">→ {{ dateTime(row.ends_at) }}</span>
        </span>
      </template>
      <template #cell-status="{ row }">
        <span class="badges">
          <BaseBadge :tone="phaseTone[phase(row)]" size="sm" dot>{{
            $t(`admin.events.${phase(row)}`)
          }}</BaseBadge>
          <BaseBadge v-if="!row.is_active" size="sm">{{ $t('admin.common.inactive') }}</BaseBadge>
        </span>
      </template>
      <template #cell-products="{ row }">
        <span class="tabular">{{ row.product_ids.length }}</span>
      </template>
      <template #cell-actions="{ row }">
        <span class="admin-row-actions">
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.edit')} ${row.title_en}`"
            @click="openEdit(row)"
          >
            <Pencil :size="16" aria-hidden="true" />
          </BaseButton>
          <BaseButton
            size="sm"
            variant="ghost"
            :aria-label="`${$t('admin.common.delete')} ${row.title_en}`"
            @click="((toDelete = row), (confirmOpen = true))"
          >
            <Trash2 :size="16" aria-hidden="true" />
          </BaseButton>
        </span>
      </template>
      <template #empty>
        <EmptyState :title="$t('admin.events.empty')">
          <BaseButton @click="openNew">{{ $t('admin.events.new') }}</BaseButton>
        </EmptyState>
      </template>
    </DataTable>

    <BaseModal
      v-model:open="open"
      size="xl"
      :title="editing ? $t('admin.events.edit') : $t('admin.events.new')"
    >
      <form id="event-form" class="grid-form cols-2" novalidate @submit.prevent="save">
        <BaseInput
          v-model="form.values.title_en"
          :label="$t('admin.common.titleEn')"
          :error="form.errors.value.title_en"
          required
        />
        <BaseInput
          v-model="form.values.title_my"
          lang="my"
          :label="$t('admin.common.titleMy')"
          :error="form.errors.value.title_my"
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
        <BaseInput
          v-model="form.values.description_en"
          multiline
          :rows="3"
          :label="$t('admin.common.descriptionEn')"
          :error="form.errors.value.description_en"
          optional
        />
        <BaseInput
          v-model="form.values.description_my"
          multiline
          :rows="3"
          lang="my"
          :label="$t('admin.common.descriptionMy')"
          :error="form.errors.value.description_my"
          optional
        />
        <BaseInput
          v-model="form.values.starts_at"
          type="datetime-local"
          :label="$t('admin.events.startsAt')"
          :error="form.errors.value.starts_at"
          required
        />
        <BaseInput
          v-model="form.values.ends_at"
          type="datetime-local"
          :label="$t('admin.events.endsAt')"
          :hint="$t('admin.events.timezoneHint')"
          :error="form.errors.value.ends_at"
          required
        />
        <div class="span-2">
          <BaseCheckbox v-model="form.values.is_active" toggle :label="$t('admin.common.active')" />
        </div>
        <div class="span-2">
          <p class="field-label">{{ $t('admin.events.products') }}</p>
          <ProductPicker v-model="form.values.product_ids" />
        </div>
        <div class="span-2">
          <p class="field-label">{{ $t('admin.events.banner') }}</p>
          <div v-if="editing" class="banner-row">
            <ImagePlaceholder :src="editing.banner_url" :name="editing.title_en" ratio="16 / 6" />
            <FileDrop
              :label="
                editing.banner_url
                  ? $t('admin.common.replaceImage')
                  : $t('admin.common.uploadImage')
              "
              :hint="$t('admin.events.bannerHint')"
              :preview="false"
              :disabled="uploading"
              @select="uploadBanner"
            />
          </div>
          <p v-else class="muted text-sm">{{ $t('admin.events.bannerAfterSave') }}</p>
        </div>
        <p v-if="formError" class="span-2 text-danger text-sm" role="alert">{{ formError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="open = false">{{ $t('common.close') }}</BaseButton>
        <BaseButton type="submit" form="event-form" :loading="form.submitting.value">{{
          $t('common.save')
        }}</BaseButton>
      </template>
    </BaseModal>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('admin.common.confirmDelete')"
      :message="toDelete ? `${toDelete.title_en} — ${$t('admin.common.confirmDeleteText')}` : ''"
      :confirm-label="$t('admin.common.delete')"
      danger
      :loading="deleting"
      @confirm="remove"
    />
  </div>
</template>

<style scoped>
.banner {
  width: 96px !important;
  flex: none;
}
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
.period {
  display: inline-flex;
  flex-direction: column;
  white-space: nowrap;
}
.badges {
  display: inline-flex;
  gap: 4px;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.field-label {
  margin: 0 0 var(--sp-2);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
}
.banner-row {
  display: grid;
  gap: var(--sp-3);
}
@media (min-width: 640px) {
  .banner-row {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}
</style>
