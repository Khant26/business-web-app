<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Save } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import MoneyInput from '@/components/admin/MoneyInput.vue'
import { intField, optionalText, requiredText, satangField } from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import { adminApi } from '@/api/admin'
import { useAsync } from '@/composables/useAsync'
import { useForm, msg } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useUiStore } from '@/stores/ui'
import { useSettingsStore } from '@/stores/settings'
import { isThaiPhone } from '@/utils/misc'
import type { AdminSettings } from '@/types/api'

const { t } = useI18n()
const ui = useUiStore()
const publicSettings = useSettingsStore()
const { errorText } = useErrorMessage()

const settings = useAsync(() => adminApi.settings())

const digits = (s: string) => s.replace(/[\s-]/g, '')

const schema = z.object({
  shop_name: requiredText(120),
  promptpay_id: z
    .string()
    .trim()
    .transform(digits)
    .refine(
      (v) => v === '' || /^(0\d{9}|66\d{9}|\d{13}|\d{15})$/.test(v),
      msg('admin.settings.promptpayInvalid'),
    ),
  promptpay_name: optionalText(120),
  vat_percent: z
    .union([z.string(), z.number()])
    .refine((v) => String(v).trim() !== '', msg('validation.required'))
    .transform((v) => Number(v))
    .pipe(
      z
        .number({ message: msg('validation.number') })
        .min(0, msg('validation.min', { min: 0 }))
        .max(30, msg('validation.max', { max: 30 })),
    ),
  payment_window_minutes: intField({ min: 5, max: 1440 }),
  cod_max_satang: satangField(),
  support_email: z
    .string()
    .trim()
    .refine((v) => v === '' || z.string().email().safeParse(v).success, msg('validation.email')),
  support_phone: z
    .string()
    .trim()
    .refine((v) => v === '' || isThaiPhone(v), msg('validation.phone')),
})

const form = useForm(schema, {
  shop_name: '',
  promptpay_id: '',
  promptpay_name: '',
  vat_percent: '7',
  payment_window_minutes: '30',
  cod_max_satang: 0,
  support_email: '',
  support_phone: '',
})
const formError = ref('')

function fill(s: AdminSettings) {
  form.reset({
    shop_name: s.shop_name,
    promptpay_id: s.promptpay_id,
    promptpay_name: s.promptpay_name,
    vat_percent: String(s.vat_rate_bp / 100),
    payment_window_minutes: String(s.payment_window_minutes),
    cod_max_satang: s.cod_max_satang,
    support_email: s.support_email,
    support_phone: s.support_phone,
  })
}

async function load() {
  const s = await settings.run()
  if (s) fill(s)
}
onMounted(load)

async function save() {
  const d = form.validate()
  if (!d) return
  const body: AdminSettings = {
    shop_name: d.shop_name,
    promptpay_id: d.promptpay_id,
    promptpay_name: d.promptpay_name,
    vat_rate_bp: Math.round(d.vat_percent * 100),
    payment_window_minutes: d.payment_window_minutes,
    cod_max_satang: d.cod_max_satang,
    support_email: d.support_email,
    support_phone: d.support_phone,
  }
  form.submitting.value = true
  formError.value = ''
  try {
    const s = await adminApi.updateSettings(body)
    settings.set(s)
    fill(s)
    void publicSettings.load(true)
    ui.success(t('admin.settings.saved'))
  } catch (e) {
    if (form.applyApiError(e)) {
      const errs = { ...form.errors.value }
      if (errs.vat_rate_bp) errs.vat_percent = errs.vat_rate_bp
      form.errors.value = errs
    } else formError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}
</script>

<template>
  <div class="settings">
    <AdminPageHeader :title="$t('admin.nav.settings')" />

    <ErrorState v-if="settings.error.value" :error="settings.error.value" @retry="load" />
    <div v-else-if="!settings.data.value" class="stack">
      <div v-for="i in 3" :key="i" class="card card-pad">
        <SkeletonBlock :lines="3" height="40px" />
      </div>
    </div>

    <form v-else class="stack" novalidate @submit.prevent="save">
      <section class="card card-pad">
        <h2 class="admin-section-title">{{ $t('admin.settings.shop') }}</h2>
        <div class="grid-form cols-2">
          <BaseInput
            v-model="form.values.shop_name"
            class="span-2"
            :label="$t('admin.settings.shopName')"
            :error="form.errors.value.shop_name"
            required
          />
        </div>
      </section>

      <section class="card card-pad">
        <h2 class="admin-section-title">{{ $t('admin.settings.payment') }}</h2>
        <div class="grid-form cols-2">
          <BaseInput
            v-model="form.values.promptpay_id"
            :label="$t('admin.settings.promptpayId')"
            :hint="$t('admin.settings.promptpayIdHint')"
            :error="form.errors.value.promptpay_id"
            inputmode="numeric"
            autocomplete="off"
          />
          <BaseInput
            v-model="form.values.promptpay_name"
            :label="$t('admin.settings.promptpayName')"
            :hint="$t('admin.settings.promptpayNameHint')"
            :error="form.errors.value.promptpay_name"
          />
          <MoneyInput
            v-model="form.values.cod_max_satang"
            :label="$t('admin.settings.codMax')"
            :hint="$t('admin.settings.codMaxHint')"
            :error="form.errors.value.cod_max_satang"
            required
          />
          <BaseInput
            v-model="form.values.payment_window_minutes"
            type="number"
            inputmode="numeric"
            min="5"
            max="1440"
            :label="$t('admin.settings.paymentWindow')"
            :hint="$t('admin.settings.paymentWindowHint')"
            :error="form.errors.value.payment_window_minutes"
            required
          />
        </div>
      </section>

      <section class="card card-pad">
        <h2 class="admin-section-title">{{ $t('admin.settings.taxes') }}</h2>
        <div class="grid-form cols-2">
          <BaseInput
            v-model="form.values.vat_percent"
            type="number"
            inputmode="decimal"
            step="0.01"
            min="0"
            max="30"
            :label="$t('admin.settings.vatPercent')"
            :error="form.errors.value.vat_percent"
            required
          />
        </div>
      </section>

      <section class="card card-pad">
        <h2 class="admin-section-title">{{ $t('admin.settings.support') }}</h2>
        <div class="grid-form cols-2">
          <BaseInput
            v-model="form.values.support_email"
            type="email"
            :label="$t('admin.settings.supportEmail')"
            :error="form.errors.value.support_email"
            autocomplete="email"
          />
          <BaseInput
            v-model="form.values.support_phone"
            type="tel"
            :label="$t('admin.settings.supportPhone')"
            :error="form.errors.value.support_phone"
            autocomplete="tel"
          />
        </div>
      </section>

      <div class="save">
        <p v-if="formError" class="text-danger text-sm" role="alert">{{ formError }}</p>
        <BaseButton type="submit" size="lg" :loading="form.submitting.value">
          <Save :size="18" aria-hidden="true" />{{ $t('common.save') }}
        </BaseButton>
      </div>
    </form>
  </div>
</template>

<style scoped>
.settings {
  max-width: 880px;
}
.save {
  position: sticky;
  bottom: var(--sp-3);
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: var(--sp-3);
  flex-wrap: wrap;
}
.save p {
  margin: 0;
}
</style>
