<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'
import { Search, UserPlus } from 'lucide-vue-next'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import DataTable from '@/components/admin/DataTable.vue'
import { useAdminList } from '@/components/admin/useAdminList'
import { emailField, optionalText } from '@/components/admin/formFields'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseSelect from '@/components/ui/BaseSelect.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import BasePagination from '@/components/ui/BasePagination.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { adminApi } from '@/api/admin'
import { useForm } from '@/composables/useForm'
import { useLocalized } from '@/composables/useLocalized'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import type { AdminUser, Role, User } from '@/types/api'
import type { Column } from '@/components/admin/types'
import type { SelectOption } from '@/types/ui'

const { t } = useI18n()
const ui = useUiStore()
const auth = useAuthStore()
const { date } = useLocalized()
const { errorText } = useErrorMessage()

const { filters, list, hasFilters, clear } = useAdminList(
  { q: '', role: '' as Role | '' },
  (page, page_size, f) =>
    adminApi.users({ page, page_size, q: f.q.trim() || undefined, role: f.role || undefined }),
  {
    debounce: ['q'],
    sanitize: (f) => {
      if (!['customer', 'staff', 'admin'].includes(f.role)) f.role = ''
    },
  },
)

const roleFilterOptions = computed<SelectOption<Role | ''>[]>(() => [
  { value: '', label: t('admin.users.allRoles') },
  { value: 'customer', label: t('admin.roles.customer') },
  { value: 'staff', label: t('admin.roles.staff') },
  { value: 'admin', label: t('admin.roles.admin') },
])
const roleOptions = computed<SelectOption<Role>[]>(() => [
  { value: 'customer', label: t('admin.roles.customer') },
  { value: 'staff', label: t('admin.roles.staff') },
  { value: 'admin', label: t('admin.roles.admin') },
])

const columns = computed<Column[]>(() => [
  { key: 'email', label: t('admin.common.email'), primary: true },
  { key: 'role', label: t('admin.users.role') },
  { key: 'order_count', label: t('admin.users.orders'), align: 'right' },
  { key: 'created_at', label: t('admin.users.joined') },
  { key: 'is_active', label: t('admin.common.status') },
  { key: 'actions', label: t('admin.common.actions'), align: 'right' },
])

const isSelf = (u: AdminUser) => u.id === auth.user?.id

function patchRow(id: number, patch: Partial<AdminUser> | User) {
  const d = list.data.value
  if (d)
    list.data.value = { ...d, items: d.items.map((u) => (u.id === id ? { ...u, ...patch } : u)) }
}

const busy = ref<number | null>(null)

async function changeRole(u: AdminUser, role: Role) {
  if (role === u.role) return
  const prev = u.role
  patchRow(u.id, { role })
  busy.value = u.id
  try {
    const updated = await adminApi.updateUser(u.id, { role })
    patchRow(u.id, updated)
    ui.success(t('admin.users.roleChanged'))
  } catch (e) {
    patchRow(u.id, { role: prev })
    ui.error(errorText(e))
  } finally {
    busy.value = null
  }
}

const toDisable = ref<AdminUser | null>(null)
const confirmOpen = ref(false)

async function setActive(u: AdminUser, is_active: boolean) {
  busy.value = u.id
  try {
    const updated = await adminApi.updateUser(u.id, { is_active })
    patchRow(u.id, { ...updated, is_active })
    ui.success(t('admin.users.statusChanged'))
    confirmOpen.value = false
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    busy.value = null
  }
}

function askDisable(u: AdminUser) {
  toDisable.value = u
  confirmOpen.value = true
}

// ---------- Add staff ----------
const schema = z.object({
  email: emailField(),
  name: optionalText(120),
  role: z.enum(['staff', 'admin']),
})
const form = useForm(schema, { email: '', name: '', role: 'staff' })
const addOpen = ref(false)
const addError = ref('')
const staffRoleOptions = computed<SelectOption<'staff' | 'admin'>[]>(() => [
  { value: 'staff', label: t('admin.roles.staff') },
  { value: 'admin', label: t('admin.roles.admin') },
])

function openAdd() {
  form.reset({ email: '', name: '', role: 'staff' })
  addError.value = ''
  addOpen.value = true
}

async function add() {
  const d = form.validate()
  if (!d) return
  form.submitting.value = true
  addError.value = ''
  try {
    await adminApi.createUser(d)
    addOpen.value = false
    ui.success(t('admin.users.added'))
    void list.load()
  } catch (e) {
    if (!form.applyApiError(e)) addError.value = errorText(e)
  } finally {
    form.submitting.value = false
  }
}
</script>

<template>
  <div>
    <AdminPageHeader :title="$t('admin.nav.users')">
      <BaseButton @click="openAdd"
        ><UserPlus :size="18" aria-hidden="true" />{{ $t('admin.users.add') }}</BaseButton
      >
    </AdminPageHeader>

    <div class="admin-toolbar">
      <div class="grow admin-search">
        <Search :size="16" aria-hidden="true" />
        <label for="users-q" class="sr-only">{{ $t('admin.common.search') }}</label>
        <input
          id="users-q"
          v-model="filters.q"
          type="search"
          :placeholder="$t('admin.users.searchPlaceholder')"
        />
      </div>
      <BaseSelect
        v-model="filters.role"
        class="fixed"
        :label="$t('admin.users.role')"
        :options="roleFilterOptions"
      />
    </div>

    <ErrorState v-if="list.error.value" :error="list.error.value" @retry="list.retry" />
    <template v-else>
      <DataTable
        :columns="columns"
        :rows="list.items.value"
        :row-key="(u) => u.id"
        :loading="list.loading.value && !list.data.value"
        :caption="$t('admin.nav.users')"
      >
        <template #cell-email="{ row }">
          <span class="who">
            <span class="nm">
              {{ row.name || row.email }}
              <BaseBadge v-if="isSelf(row)" tone="primary" size="sm">{{
                $t('admin.users.you')
              }}</BaseBadge>
            </span>
            <span v-if="row.name" class="sub">{{ row.email }}</span>
            <span v-if="row.phone" class="sub">{{ row.phone }}</span>
          </span>
        </template>
        <template #cell-role="{ row }">
          <BaseSelect
            class="role-select"
            :model-value="row.role"
            :options="roleOptions"
            :disabled="isSelf(row) || busy === row.id"
            :title="isSelf(row) ? $t('admin.users.selfHint') : undefined"
            :aria-label="`${$t('admin.users.role')}: ${row.email}`"
            @update:model-value="(v) => v && changeRole(row, v)"
          />
        </template>
        <template #cell-order_count="{ row }"
          ><span class="tabular">{{ row.order_count }}</span></template
        >
        <template #cell-created_at="{ row }"
          ><span class="nowrap">{{ date(row.created_at) }}</span></template
        >
        <template #cell-is_active="{ row }">
          <BaseBadge :tone="row.is_active ? 'success' : 'danger'" size="sm" dot>
            {{ row.is_active ? $t('admin.users.enabled') : $t('admin.users.disabled') }}
          </BaseBadge>
        </template>
        <template #cell-actions="{ row }">
          <span class="admin-row-actions">
            <BaseButton
              v-if="row.is_active"
              size="sm"
              variant="secondary"
              :disabled="isSelf(row) || busy === row.id"
              :title="isSelf(row) ? $t('admin.users.selfHint') : undefined"
              @click="askDisable(row)"
            >
              {{ $t('admin.users.disable') }}
            </BaseButton>
            <BaseButton
              v-else
              size="sm"
              variant="secondary"
              :loading="busy === row.id"
              @click="setActive(row, true)"
            >
              {{ $t('admin.users.enable') }}
            </BaseButton>
          </span>
        </template>
        <template #empty>
          <EmptyState :title="$t('admin.users.empty')">
            <BaseButton v-if="hasFilters" variant="secondary" @click="clear">
              {{ $t('admin.common.clearFilters') }}
            </BaseButton>
          </EmptyState>
        </template>
      </DataTable>
      <BasePagination
        :page="list.page.value"
        :pages="list.pages.value"
        @update:page="list.setPage"
      />
    </template>

    <BaseModal v-model:open="addOpen" size="sm" :title="$t('admin.users.addTitle')">
      <form id="add-user" class="stack" novalidate @submit.prevent="add">
        <p class="muted text-sm hint">{{ $t('admin.users.addHint') }}</p>
        <BaseInput
          v-model="form.values.email"
          type="email"
          :label="$t('admin.common.email')"
          :error="form.errors.value.email"
          required
          autocomplete="off"
          data-autofocus
        />
        <BaseInput
          v-model="form.values.name"
          :label="$t('admin.common.name')"
          :error="form.errors.value.name"
          optional
        />
        <BaseSelect
          v-model="form.values.role"
          :label="$t('admin.users.role')"
          :options="staffRoleOptions"
        />
        <p v-if="addError" class="text-danger text-sm" role="alert">{{ addError }}</p>
      </form>
      <template #footer>
        <BaseButton variant="secondary" @click="addOpen = false">{{
          $t('common.cancel')
        }}</BaseButton>
        <BaseButton type="submit" form="add-user" :loading="form.submitting.value">{{
          $t('common.add')
        }}</BaseButton>
      </template>
    </BaseModal>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('admin.users.disableTitle')"
      :message="toDisable ? $t('admin.users.disableText', { email: toDisable.email }) : ''"
      :confirm-label="$t('admin.users.disable')"
      danger
      :loading="busy !== null"
      @confirm="toDisable && setActive(toDisable, false)"
    />
  </div>
</template>

<style scoped>
.who {
  display: inline-flex;
  flex-direction: column;
  min-width: 0;
}
.nm {
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
  font-weight: var(--fw-semibold);
  overflow-wrap: anywhere;
}
.sub {
  font-size: var(--fs-xs);
  color: var(--c-muted);
  overflow-wrap: anywhere;
}
.role-select {
  min-width: 130px;
}
.role-select :deep(select) {
  min-height: 36px;
}
.hint {
  margin: 0;
}
</style>
