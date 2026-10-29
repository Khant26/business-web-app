export type AppLocale = 'en' | 'my'

type Bilingual<K extends string> = { [P in `${K}_en` | `${K}_my`]?: string | null }

/**
 * Picks `<field>_<locale>` from an API object, falling back to English when the
 * Myanmar text is missing or blank.
 */
export function pickLocalized<K extends string>(
  obj: Bilingual<K> | null | undefined,
  field: K,
  locale: string,
): string {
  if (!obj) return ''
  const en = (obj[`${field}_en` as keyof Bilingual<K>] ?? '') as string
  if (locale !== 'my') return en
  const my = (obj[`${field}_my` as keyof Bilingual<K>] ?? '') as string
  return my.trim() ? my : en
}
