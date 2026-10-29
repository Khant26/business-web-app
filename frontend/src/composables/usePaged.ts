import { computed, ref, watch, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Paged } from '@/types/api'
import { useAsync } from './useAsync'

/**
 * Paged list loader. The page number is synced to `?page=` in the URL when
 * `syncRoute` is true. Call `reload()` after filters change (resets to page 1).
 */
export function usePaged<T>(
  fetcher: (page: number, pageSize: number) => Promise<Paged<T>>,
  opts: { pageSize?: number; syncRoute?: boolean; immediate?: boolean } = {},
) {
  const pageSize = opts.pageSize ?? 20
  const route = opts.syncRoute ? useRoute() : null
  const router = opts.syncRoute ? useRouter() : null

  const initialPage = route ? Math.max(1, Number(route.query.page) || 1) : 1
  const page: Ref<number> = ref(initialPage)
  const state = useAsync((p: number) => fetcher(p, pageSize))

  const items = computed<T[]>(() => state.data.value?.items ?? [])
  const total = computed(() => state.data.value?.total ?? 0)
  const pages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

  function load() {
    return state.run(page.value)
  }

  function setPage(p: number) {
    page.value = Math.max(1, p)
  }

  function reload() {
    if (page.value !== 1) page.value = 1
    else void load()
  }

  watch(page, (p) => {
    if (route && router) {
      const q = { ...route.query }
      if (p > 1) q.page = String(p)
      else delete q.page
      void router.replace({ query: q })
    }
    void load()
  })

  if (route) {
    watch(
      () => route.query.page,
      (v) => {
        const p = Math.max(1, Number(v) || 1)
        if (p !== page.value) page.value = p
      },
    )
  }

  if (opts.immediate !== false) void load()

  return {
    page,
    pageSize,
    items,
    total,
    pages,
    loading: state.loading,
    error: state.error,
    data: state.data,
    load,
    reload,
    setPage,
    retry: state.retry,
  }
}
