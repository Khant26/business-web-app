/** Zod field helpers for admin forms. Messages are i18n keys (see useForm's msg()). */
import { z } from 'zod'
import { msg } from '@/composables/useForm'

export const SLUG_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/

export const slugField = () =>
  z
    .string()
    .trim()
    .min(1, msg('validation.required'))
    .max(80, msg('validation.maxLength', { max: 80 }))
    .regex(SLUG_RE, msg('validation.slug'))

export const requiredText = (max = 200) =>
  z
    .string()
    .trim()
    .min(1, msg('validation.required'))
    .max(max, msg('validation.maxLength', { max }))

export const optionalText = (max = 200) =>
  z.string().trim().max(max, msg('validation.maxLength', { max }))

/** Integer typed into a text/number input (string | number), required. */
export const intField = (opts: { min?: number; max?: number } = {}) => {
  let n = z.number({ message: msg('validation.number') }).int(msg('validation.integer'))
  if (opts.min !== undefined) n = n.min(opts.min, msg('validation.min', { min: opts.min }))
  if (opts.max !== undefined) n = n.max(opts.max, msg('validation.max', { max: opts.max }))
  return z
    .union([z.string(), z.number()])
    .refine((v) => String(v).trim() !== '', msg('validation.required'))
    .transform((v) => Number(v))
    .pipe(n)
}

/** Optional integer: empty → null. */
export const optionalIntField = (opts: { min?: number; max?: number } = {}) => {
  let n = z.number({ message: msg('validation.number') }).int(msg('validation.integer'))
  if (opts.min !== undefined) n = n.min(opts.min, msg('validation.min', { min: opts.min }))
  if (opts.max !== undefined) n = n.max(opts.max, msg('validation.max', { max: opts.max }))
  return z
    .union([z.string(), z.number(), z.null()])
    .transform((v) => (v === null || String(v).trim() === '' ? null : Number(v)))
    .pipe(n.nullable())
}

/** Satang value produced by MoneyInput (NaN when invalid/empty). */
export const satangField = () =>
  z
    .number({ message: msg('validation.money') })
    .int(msg('validation.money'))
    .min(0, msg('validation.min', { min: 0 }))

export const optionalSatangField = () => satangField().nullable()

export const emailField = () =>
  z.string().trim().toLowerCase().min(1, msg('validation.required')).email(msg('validation.email'))
