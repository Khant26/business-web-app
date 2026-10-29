import { watchEffect, type MaybeRefOrGetter, toValue } from 'vue'
import { useI18n } from 'vue-i18n'

/** Sets document.title to "<title> · <shop name>" reactively (for pages with dynamic titles). */
export function usePageTitle(title: MaybeRefOrGetter<string | null | undefined>) {
  const { t } = useI18n()
  watchEffect(() => {
    const v = toValue(title)
    if (v) document.title = `${v} · ${t('app.name')}`
  })
}
