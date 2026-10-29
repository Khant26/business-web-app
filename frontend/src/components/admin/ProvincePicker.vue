<script setup lang="ts">
/**
 * Multi-select of Thai provinces (English names are the values). Provinces that
 * belong to another zone are shown with that zone's name and disabled.
 */
import { computed, onMounted, ref } from 'vue'
import { Search } from 'lucide-vue-next'
import { catalogApi } from '@/api/catalog'
import BaseCheckbox from '@/components/ui/BaseCheckbox.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { useAsync } from '@/composables/useAsync'

const model = defineModel<string[]>({ required: true })
const props = defineProps<{ takenBy: Record<string, string>; error?: string }>()

const provinces = useAsync(() => catalogApi.provinces())
onMounted(() => void provinces.run())

const q = ref('')
const onlyFree = ref(false)

const filtered = computed(() => {
  const term = q.value.trim().toLowerCase()
  return (provinces.data.value ?? [])
    .filter((p) => !onlyFree.value || !props.takenBy[p.name_en])
    .filter(
      (p) => !term || p.name_en.toLowerCase().includes(term) || p.name_th.includes(q.value.trim()),
    )
    .sort((a, b) => a.name_en.localeCompare(b.name_en))
})

const selected = computed(() => new Set(model.value))

function toggle(name: string, on: boolean) {
  const next = new Set(model.value)
  if (on) next.add(name)
  else next.delete(name)
  model.value = [...next].sort()
}

function selectVisible() {
  const next = new Set(model.value)
  for (const p of filtered.value) if (!props.takenBy[p.name_en]) next.add(p.name_en)
  model.value = [...next].sort()
}
</script>

<template>
  <div class="pp" :class="{ invalid: !!error }">
    <div class="bar">
      <label class="search">
        <Search :size="16" aria-hidden="true" />
        <span class="sr-only">{{ $t('admin.zones.searchProvinces') }}</span>
        <input v-model="q" type="search" :placeholder="$t('admin.zones.searchProvinces')" />
      </label>
      <BaseCheckbox v-model="onlyFree" :label="$t('admin.zones.onlyUnassigned')" />
      <span class="count">{{ $t('admin.zones.selectedN', { n: model.length }) }}</span>
      <BaseButton size="sm" variant="secondary" @click="selectVisible">
        {{ $t('admin.zones.selectVisible') }}
      </BaseButton>
      <BaseButton size="sm" variant="ghost" :disabled="!model.length" @click="model = []">
        {{ $t('admin.zones.clearAll') }}
      </BaseButton>
    </div>
    <div v-if="provinces.loading.value" class="grid">
      <SkeletonBlock v-for="i in 12" :key="i" height="36px" />
    </div>
    <ErrorState
      v-else-if="provinces.error.value"
      compact
      :error="provinces.error.value"
      :title="$t('admin.zones.provincesLoadError')"
      @retry="provinces.retry"
    />
    <div v-else class="grid" role="group" :aria-label="$t('admin.zones.provinces')">
      <label
        v-for="p in filtered"
        :key="p.name_en"
        class="opt"
        :class="{ on: selected.has(p.name_en), taken: !!takenBy[p.name_en] }"
      >
        <input
          type="checkbox"
          :checked="selected.has(p.name_en)"
          :disabled="!!takenBy[p.name_en]"
          @change="toggle(p.name_en, ($event.target as HTMLInputElement).checked)"
        />
        <span class="names">
          <span class="en">{{ p.name_en }}</span>
          <span class="th" lang="th">{{ p.name_th }}</span>
          <span v-if="takenBy[p.name_en]" class="zone">
            {{ $t('admin.zones.takenBy', { zone: takenBy[p.name_en] }) }}
          </span>
        </span>
      </label>
    </div>
    <p v-if="error" class="err" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.pp {
  border: 1px solid var(--c-border-strong);
  border-radius: var(--radius-md);
  padding: var(--sp-3);
}
.pp.invalid {
  border-color: var(--c-danger);
}
.bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--sp-2) var(--sp-3);
  margin-bottom: var(--sp-3);
}
.search {
  flex: 1 1 200px;
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
  height: 38px;
  width: 100%;
  background: transparent;
  font-size: 16px;
}
.count {
  font-size: var(--fs-xs);
  color: var(--c-muted);
  margin-left: auto;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
  gap: var(--sp-2);
  max-height: 360px;
  overflow-y: auto;
  padding: 2px;
}
.opt {
  display: flex;
  align-items: flex-start;
  gap: var(--sp-2);
  padding: var(--sp-2);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--fs-sm);
}
.opt:hover {
  border-color: var(--c-primary-300);
}
.opt.on {
  border-color: var(--c-primary);
  background: var(--c-primary-50);
}
.opt.taken {
  cursor: not-allowed;
  background: var(--c-surface-2);
  color: var(--c-muted);
}
.opt input {
  margin-top: 3px;
  accent-color: var(--c-primary);
}
.names {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
  min-width: 0;
}
.en {
  font-weight: var(--fw-medium);
}
.th,
.zone {
  font-size: var(--fs-xs);
  color: var(--c-muted);
}
.zone {
  color: var(--c-warning);
}
.err {
  margin: var(--sp-2) 0 0;
  color: var(--c-danger);
  font-size: var(--fs-xs);
}
</style>
