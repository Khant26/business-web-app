import { useI18n } from 'vue-i18n'
import { isApiError } from '@/api/client'

/** Translates thrown errors (ApiError codes → errors.<code>) into user-facing text. */
export function useErrorMessage() {
  const { t, te } = useI18n()

  function errorText(e: unknown): string {
    if (isApiError(e)) {
      if (e.code === 'rate_limited' && e.retryAfter) {
        return t('errors.rate_limited_wait', { s: e.retryAfter })
      }
      const key = `errors.${e.code}`
      if (te(key)) return t(key)
      if (e.message) return e.message
    }
    return t('errors.unknown')
  }

  function codeText(code: string | null | undefined): string {
    if (!code) return ''
    const key = `errors.${code}`
    return te(key) ? t(key) : code
  }

  return { errorText, codeText }
}
