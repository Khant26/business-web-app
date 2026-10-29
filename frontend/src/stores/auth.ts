import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authApi } from '@/api/auth'
import { isApiError } from '@/api/client'
import type { User } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  /** True once the initial /auth/me probe has finished. */
  const ready = ref(false)
  let inflight: Promise<User | null> | null = null

  const isLoggedIn = computed(() => user.value !== null)
  const isStaff = computed(() => user.value?.role === 'staff' || user.value?.role === 'admin')
  const isAdmin = computed(() => user.value?.role === 'admin')

  /** Loads the current user once (safe to call repeatedly; used by router guards). */
  function init(): Promise<User | null> {
    if (ready.value) return Promise.resolve(user.value)
    if (!inflight) {
      inflight = authApi
        .me()
        .then((r) => {
          user.value = r.user
          return r.user
        })
        .catch((e: unknown) => {
          if (!isApiError(e) || e.status !== 401) console.warn('auth/me failed', e)
          user.value = null
          return null
        })
        .finally(() => {
          ready.value = true
          inflight = null
        })
    }
    return inflight
  }

  async function requestOtp(email: string) {
    await authApi.requestOtp(email.trim().toLowerCase())
  }

  async function verifyOtp(email: string, code: string): Promise<User> {
    const r = await authApi.verifyOtp(email.trim().toLowerCase(), code)
    user.value = r.user
    ready.value = true
    return r.user
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      user.value = null
    }
  }

  function setUser(u: User | null) {
    user.value = u
  }

  /** Called by the API client's 401 hook. */
  function clear() {
    user.value = null
  }

  return {
    user,
    ready,
    isLoggedIn,
    isStaff,
    isAdmin,
    init,
    requestOtp,
    verifyOtp,
    logout,
    setUser,
    clear,
  }
})
