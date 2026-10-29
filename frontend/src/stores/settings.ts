import { defineStore } from 'pinia'
import { ref } from 'vue'
import { catalogApi } from '@/api/catalog'
import type { PublicSettings } from '@/types/api'

const DEFAULTS: PublicSettings = {
  shop_name: 'Sawyun Ecommerce',
  vat_rate_bp: 700,
  payment_window_minutes: 30,
  cod_max_satang: 1_000_000,
  support_email: '',
  support_phone: '',
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<PublicSettings>({ ...DEFAULTS })
  const loaded = ref(false)
  let inflight: Promise<void> | null = null

  function load(force = false): Promise<void> {
    if (loaded.value && !force) return Promise.resolve()
    if (!inflight) {
      inflight = catalogApi
        .settings()
        .then((s) => {
          settings.value = { ...DEFAULTS, ...s }
          loaded.value = true
        })
        .catch(() => {
          /* keep defaults */
        })
        .finally(() => {
          inflight = null
        })
    }
    return inflight
  }

  return { settings, loaded, load }
})
