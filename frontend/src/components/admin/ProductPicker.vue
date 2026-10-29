<script setup lang="ts">
/** Searchable multi-select of products (by id). */
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Search, X } from 'lucide-vue-next'
import { adminApi } from '@/api/admin'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ImagePlaceholder from '@/components/ui/ImagePlaceholder.vue'
import type { AdminProductRow } from '@/types/api'

const model = defineModel<number[]>({ required: true })

const q = ref('')
const loading = ref(false)
const results = ref<AdminProductRow[]>([])
const names = reactive(new Map<number, string>())
let timer: ReturnType<typeof setTimeout> | undefined
let seq = 0

async function search() {
  const id = ++seq
  loading.value = true
  try {
    const res = await adminApi.products({ q: q.value.trim() || undefined, page_size: 100 })
    if (id !== seq) return
    results.value = res.items
    for (const p of res.items) names.set(p.id, p.name_en)
  } catch {
    if (id === seq) results.value = []
  } finally {
    if (id === seq) loading.value = false
  }
}

watch(q, () => {
  clearTimeout(timer)
  timer = setTimeout(() => void search(), 250)
})
onMounted(() => void search())

const selected = computed(() => new Set(model.value))

function toggle(id: number) {
  model.value = selected.value.has(id) ? model.value.filter((x) => x !== id) : [...model.value, id]
}
</script>

<template>
  <div class="ppk">
    <ul v-if="model.length" class="chips">
      <li v-for="id in model" :key="id" class="chip">
        <span>{{ names.get(id) ?? `#${id}` }}</span>
        <button
          type="button"
          :aria-label="`${$t('common.remove')} ${names.get(id) ?? id}`"
          @click="toggle(id)"
        >
          <X :size="14" aria-hidden="true" />
        </button>
      </li>
    </ul>
    <label class="search">
      <Search :size="16" aria-hidden="true" />
      <span class="sr-only">{{ $t('admin.events.searchProducts') }}</span>
      <input v-model="q" type="search" :placeholder="$t('admin.events.searchProducts')" />
    </label>
    <p class="count">{{ $t('admin.events.selectedN', { n: model.length }) }}</p>
    <div class="list">
      <template v-if="loading">
        <SkeletonBlock v-for="i in 4" :key="i" height="40px" />
      </template>
      <p v-else-if="results.length === 0" class="none">{{ $t('admin.events.noMatches') }}</p>
      <label
        v-for="p in loading ? [] : results"
        :key="p.id"
        class="row"
        :class="{ on: selected.has(p.id) }"
      >
        <input type="checkbox" :checked="selected.has(p.id)" @change="toggle(p.id)" />
        <ImagePlaceholder class="thumb" :src="p.image_url" :name="p.name_en" small />
        <span class="nm">{{ p.name_en }}</span>
        <span class="st">{{ $t(`admin.products.status.${p.status}`) }}</span>
      </label>
    </div>
  </div>
</template>

<style scoped>
.ppk {
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
.chips {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--c-primary-50);
  color: var(--c-primary-hover);
  border: 1px solid var(--c-primary-200);
  border-radius: var(--radius-full);
  padding: 2px 4px 2px 10px;
  font-size: var(--fs-xs);
  font-weight: var(--fw-medium);
}
.chip button {
  border: 0;
  background: transparent;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  cursor: pointer;
  color: inherit;
}
.chip button:hover {
  background: var(--c-primary-100);
}
.search {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  border: 1px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  padding: 0 var(--sp-3);
  color: var(--c-muted);
}
.search:focus-within {
  border-color: var(--c-primary-600);
  box-shadow: var(--ring);
}
.search input {
  border: 0;
  outline: none;
  height: 40px;
  width: 100%;
  background: transparent;
  font-size: 16px;
}
.count {
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.list {
  max-height: 260px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-md);
  padding: var(--sp-2);
}
.row {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: 4px var(--sp-2);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: var(--fs-sm);
}
.row:hover {
  background: var(--c-surface-2);
}
.row.on {
  background: var(--c-primary-50);
}
.row input {
  accent-color: var(--c-primary);
}
.thumb {
  width: 32px;
  flex: none;
}
.nm {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.st {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.none {
  margin: 0;
  padding: var(--sp-3);
  text-align: center;
  color: var(--c-muted);
  font-size: var(--fs-sm);
}
</style>
