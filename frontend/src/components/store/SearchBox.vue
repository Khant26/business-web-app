<script setup lang="ts">
import { ref, useId, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, X } from 'lucide-vue-next'

defineProps<{ autofocus?: boolean }>()
const emit = defineEmits<{ submitted: [] }>()

const route = useRoute()
const router = useRouter()
const q = ref(typeof route.query.q === 'string' ? route.query.q : '')
const input = ref<HTMLInputElement | null>(null)
const inputId = `search-${useId()}`

watch(
  () => route.query.q,
  (v) => {
    if (route.name === 'products') q.value = typeof v === 'string' ? v : ''
  },
)

function submit() {
  const term = q.value.trim()
  const query =
    route.name === 'products'
      ? { ...route.query, q: term || undefined, page: undefined }
      : { q: term || undefined }
  void router.push({ name: 'products', query })
  input.value?.blur()
  emit('submitted')
}

function clear() {
  q.value = ''
  input.value?.focus()
}

defineExpose({ focus: () => input.value?.focus() })
</script>

<template>
  <form class="search" role="search" @submit.prevent="submit">
    <label :for="inputId" class="sr-only">{{ $t('common.search') }}</label>
    <Search class="ic" :size="18" aria-hidden="true" />
    <input
      :id="inputId"
      ref="input"
      v-model="q"
      type="search"
      enterkeyhint="search"
      autocomplete="off"
      :placeholder="$t('common.searchPlaceholder')"
      :autofocus="autofocus"
    />
    <button v-if="q" type="button" class="clear" :aria-label="$t('common.clear')" @click="clear">
      <X :size="16" aria-hidden="true" />
    </button>
  </form>
</template>

<style scoped>
.search {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}
.ic {
  position: absolute;
  left: 14px;
  color: var(--c-muted);
  pointer-events: none;
}
input {
  width: 100%;
  height: 44px;
  padding: 0 40px 0 42px;
  border-radius: var(--radius-full);
  border: 1px solid var(--c-border);
  background: var(--c-surface-2);
  font-size: 16px;
  transition:
    background var(--dur-fast),
    border-color var(--dur-fast),
    box-shadow var(--dur-fast);
}
@media (min-width: 900px) {
  input {
    font-size: var(--fs-sm);
  }
}
input::-webkit-search-cancel-button {
  display: none;
}
input:focus {
  outline: none;
  background: var(--c-surface);
  border-color: var(--c-primary-500);
  box-shadow: var(--ring);
}
.clear {
  position: absolute;
  right: 8px;
  width: 30px;
  height: 30px;
  border: 0;
  background: transparent;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--c-muted);
  cursor: pointer;
}
.clear:hover {
  background: var(--c-surface-3);
}
</style>
