/** "Summer Sale 2026!" → "summer-sale-2026" */
export function slugify(input: string): string {
  return input
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 80)
}

/** Thai phone numbers: mobile 0[689]xxxxxxxx (10 digits) or landline 0[2-7]xxxxxxx (9 digits). */
export const THAI_PHONE_RE = /^0(?:[689]\d{8}|[2-7]\d{7})$/

export function normalizePhone(input: string): string {
  let digits = input.replace(/[^\d+]/g, '')
  if (digits.startsWith('+66')) digits = '0' + digits.slice(3)
  else if (digits.startsWith('66') && digits.length === 11) digits = '0' + digits.slice(2)
  return digits.replace(/\D/g, '')
}

export function isThaiPhone(input: string): boolean {
  return THAI_PHONE_RE.test(normalizePhone(input))
}

/** Absolute-safe image URL: API returns /uploads/... paths; leave other URLs untouched. */
export function imageUrl(url: string | null | undefined): string {
  return url ?? ''
}

/** First grapheme, uppercase — for placeholders/avatars. */
export function initial(text: string | null | undefined): string {
  const s = (text ?? '').trim()
  if (!s) return '?'
  const first = Array.from(s)[0] ?? '?'
  return first.toUpperCase()
}

/** Deterministic hue (0–359) from a string, for tinted placeholders. */
export function hueFrom(text: string): number {
  let h = 0
  for (let i = 0; i < text.length; i++) h = (h * 31 + text.charCodeAt(i)) >>> 0
  return h % 360
}

export function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}
