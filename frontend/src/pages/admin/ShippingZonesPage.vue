<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { MapPin, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import MoneyInput from '@/components/admin/MoneyInput.vue'
import ProvincePicker from '@/components/admin/ProvincePicker.vue'
import {
  intField,
  optionalSatangField,
  optionalText,
  requiredText,
  satangField,
} from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import { adminApi } from '@/api/admin'
import { catalogApi } from '@/api/catalog'
import { isApiError } from '@/api/client'
import { useAsync } from '@/composables/useAsync'
import { useForm } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import { formatBaht } from '@/utils/money'
import type { ShippingZone } from '@/types/api'

const { t } = useI18n()
const ui = useUiStore()
const { errorText } = useErrorMessage()

const zones = useAsync(() => adminApi.zones())
const provinces = useAsync(() => catalogApi.provinces())
onMounted(() => {
  void zones.run()
  void provinces.run()
})

const rows = computed(() =>
  [...(zones.data.value ?? [])].sort(
    (a, b) =>
      Number(b.is_default) - Number(a.is_default) || a.sort_order - b.sort_order || a.id - b.id,
  ),
)

const assigned = computed(() => new Set((zones.data.value ?? []).flatMap((zn) => zn.provinces)))
const unassigned = computed(() =>
  (provinces.data.value ?? []).filter((p) => !assigned.value.has(p.name_en)).map((p) => p.name_en),
)

function replaceRow(zn: ShippingZone) {
  let list = zones.data.value ?? []
  // Only one default zone: the backend switches it, mirror that locally.
  if (zn.is_default) list = list.map((x) => (x.id === zn.id ? x : { ...x, is_default: false }))
  zones.set(
    list.some((x) => x.id === zn.id) ? list.map((x) => (x.id === zn.id ? zn : x)) : [...list, zn],
  )
}

// ---------- Form ----------
const schema = z.object({
  name_en: requiredText(120),
  name_my: optionalText(120),
  fee_satang: satangField(),
  free_over_satang: optionalSatangField(),
  is_default: z.boolean(),
  sort_order: intField({ min: -100000, max: 100000 }),
  provinces: z.array(z.string()),
})
const blank = {
  name_en: '',
  name_my: '',
  fee_satang: NaN,
  free_over_satang: null as number | null,
  is_default: false,
  sort_order: '0',
  provinces: [] as string[],
}
const form = useForm(schema, { ...blank })
const open = ref(false)
const editing = ref<ShippingZone | null>(null)
const formError = ref('')

const takenBy = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const zn of zones.data.value ?? []) {
    if (zn.id === editing.value?.id) continue
    for (const p of zn.provinces) map[p] = zn.name_en
  }
  return map
})

function openNew() {
  editing.value = null
  formError.value = ''
  form.reset({ ...blank, provinces: [], sort_order: String(rows.value.length) })
  open.value = true
}
function openEdit(zn: ShippingZone) {
  editing.value = zn
  formError.value = ''
  form.reset({
    name_en: zn.name_en,
    name_my: zn.name_my,
    fee_satang: zn.fee_satang,
    free_over_satang: zn.free_over_satang,
    is_default: zn.is_default,
    sort_order: String(zn.sort_order),
    provinces: [...zn.provinces],
  })
  open.value = true
}

async function save() {
  const d = form.validate()
  if (!d) return
  form.submitting.value = true
  formError.value = ''
  try {
    const zn = editing.value
      ? await adminApi.updateZone(editing.value.id, d)
      : await adminApi.createZone(d)
    replaceRow(zn)
    open.value = false
    ui.success(editing.value ? t('admin.common.saved') : t('admin.common.created'))
  } catch (e) {
    if (isApiError(e) && e.code === 'province_taken') {
      form.errors.value = { ...form.errors.value, provinces: t('errors.province_taken') }
      void zones.run()
    } else if (!form.applyApiError(e)) formError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}

// ---------- Delete ----------
const toDelete = ref<ShippingZone | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)
async function remove() {
  const zn = toDelete.value
  if (!zn) return
  deleting.value = true
  try {
    await adminApi.deleteZone(zn.id)
    zones.set((zones.data.value ?? []).filter((x) => x.id !== zn.id))
    confirmOpen.value = false
    ui.success(t('admin.common.deleted'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.zones')">
      <BaseButton @click="openNew"
        ><Plus :size="18" aria-hidden="true" />{{ $t('admin.zones.new') }}</BaseButton
      >
    </AdminPageHeader>

    <div
      v-if="provinces.data.value && zones.data.value"
      class="unassigned"
      :class="{ ok: unassigned.length === 0 }"
      role="status"
    >
      <MapPin :size="18" aria-hidden="true" />
      <div>
        <strong>{{
          $t('admin.zones.unassigned', { n: unassigned.length }, unassigned.length)
        }}</strong>
        <template v-if="unassigned.length">
          <span class="muted"> — {{ $t('admin.zones.unassignedHint') }}</span>
          <p class="list">{{ unassigned.join(', ') }}</p>
        </template>
      </div>
    </div>

    <ErrorState v-if="zones.error.value" :error="zones.error.value" @retry="zones.retry" />
    <div v-else-if="zones.loading.value && !zones.data.value" class="grid">
      <div v-for="i in 3" :key="i" class="card card-pad">
        <SkeletonBlock :lines="4" height="16px" />
      </div>
    </div>
    <EmptyState v-else-if="rows.length === 0" :title="$t('admin.zones.empty')">
      <BaseButton @click="openNew">{{ $t('admin.zones.new') }}</BaseButton>
    </EmptyState>
    <ul v-else class="grid">
      <li v-for="zone in rows" :key="zone.id" class="card card-pad zone">
        <div class="zhead">
          <div>
            <h2 class="zname">{{ zone.name_en }}</h2>
            <p v-if="zone.name_my" class="sub" lang="my">{{ zone.name_my }}</p>
          </div>
          <BaseBadge v-if="zone.is_default" tone="primary" size="sm">{{
            $t('admin.zones.default')
          }}</BaseBadge>
        </div>
        <p class="fee">
          <strong>{{
            zone.fee_satang === 0 ? $t('admin.zones.free') : formatBaht(zone.fee_satang)
          }}</strong>
          <span v-if="zone.free_over_satang !== null" class="muted text-sm">
            · {{ $t('admin.zones.freeOverShort', { amount: formatBaht(zone.free_over_satang) }) }}
          </span>
        </p>
        <p class="sub">
          {{ $t('admin.zones.provincesN', { n: zone.provinces.length }, zone.provinces.length) }}
        </p>
        <ul v-if="zone.provinces.length" class="chips">
          <li v-for="p in zone.provinces.slice(0, 12)" :key="p">{{ p }}</li>
          <li v-if="zone.provinces.length > 12" class="more">+{{ zone.provinces.length - 12 }}</li>
        </ul>
        <div class="actions">
          <BaseButton size="sm" variant="secondary" @click="openEdit(zone)">
            <Pencil :size="15" aria-hidden="true" />{{ $t('admin.common.edit') }}
          </BaseButton>
          <BaseButton
            size="sm"
            variant="ghost"
            :disabled="zone.is_default"
            :title="zone.is_default ? $t('admin.zones.cannotDeleteDefault') : undefined"
            :aria-label="`${$t('admin.common.delete')} ${zone.name_en}`"
            @click="((toDelete = zone), (confirmOpen = true))"
          >
            <Trash2 :size="15" aria-hidden="true" />
          </BaseButton>
        </div>
      </li>
    </ul>

    <BaseModal
      v-model:open="open"
      size="xl"
      :title="editing ? $t('admin.zones.edit') : $t('admin.zones.new')"
    >
      <form id="zone-form" class="grid-form cols-2" novalidate @submit.prevent="save">
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
        <MoneyInput
          v-model="form.values.fee_satang"
          :label="$t('admin.zones.fee')"
          :error="form.errors.value.fee_satang"
          required
        />
        <MoneyInput
          v-model="form.values.free_over_satang"
          nullable
          :label="$t('admin.zones.freeOver')"
          :hint="$t('admin.zones.freeOverHint')"
          :error="form.errors.value.free_over_satang"
          optional
        />
        <BaseInput
          v-model="form.values.sort_order"
          type="number"
          inputmode="numeric"
          :label="$t('admin.common.sortOrder')"
          :error="form.errors.value.sort_order"
        />
        <div class="default-box">
          <BaseCheckbox
            v-model="form.values.is_default"
            toggle
            :label="$t('admin.zones.default')"
            :hint="$t('admin.zones.defaultHint')"
            :disabled="editing?.is_default"
          />
        </div>
        <div class="span-2">
          <p class="field-label">{{ $t('admin.zones.provinces') }}</p>
          <ProvincePicker
            v-model="form.values.provinces"
            :taken-by="takenBy"
            :error="form.errors.value.provinces"
          />
        </div>
        <p v-if="formError" class="span-2 text-danger text-sm" role="alert">{{ formError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="open = false">{{ $t('common.cancel') }}</BaseButton>
        <BaseButton type="submit" form="zone-form" :loading="form.submitting.value">{{
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
.unassigned {
  display: flex;
  gap: var(--sp-3);
  align-items: flex-start;
  padding: var(--sp-3) var(--sp-4);
  border-radius: var(--radius-md);
  background: var(--c-warning-bg);
  color: var(--c-text);
  margin-bottom: var(--sp-4);
  font-size: var(--fs-sm);
}
.unassigned svg {
  flex: none;
  margin-top: 3px;
  color: var(--c-warning);
}
.unassigned.ok {
  background: var(--c-success-bg);
}
.unassigned.ok svg {
  color: var(--c-success);
}
.list {
  margin: var(--sp-1) 0 0;
  color: var(--c-text-2);
}
.grid {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--sp-4);
}
@media (min-width: 768px) {
  .grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
.zone {
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
.zhead {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--sp-2);
}
.zname {
  font-size: var(--fs-lg);
  margin: 0;
}
.sub {
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.fee {
  margin: 0;
}
.chips {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.chips li {
  font-size: 11px;
  background: var(--c-surface-2);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-full);
  padding: 1px 8px;
}
.chips .more {
  background: var(--c-primary-50);
  color: var(--c-primary);
}
.actions {
  margin-top: auto;
  padding-top: var(--sp-2);
  display: flex;
  gap: var(--sp-2);
}
.default-box {
  align-self: end;
  padding-bottom: var(--sp-2);
}
.field-label {
  margin: 0 0 var(--sp-2);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
  color: var(--c-text-2);
}
</style>
