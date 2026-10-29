import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { pickLocalized, type AppLocale } from '@/utils/localized'
import { formatDate, formatDateTime } from '@/utils/date'

/** Locale-aware helpers for bilingual API objects and dates. */
export function useLocalized() {
  const { locale } = useI18n()
  const current = computed(() => locale.value as AppLocale)

  function loc<K extends string>(
    obj: { [P in `${K}_en` | `${K}_my`]?: string | null } | null | undefined,
    field: K,
  ): string {
    return pickLocalized(obj, field, current.value)
  }

  const date = (iso: string | null | undefined) => formatDate(iso, current.value)
  const dateTime = (iso: string | null | undefined) => formatDateTime(iso, current.value)

  return { locale: current, loc, date, dateTime }
}
