import { defineStore } from 'pinia'
import { ref } from 'vue'
import { accountApi } from '@/api/account'

export const useWishlistStore = defineStore('wishlist', () => {
  const ids = ref<Set<number>>(new Set())
  const loaded = ref(false)

  async function load() {
    try {
      ids.value = new Set(await accountApi.wishlistIds())
      loaded.value = true
    } catch {
      /* ignore: not logged in / offline */
    }
  }

  function reset() {
    ids.value = new Set()
    loaded.value = false
  }

  const has = (productId: number) => ids.value.has(productId)

  /** Optimistically toggles; reverts on error. Returns the new state. */
  async function toggle(productId: number): Promise<boolean> {
    const was = ids.value.has(productId)
    const next = new Set(ids.value)
    if (was) next.delete(productId)
    else next.add(productId)
    ids.value = next
    try {
      if (was) await accountApi.removeWishlist(productId)
      else await accountApi.addWishlist(productId)
      return !was
    } catch (e) {
      const revert = new Set(ids.value)
      if (was) revert.add(productId)
      else revert.delete(productId)
      ids.value = revert
      throw e
    }
  }

  return { ids, loaded, load, reset, has, toggle }
})
