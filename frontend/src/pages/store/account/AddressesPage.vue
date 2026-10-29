<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { MapPin, Pencil, Plus, Star, Trash2 } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import AddressForm from '@/components/store/AddressForm.vue'
import AddressCard from '@/components/store/AddressCard.vue'
import {
  addressToFields,
  emptyAddress,
  useAddressValidation,
  type AddressFields,
} from '@/components/store/checkout/address'
import { accountApi } from '@/api/account'
import { isApiError } from '@/api/client'
import { useAsync } from '@/composables/useAsync'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import type { Address } from '@/types/api'

const auth = useAuthStore()
const ui = useUiStore()
const { t } = useI18n()
const { errorText } = useErrorMessage()
const { validate, fromApiFields } = useAddressValidation()

const state = useAsync(() => accountApi.addresses())
onMounted(() => void state.run())
const list = computed(() =>
  [...(state.data.value ?? [])].sort((a, b) => Number(b.is_default) - Number(a.is_default)),
)

// Edit / create
const editOpen = ref(false)
const editing = ref<Address | null>(null)
const fields = ref<AddressFields>(emptyAddress())
const fieldErrors = ref<Record<string, string>>({})
const saving = ref(false)

function openNew() {
  editing.value = null
  fields.value = {
    ...emptyAddress(),
    recipient_name: auth.user?.name ?? '',
    phone: auth.user?.phone ?? '',
    is_default: list.value.length === 0,
  }
  fieldErrors.value = {}
  editOpen.value = true
}
function openEdit(a: Address) {
  editing.value = a
  fields.value = addressToFields(a)
  fieldErrors.value = {}
  editOpen.value = true
}

async function save() {
  const r = validate(fields.value)
  if (!r.ok) {
    fieldErrors.value = r.errors
    return
  }
  saving.value = true
  try {
    if (editing.value) await accountApi.updateAddress(editing.value.id, r.data)
    else await accountApi.createAddress(r.data)
    editOpen.value = false
    ui.success(t('account.addresses.saved'))
    void state.run()
  } catch (e) {
    if (isApiError(e) && Object.keys(e.fields).length) fieldErrors.value = fromApiFields(e.fields)
    else ui.error(errorText(e))
  } finally {
    saving.value = false
  }
}

// Default
const busyId = ref<number | null>(null)
async function makeDefault(a: Address) {
  busyId.value = a.id
  try {
    await accountApi.updateAddress(a.id, { ...addressToFields(a), is_default: true })
    ui.success(t('account.addresses.defaultSet'))
    await state.run()
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    busyId.value = null
  }
}

// Delete
const toDelete = ref<Address | null>(null)
const deleteOpen = ref(false)
const deleting = ref(false)
function askDelete(a: Address) {
  toDelete.value = a
  deleteOpen.value = true
}
async function confirmDelete() {
  if (!toDelete.value) return
  deleting.value = true
  try {
    await accountApi.deleteAddress(toDelete.value.id)
    deleteOpen.value = false
    ui.success(t('account.addresses.deleted'))
    void state.run()
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <div class="head">
      <h1 class="h">{{ $t('account.addresses.title') }}</h1>
      <BaseButton v-if="list.length" size="sm" @click="openNew">
        <Plus :size="16" aria-hidden="true" />{{ $t('account.addresses.add') }}
      </BaseButton>
    </div>

    <div v-if="state.loading.value && !state.data.value" class="grid" aria-busy="true">
      <SkeletonBlock v-for="i in 2" :key="i" height="150px" radius="var(--radius-lg)" />
    </div>
    <ErrorState v-else-if="state.error.value" :error="state.error.value" @retry="state.retry" />
    <EmptyState
      v-else-if="list.length === 0"
      :icon="MapPin"
      :title="$t('account.addresses.empty')"
      :text="$t('account.addresses.emptyText')"
    >
      <BaseButton @click="openNew">
        <Plus :size="16" aria-hidden="true" />{{ $t('account.addresses.add') }}
      </BaseButton>
    </EmptyState>
    <ul v-else class="grid">
      <li v-for="a in list" :key="a.id" class="card card-pad item" :class="{ def: a.is_default }">
        <AddressCard :address="a" />
        <div class="actions">
          <BaseButton variant="ghost" size="sm" @click="openEdit(a)">
            <Pencil :size="15" aria-hidden="true" />{{ $t('common.edit') }}
          </BaseButton>
          <BaseButton
            v-if="!a.is_default"
            variant="ghost"
            size="sm"
            :loading="busyId === a.id"
            @click="makeDefault(a)"
          >
            <Star :size="15" aria-hidden="true" />{{ $t('account.addresses.makeDefault') }}
          </BaseButton>
          <BaseButton variant="ghost" size="sm" class="del" @click="askDelete(a)">
            <Trash2 :size="15" aria-hidden="true" />{{ $t('common.delete') }}
          </BaseButton>
        </div>
      </li>
    </ul>

    <BaseModal
      v-model:open="editOpen"
      size="lg"
      :title="editing ? $t('account.addresses.edit') : $t('account.addresses.new')"
    >
      <form id="addr-form" novalidate @submit.prevent="save">
        <AddressForm v-model="fields" :errors="fieldErrors" show-label show-default />
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="editOpen = false">{{
          $t('common.cancel')
        }}</BaseButton>
        <BaseButton type="submit" form="addr-form" :loading="saving">{{
          $t('common.save')
        }}</BaseButton>
      </template>
    </BaseModal>

    <ConfirmDialog
      v-model:open="deleteOpen"
      :title="$t('account.addresses.deleteTitle')"
      :message="$t('account.addresses.deleteMessage')"
      :confirm-label="$t('common.delete')"
      :loading="deleting"
      danger
      @confirm="confirmDelete"
    />
  </div>
</template>

<style scoped>
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--sp-3);
  margin-bottom: var(--sp-4);
}
.h {
  font-size: var(--fs-xl);
  margin: 0;
}
.grid {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--sp-4);
}
@media (min-width: 768px) {
  .grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
.item {
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.item.def {
  border-color: var(--c-primary-300);
  box-shadow: 0 0 0 1px var(--c-primary-200);
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sp-1);
  margin-top: auto;
  padding-top: var(--sp-2);
  border-top: 1px solid var(--c-border);
}
.del {
  color: var(--c-danger);
}
</style>
