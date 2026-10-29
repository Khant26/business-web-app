import { ref, shallowRef, type Ref, type ShallowRef } from 'vue'

export interface AsyncState<T, A extends unknown[]> {
  data: ShallowRef<T | null>
  error: Ref<unknown>
  loading: Ref<boolean>
  run: (...args: A) => Promise<T | null>
  retry: () => Promise<T | null>
  set: (v: T | null) => void
}

/**
 * Wraps an async loader with loading/error state. Only the latest call wins
 * (stale responses are dropped), so it is safe to call on every route change.
 */
export function useAsync<T, A extends unknown[] = []>(
  fn: (...args: A) => Promise<T>,
  opts: { immediate?: boolean; initialArgs?: A } = {},
): AsyncState<T, A> {
  const data = shallowRef<T | null>(null)
  const error = ref<unknown>(null)
  const loading = ref(false)
  let lastArgs: A = (opts.initialArgs ?? []) as unknown as A
  let seq = 0

  async function run(...args: A): Promise<T | null> {
    lastArgs = args
    const id = ++seq
    loading.value = true
    error.value = null
    try {
      const result = await fn(...args)
      if (id === seq) data.value = result
      return result
    } catch (e) {
      if (id === seq) error.value = e
      return null
    } finally {
      if (id === seq) loading.value = false
    }
  }

  const retry = () => run(...lastArgs)
  const set = (v: T | null) => {
    data.value = v
  }

  if (opts.immediate) void run(...lastArgs)

  return { data, error, loading, run, retry, set }
}
