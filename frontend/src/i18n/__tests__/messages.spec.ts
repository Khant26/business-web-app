import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import en from '../locales/en.json'
import my from '../locales/my.json'
import accountEn from '../locales/account.en.json'
import accountMy from '../locales/account.my.json'
import adminEn from '../locales/admin.en.json'
import adminMy from '../locales/admin.my.json'

type Tree = { [key: string]: string | Tree }

function keys(tree: Tree, prefix = ''): string[] {
  return Object.entries(tree).flatMap(([k, v]) =>
    typeof v === 'string' ? [prefix + k] : keys(v, `${prefix}${k}.`),
  )
}

const bundles = {
  en: { ...en, ...accountEn, ...adminEn } as Tree,
  my: { ...my, ...accountMy, ...adminMy } as Tree,
}

describe('locale messages', () => {
  it('English and Myanmar have the same keys', () => {
    expect(keys(bundles.my).sort()).toEqual(keys(bundles.en).sort())
  })

  // Catches vue-i18n syntax problems such as a bare "@" (linked message) or unbalanced braces.
  it.each(['en', 'my'] as const)('every %s message compiles', (locale) => {
    const i18n = createI18n({
      legacy: false,
      locale,
      messages: bundles,
      missingWarn: false,
      fallbackWarn: false,
    })
    for (const key of keys(bundles[locale])) {
      expect(() => i18n.global.t(key, { n: 1, count: 1 }), key).not.toThrow()
    }
  })
})
