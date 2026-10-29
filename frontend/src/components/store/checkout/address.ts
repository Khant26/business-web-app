import { z } from 'zod'
import { useI18n } from 'vue-i18n'
import { msg } from '@/composables/useForm'
import { isThaiPhone, normalizePhone } from '@/utils/misc'
import type { Address, AddressInput } from '@/types/api'

export type AddressFields = AddressInput

export function emptyAddress(): AddressFields {
  return {
    label: '',
    recipient_name: '',
    phone: '',
    line1: '',
    subdistrict: '',
    district: '',
    province: '',
    postcode: '',
    is_default: false,
  }
}

export function addressToFields(a: Address): AddressFields {
  return {
    label: a.label,
    recipient_name: a.recipient_name,
    phone: a.phone,
    line1: a.line1,
    subdistrict: a.subdistrict,
    district: a.district,
    province: a.province,
    postcode: a.postcode,
    is_default: a.is_default,
  }
}

const req = (max: number) =>
  z.string().trim().min(1, 'validation.required').max(max, msg('validation.maxLength', { max }))

export const addressSchema = z.object({
  label: z
    .string()
    .trim()
    .max(40, msg('validation.maxLength', { max: 40 })),
  recipient_name: req(100),
  phone: z
    .string()
    .trim()
    .min(1, 'validation.required')
    .refine(isThaiPhone, 'validation.phone')
    .transform(normalizePhone),
  line1: req(200),
  subdistrict: req(100),
  district: req(100),
  province: req(100),
  postcode: z
    .string()
    .trim()
    .regex(/^\d{5}$/, 'validation.postcode'),
  is_default: z.boolean(),
})

/** Validates address fields with translated per-field messages. */
export function useAddressValidation() {
  const { t, te } = useI18n()

  function translate(message: string): string {
    const [key = '', query] = message.split('?')
    if (!te(key)) return message
    return t(key, query ? Object.fromEntries(new URLSearchParams(query)) : {})
  }

  function validate(
    fields: AddressFields,
  ): { ok: true; data: AddressFields } | { ok: false; errors: Record<string, string> } {
    const r = addressSchema.safeParse(fields)
    if (r.success) return { ok: true, data: r.data }
    const errors: Record<string, string> = {}
    for (const issue of r.error.issues) {
      const k = issue.path.join('.')
      if (!errors[k]) errors[k] = translate(issue.message)
    }
    return { ok: false, errors }
  }

  /** Maps API `fields` (e.g. "address.phone" or "phone") to address field errors. */
  function fromApiFields(fields: Record<string, string>): Record<string, string> {
    const out: Record<string, string> = {}
    for (const [k, v] of Object.entries(fields)) {
      const key = k.replace(/^address\./, '')
      out[key] = te(`errors.${v}`) ? t(`errors.${v}`) : v
    }
    return out
  }

  return { validate, fromApiFields }
}
