import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastKind = 'success' | 'error' | 'info' | 'warning'

export interface Toast {
  id: number
  kind: ToastKind
  message: string
  timeout: number
}

let nextId = 1

export const useUiStore = defineStore('ui', () => {
  const toasts = ref<Toast[]>([])

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function toast(message: string, kind: ToastKind = 'info', timeout = 4000): number {
    const id = nextId++
    toasts.value = [...toasts.value.slice(-3), { id, kind, message, timeout }]
    if (timeout > 0) setTimeout(() => dismiss(id), timeout)
    return id
  }

  const success = (m: string) => toast(m, 'success')
  const error = (m: string) => toast(m, 'error', 6000)
  const info = (m: string) => toast(m, 'info')
  const warning = (m: string) => toast(m, 'warning', 5000)

  return { toasts, toast, success, error, info, warning, dismiss }
})
