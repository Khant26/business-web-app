import { reactive, ref, type UnwrapRef } from 'vue'
import { useI18n } from 'vue-i18n'
import type { z } from 'zod'
import { isApiError } from '@/api/client'

/**
 * Builds a zod issue message that carries an i18n key plus params,
 * e.g. msg('validation.maxLength', { max: 1000 }) → "validation.maxLength?max=1000".
 */
export function msg(key: string, params?: Record<string, string | number>): string {
  if (!params) return key
  const qs = new URLSearchParams(Object.entries(params).map(([k, v]) => [k, String(v)]))
  return `${key}?${qs.toString()}`
}

/**
 * Small form helper: reactive values, zod validation with translated per-field
 * errors, and mapping of API `fields` errors onto the same error map.
 */
export function useForm<S extends z.ZodType<Record<string, unknown>>>(
  schema: S,
  initial: z.input<S>,
) {
  const { t, te } = useI18n()
  const values = reactive({ ...(initial as object) }) as UnwrapRef<z.input<S>>
  const errors = ref<Record<string, string>>({})
  const submitting = ref(false)

  function translate(message: string): string {
    const [key = '', query] = message.split('?')
    if (!te(key)) return message
    const params = query ? Object.fromEntries(new URLSearchParams(query)) : {}
    return t(key, params)
  }

  /** Returns parsed output, or null after filling `errors`. */
  function validate(): z.output<S> | null {
    const result = schema.safeParse(values)
    if (result.success) {
      errors.value = {}
      return result.data
    }
    const next: Record<string, string> = {}
    for (const issue of result.error.issues) {
      const path = issue.path.join('.') || '_form'
      if (!next[path]) next[path] = translate(issue.message)
    }
    errors.value = next
    return null
  }

  function validateField(field: string) {
    const result = schema.safeParse(values)
    const next = { ...errors.value }
    delete next[field]
    if (!result.success) {
      const issue = result.error.issues.find((i) => i.path.join('.') === field)
      if (issue) next[field] = translate(issue.message)
    }
    errors.value = next
  }

  /** Copies API validation `fields` into the error map. Returns true when any were applied. */
  function applyApiError(e: unknown): boolean {
    if (!isApiError(e) || !e.fields || Object.keys(e.fields).length === 0) return false
    const next = { ...errors.value }
    for (const [field, message] of Object.entries(e.fields)) {
      next[field] = te(`errors.${message}`) ? t(`errors.${message}`) : message
    }
    errors.value = next
    return true
  }

  function reset(v: z.input<S> = initial) {
    Object.assign(values as object, v as object)
    errors.value = {}
  }

  return { values, errors, submitting, validate, validateField, applyApiError, reset, translate }
}
