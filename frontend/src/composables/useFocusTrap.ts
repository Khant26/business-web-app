import { nextTick, onBeforeUnmount, watch, type Ref } from 'vue'

const FOCUSABLE =
  'a[href], area[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"]), [contenteditable="true"]'

/**
 * Traps Tab focus inside `container` while `active` is true, closes on Escape,
 * locks body scroll, and restores focus to the previously focused element.
 */
export function useFocusTrap(
  container: Ref<HTMLElement | null>,
  active: Ref<boolean>,
  onEscape?: () => void,
) {
  let previous: HTMLElement | null = null

  function focusables(): HTMLElement[] {
    if (!container.value) return []
    return Array.from(container.value.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
      (el) => el.offsetParent !== null || el === document.activeElement,
    )
  }

  function onKeydown(e: KeyboardEvent) {
    if (!active.value) return
    if (e.key === 'Escape' && onEscape) {
      e.stopPropagation()
      onEscape()
      return
    }
    if (e.key !== 'Tab') return
    const els = focusables()
    if (els.length === 0) {
      e.preventDefault()
      container.value?.focus()
      return
    }
    const first = els[0]!
    const last = els[els.length - 1]!
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault()
      first.focus()
    }
  }

  async function activate() {
    previous = document.activeElement as HTMLElement | null
    document.addEventListener('keydown', onKeydown, true)
    document.body.style.overflow = 'hidden'
    await nextTick()
    const target =
      container.value?.querySelector<HTMLElement>('[autofocus], [data-autofocus]') ??
      focusables()[0] ??
      container.value
    target?.focus()
  }

  function deactivate() {
    document.removeEventListener('keydown', onKeydown, true)
    document.body.style.overflow = ''
    if (previous && typeof previous.focus === 'function') previous.focus()
    previous = null
  }

  watch(
    active,
    (v, old) => {
      if (v) void activate()
      else if (old) deactivate()
    },
    { immediate: true },
  )

  onBeforeUnmount(() => {
    if (active.value) deactivate()
  })
}
