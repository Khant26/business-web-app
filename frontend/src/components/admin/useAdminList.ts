import { computed, reactive, watch } from 'vue'
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'
import { usePaged } from '@/composables/usePaged'
import type { Paged } from '@/types/api'

type FilterValue = string | boolean

/**
 * Paged admin list whose filters and page are mirrored in the URL query.
 * Booleans are encoded as `1`; values equal to the defaults are omitted.
 * Keys in `debounce` (e.g. a search box) reload after a short pause.
 */
export function useAdminList<T, F extends Record<string, FilterValue>>(
  defaults: F,
  fetcher: (page: number, pageSize: number, filters: F) => Promise<Paged<T>>,
  opts: { pageSize?: number; debounce?: (keyof F)[]; sanitize?: (f: F) => void } = {},
) {
  const route = useRoute()
  const router = useRouter()

  const initial = { ...defaults } as Record<string, FilterValue>
  for (const [key, def] of Object.entries(defaults)) {
    const v = route.query[key]
    if (typeof v !== 'string') continue
    initial[key] = typeof def === 'boolean' ? v === '1' : v
  }
  const filters = reactive(initial) as F
  opts.sanitize?.(filters)

  const list = usePaged((page, pageSize) => fetcher(page, pageSize, { ...filters } as F), {
    pageSize: opts.pageSize ?? 20,
    immediate: false,
  })

  function sync() {
    const query: LocationQueryRaw = {}
    for (const [key, def] of Object.entries(defaults)) {
      const v = (filters as Record<string, FilterValue>)[key]
      if (v === def || v === '' || v === false) continue
      query[key] = v === true ? '1' : String(v)
    }
    if (list.page.value > 1) query.page = String(list.page.value)
    void router.replace({ query })
  }

  watch(list.page, sync)

  const debounced = new Set<string>((opts.debounce ?? []).map(String))
  const keys = Object.keys(defaults)
  let timer: ReturnType<typeof setTimeout> | undefined

  function refresh() {
    list.reload()
    sync()
  }

  watch(
    () =>
      keys.filter((k) => !debounced.has(k)).map((k) => (filters as Record<string, FilterValue>)[k]),
    refresh,
  )
  watch(
    () => [...debounced].map((k) => (filters as Record<string, FilterValue>)[k]),
    () => {
      clearTimeout(timer)
      timer = setTimeout(refresh, 350)
    },
  )

  const startPage = Math.max(1, Number(route.query.page) || 1)
  if (startPage > 1) list.setPage(startPage)
  else void list.load()

  const hasFilters = computed(() =>
    keys.some((k) => (filters as Record<string, FilterValue>)[k] !== defaults[k]),
  )

  function clear() {
    Object.assign(filters, defaults)
  }

  return { filters, list, hasFilters, clear }
}
