<script setup lang="ts">
import { z } from 'zod'
import { useI18n } from 'vue-i18n'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { useForm, msg } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useLocalized } from '@/composables/useLocalized'
import { accountApi } from '@/api/account'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { isThaiPhone, normalizePhone } from '@/utils/misc'

const auth = useAuthStore()
const ui = useUiStore()
const { t } = useI18n()
const { errorText } = useErrorMessage()
const { date } = useLocalized()

const schema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'validation.required')
    .max(100, msg('validation.maxLength', { max: 100 })),
  phone: z
    .string()
    .trim()
    .refine((v) => v === '' || isThaiPhone(v), 'validation.phone')
    .transform((v) => (v ? normalizePhone(v) : '')),
})

const { values, errors, submitting, validate, validateField, applyApiError } = useForm(schema, {
  name: auth.user?.name ?? '',
  phone: auth.user?.phone ?? '',
})

async function save() {
  const data = validate()
  if (!data) return
  submitting.value = true
  try {
    const r = await accountApi.updateProfile(data)
    auth.setUser(r.user)
    values.name = r.user.name
    values.phone = r.user.phone
    ui.success(t('account.profile.saved'))
  } catch (e) {
    if (!applyApiError(e)) ui.error(errorText(e))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="card card-pad">
    <h1 class="h">{{ $t('account.profile.title') }}</h1>
    <p class="muted intro">{{ $t('account.profile.intro') }}</p>

    <form class="grid-form" novalidate @submit.prevent="save">
      <BaseInput
        v-model="values.name"
        :label="$t('account.profile.name')"
        :error="errors.name"
        autocomplete="name"
        required
        maxlength="100"
        @blur="validateField('name')"
      />
      <BaseInput
        v-model="values.phone"
        :label="$t('account.profile.phone')"
        :placeholder="$t('address.phonePlaceholder')"
        :error="errors.phone"
        type="tel"
        inputmode="tel"
        autocomplete="tel"
        optional
        maxlength="20"
        @blur="validateField('phone')"
      />
      <BaseInput
        :model-value="auth.user?.email ?? ''"
        :label="$t('account.profile.email')"
        :hint="$t('account.profile.emailHint')"
        type="email"
        readonly
      />
      <div class="row foot">
        <BaseButton type="submit" :loading="submitting">
          {{ submitting ? $t('common.saving') : $t('common.save') }}
        </BaseButton>
        <span v-if="auth.user" class="muted text-sm">
          {{ $t('account.profile.memberSince', { date: date(auth.user.created_at) }) }}
        </span>
      </div>
    </form>
  </div>
</template>

<style scoped>
.h {
  font-size: var(--fs-xl);
  margin-bottom: var(--sp-1);
}
.intro {
  margin-bottom: var(--sp-5);
}
.grid-form {
  max-width: 520px;
}
.foot {
  margin-top: var(--sp-2);
}
</style>
