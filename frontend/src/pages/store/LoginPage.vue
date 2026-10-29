<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { ArrowLeft, Mail, ShieldCheck } from 'lucide-vue-next'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import OtpInput from '@/components/store/OtpInput.vue'
import { useForm, msg } from '@/composables/useForm'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { isApiError } from '@/api/client'

const RESEND_SECONDS = 60

const auth = useAuthStore()
const ui = useUiStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { errorText } = useErrorMessage()

const step = ref<'email' | 'code'>('email')
const emailForm = useForm(
  z.object({
    email: z.string().trim().min(1, msg('validation.required')).email(msg('validation.email')),
  }),
  { email: '' },
)
const email = ref('')
const code = ref('')
const codeError = ref('')
const formError = ref('')
const verifying = ref(false)
const sending = ref(false)
const otpRef = ref<InstanceType<typeof OtpInput> | null>(null)

// Resend countdown
const cooldown = ref(0)
let timer: ReturnType<typeof setInterval> | undefined
function startCooldown(seconds: number) {
  cooldown.value = seconds
  clearInterval(timer)
  timer = setInterval(() => {
    cooldown.value = Math.max(0, cooldown.value - 1)
    if (cooldown.value === 0) clearInterval(timer)
  }, 1000)
}
onBeforeUnmount(() => clearInterval(timer))

const redirectTo = computed(() => {
  const r = route.query.redirect
  return typeof r === 'string' && r.startsWith('/') && !r.startsWith('//') ? r : '/'
})

async function sendCode(isResend = false) {
  formError.value = ''
  if (!isResend) {
    const data = emailForm.validate()
    if (!data) return
    email.value = data.email.toLowerCase()
  }
  sending.value = true
  try {
    await auth.requestOtp(email.value)
    if (isResend) ui.info(t('auth.codeResent'))
    step.value = 'code'
    code.value = ''
    codeError.value = ''
    startCooldown(RESEND_SECONDS)
    await nextTick()
    otpRef.value?.focus()
  } catch (e) {
    if (isApiError(e) && e.code === 'rate_limited') {
      startCooldown(e.retryAfter ?? RESEND_SECONDS)
      if (!isResend && step.value === 'email') {
        // A code was likely sent moments ago — let the user enter it.
        step.value = 'code'
      }
    }
    if (!emailForm.applyApiError(e)) formError.value = errorText(e)
  } finally {
    sending.value = false
  }
}

async function verify(value = code.value) {
  if (verifying.value) return
  codeError.value = ''
  formError.value = ''
  if (!/^\d{6}$/.test(value)) {
    codeError.value = t('validation.otp')
    return
  }
  verifying.value = true
  try {
    const user = await auth.verifyOtp(email.value, value)
    ui.success(user.name ? `${t('auth.welcome')} ${user.name}` : t('auth.welcomeNew'))
    await router.replace(redirectTo.value)
  } catch (e) {
    codeError.value = errorText(e)
    if (isApiError(e)) {
      if (e.code === 'otp_invalid') {
        code.value = ''
        otpRef.value?.focus()
      } else if (e.code === 'otp_expired' || e.code === 'otp_too_many_attempts') {
        code.value = ''
        if (cooldown.value > 5) cooldown.value = 0
      } else if (e.code === 'rate_limited' && e.retryAfter) {
        startCooldown(e.retryAfter)
      }
    }
  } finally {
    verifying.value = false
  }
}

function changeEmail() {
  step.value = 'email'
  code.value = ''
  codeError.value = ''
  formError.value = ''
}
</script>

<template>
  <div class="container page login">
    <div class="panel card">
      <div class="badge-ic" aria-hidden="true">
        <Mail v-if="step === 'email'" :size="26" />
        <ShieldCheck v-else :size="26" />
      </div>

      <template v-if="step === 'email'">
        <h1>{{ $t('auth.title') }}</h1>
        <p class="muted">{{ $t('auth.subtitle') }}</p>
        <form novalidate class="stack" @submit.prevent="sendCode()">
          <BaseInput
            v-model="emailForm.values.email"
            type="email"
            inputmode="email"
            autocomplete="email"
            autofocus
            required
            :label="$t('auth.email')"
            :placeholder="$t('auth.emailPlaceholder')"
            :error="emailForm.errors.value.email"
          />
          <p v-if="formError" class="err" role="alert">{{ formError }}</p>
          <BaseButton type="submit" size="lg" block :loading="sending">
            {{ sending ? $t('auth.sending') : $t('auth.sendCode') }}
          </BaseButton>
        </form>
        <p class="muted text-xs terms">{{ $t('auth.terms') }}</p>
      </template>

      <template v-else>
        <h1>{{ $t('auth.codeTitle') }}</h1>
        <p class="muted">{{ $t('auth.codeSentTo', { email }) }}</p>
        <form novalidate class="stack" @submit.prevent="verify()">
          <OtpInput
            ref="otpRef"
            v-model="code"
            :disabled="verifying"
            :invalid="!!codeError"
            @complete="verify"
          />
          <p v-if="codeError" class="err" role="alert">{{ codeError }}</p>
          <p v-else-if="formError" class="err" role="alert">{{ formError }}</p>
          <BaseButton type="submit" size="lg" block :loading="verifying">
            {{ verifying ? $t('auth.verifying') : $t('auth.verify') }}
          </BaseButton>
          <div class="resend">
            <BaseButton variant="link" :disabled="cooldown > 0 || sending" @click="sendCode(true)">
              {{ cooldown > 0 ? $t('auth.resendIn', { s: cooldown }) : $t('auth.resend') }}
            </BaseButton>
            <BaseButton variant="link" @click="changeEmail">
              <ArrowLeft :size="14" aria-hidden="true" />{{ $t('auth.changeEmail') }}
            </BaseButton>
          </div>
          <p class="muted text-xs">{{ $t('auth.checkSpam') }}</p>
        </form>
      </template>
    </div>
  </div>
</template>

<style scoped>
.login {
  display: grid;
  place-items: center;
  min-height: calc(100dvh - var(--header-h) - 200px);
}
.panel {
  width: 100%;
  max-width: 440px;
  padding: var(--sp-8) var(--sp-5);
  box-shadow: var(--shadow-md);
}
@media (min-width: 640px) {
  .panel {
    padding: var(--sp-10);
  }
}
.badge-ic {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--c-primary-50);
  color: var(--c-primary);
  display: grid;
  place-items: center;
  margin-bottom: var(--sp-4);
}
h1 {
  font-size: var(--fs-xl);
  margin-bottom: var(--sp-2);
}
.stack {
  margin-top: var(--sp-5);
}
.err {
  margin: 0;
  color: var(--c-danger);
  font-size: var(--fs-sm);
  font-weight: var(--fw-medium);
}
.resend {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--sp-2);
}
.terms {
  margin-top: var(--sp-5);
  margin-bottom: 0;
}
</style>
