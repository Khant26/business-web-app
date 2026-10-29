<script setup lang="ts">
/**
 * Thai address form with a province → district → subdistrict cascade.
 * The postcode is filled from the chosen subdistrict. Province values are the
 * dataset's English names (the backend's canonical shipping-zone keys).
 * Falls back to a province list from the API + free-text fields if the
 * address dataset can't be loaded.
 */
import { computed, onMounted, ref } from 'vue'
import { Info } from 'lucide-vue-next'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import { loadThaiAddress, type ThaiProvince } from '@/utils/thaiAddress'
import { catalogApi } from '@/api/catalog'
import type { Province } from '@/types/api'
import type { SelectOption } from '@/types/ui'
import type { AddressFields } from './checkout/address'

const model = defineModel<AddressFields>({ required: true })
withDefaults(
  defineProps<{ errors?: Record<string, string>; showLabel?: boolean; showDefault?: boolean }>(),
  { errors: () => ({}) },
)

const data = ref<ThaiProvince[] | null>(null)
const fallbackProvinces = ref<Province[]>([])
const loading = ref(true)
const fallback = ref(false)

onMounted(async () => {
  try {
    data.value = await loadThaiAddress()
  } catch {
    fallback.value = true
    try {
      fallbackProvinces.value = await catalogApi.provinces()
    } catch {
      /* province becomes free text below */
    }
  } finally {
    loading.value = false
  }
})

function patch(p: Partial<AddressFields>) {
  model.value = { ...model.value, ...p }
}

const both = (th: string, en: string) => (th && th !== en ? `${th} · ${en}` : en)

const provinceOptions = computed<SelectOption<string>[]>(() => {
  if (data.value) return data.value.map((p) => ({ value: p.p_en, label: both(p.p_th, p.p_en) }))
  return fallbackProvinces.value.map((p) => ({
    value: p.name_en,
    label: both(p.name_th, p.name_en),
  }))
})

const province = computed(() => data.value?.find((p) => p.p_en === model.value.province) ?? null)
const district = computed(
  () => province.value?.d.find((d) => d.en === model.value.district) ?? null,
)

const districtOptions = computed<SelectOption<string>[]>(
  () => province.value?.d.map((d) => ({ value: d.en, label: both(d.th, d.en) })) ?? [],
)
const subdistrictOptions = computed<SelectOption<string>[]>(
  () =>
    district.value?.s.map((s) => ({
      value: s.en,
      label: `${both(s.th, s.en)} (${s.zip})`,
    })) ?? [],
)

const provinceModel = computed({
  get: () => model.value.province,
  set: (v: string) => {
    if (v === model.value.province) return
    patch(
      fallback.value
        ? { province: v }
        : { province: v, district: '', subdistrict: '', postcode: '' },
    )
  },
})
const districtModel = computed({
  get: () => model.value.district,
  set: (v: string) => {
    if (v === model.value.district) return
    patch({ district: v, subdistrict: '', postcode: '' })
  },
})
const subdistrictModel = computed({
  get: () => model.value.subdistrict,
  set: (v: string) => {
    const zip = district.value?.s.find((s) => s.en === v)?.zip ?? ''
    patch({ subdistrict: v, postcode: zip })
  },
})

function field<K extends keyof AddressFields>(k: K) {
  return computed({
    get: () => model.value[k],
    set: (v: AddressFields[K]) => patch({ [k]: v } as Partial<AddressFields>),
  })
}
const label = field('label')
const recipient = field('recipient_name')
const phone = field('phone')
const line1 = field('line1')
const districtText = field('district')
const subdistrictText = field('subdistrict')
const postcode = field('postcode')
const isDefault = field('is_default')

const cascade = computed(() => !fallback.value && !!data.value)
</script>

<template>
  <div class="addr grid-form cols-2">
    <BaseInput
      v-if="showLabel"
      v-model="label"
      class="span-2"
      :label="$t('address.label')"
      :placeholder="$t('address.labelPlaceholder')"
      :error="errors.label"
      optional
      maxlength="40"
    />
    <BaseInput
      v-model="recipient"
      :label="$t('address.recipient')"
      :error="errors.recipient_name"
      autocomplete="name"
      required
      maxlength="100"
    />
    <BaseInput
      v-model="phone"
      :label="$t('address.phone')"
      :placeholder="$t('address.phonePlaceholder')"
      :error="errors.phone"
      type="tel"
      inputmode="tel"
      autocomplete="tel"
      required
      maxlength="20"
    />
    <BaseInput
      v-model="line1"
      class="span-2"
      :label="$t('address.line1')"
      :placeholder="$t('address.line1Placeholder')"
      :error="errors.line1"
      autocomplete="address-line1"
      required
      maxlength="200"
    />

    <p v-if="loading" class="span-2 note muted" role="status">{{ $t('address.loadingData') }}</p>
    <p v-else-if="fallback" class="span-2 note warn" role="status">
      <Info :size="16" aria-hidden="true" />{{ $t('address.fallbackNotice') }}
    </p>

    <BaseSelect
      v-if="provinceOptions.length"
      v-model="provinceModel"
      :options="provinceOptions"
      :label="$t('address.province')"
      :placeholder="$t('address.selectProvince')"
      :error="errors.province"
      :disabled="loading"
      autocomplete="address-level1"
      required
    />
    <BaseInput
      v-else
      v-model="provinceModel"
      :label="$t('address.province')"
      :error="errors.province"
      :disabled="loading"
      autocomplete="address-level1"
      required
    />

    <template v-if="cascade">
      <BaseSelect
        v-model="districtModel"
        :options="districtOptions"
        :label="$t('address.district')"
        :placeholder="$t('address.selectDistrict')"
        :error="errors.district"
        :disabled="!province"
        required
      />
      <BaseSelect
        v-model="subdistrictModel"
        :options="subdistrictOptions"
        :label="$t('address.subdistrict')"
        :placeholder="$t('address.selectSubdistrict')"
        :error="errors.subdistrict"
        :disabled="!district"
        required
      />
      <BaseInput
        v-model="postcode"
        :label="$t('address.postcode')"
        :error="errors.postcode"
        :hint="$t('address.postcodeAuto')"
        inputmode="numeric"
        autocomplete="postal-code"
        maxlength="5"
        required
        readonly
      />
    </template>
    <template v-else>
      <BaseInput
        v-model="districtText"
        :label="$t('address.district')"
        :error="errors.district"
        :disabled="loading"
        autocomplete="address-level2"
        required
        maxlength="100"
      />
      <BaseInput
        v-model="subdistrictText"
        :label="$t('address.subdistrict')"
        :error="errors.subdistrict"
        :disabled="loading"
        autocomplete="address-level3"
        required
        maxlength="100"
      />
      <BaseInput
        v-model="postcode"
        :label="$t('address.postcode')"
        :error="errors.postcode"
        :disabled="loading"
        inputmode="numeric"
        autocomplete="postal-code"
        maxlength="5"
        required
      />
    </template>

    <BaseCheckbox
      v-if="showDefault"
      v-model="isDefault"
      class="span-2"
      :label="$t('address.setDefault')"
    />
  </div>
</template>

<style scoped>
.note {
  margin: 0;
  font-size: var(--fs-sm);
  display: flex;
  align-items: flex-start;
  gap: var(--sp-2);
}
.warn {
  padding: var(--sp-2) var(--sp-3);
  border-radius: var(--radius-md);
  background: var(--c-warning-bg);
  color: var(--c-warning);
}
.warn :deep(svg) {
  flex: none;
  margin-top: 3px;
}
</style>
