import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import my from './locales/my.json'
import accountEn from './locales/account.en.json'
import accountMy from './locales/account.my.json'
import type { AppLocale } from '@/utils/localized'

export const LOCALES: AppLocale[] = ['en', 'my']
export const LOCALE_STORAGE_KEY = 'sawyun-locale'

export type MessageSchema = typeof en

function readStoredLocale(): AppLocale {
  try {
    const v = localStorage.getItem(LOCALE_STORAGE_KEY)
    if (v === 'en' || v === 'my') return v
  } catch {
    /* storage unavailable */
  }
  return 'en'
}

export const i18n = createI18n({
  legacy: false,
  locale: readStoredLocale(),
  fallbackLocale: 'en',
  messages: {
    en: { ...en, ...accountEn },
    my: { ...my, ...accountMy },
  },
  missingWarn: false,
  fallbackWarn: false,
})

export function currentLocale(): AppLocale {
  return i18n.global.locale.value as AppLocale
}

function applyHtmlLang(locale: AppLocale) {
  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale === 'my' ? 'my' : 'en'
  }
}

export function setLocale(locale: AppLocale) {
  i18n.global.locale.value = locale
  try {
    localStorage.setItem(LOCALE_STORAGE_KEY, locale)
  } catch {
    /* ignore */
  }
  applyHtmlLang(locale)
}

applyHtmlLang(currentLocale())

let adminLoaded = false
/** Admin strings live in their own chunk; merged when the admin area is entered. */
export async function loadAdminMessages(): Promise<void> {
  if (adminLoaded) return
  const [adminEn, adminMy] = await Promise.all([
    import('./locales/admin.en.json'),
    import('./locales/admin.my.json'),
  ])
  i18n.global.mergeLocaleMessage('en', adminEn.default)
  i18n.global.mergeLocaleMessage('my', adminMy.default)
  adminLoaded = true
}
