<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { LogOut, Monitor, Smartphone, Tablet } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { accountApi } from '@/api/account'
import { useAsync } from '@/composables/useAsync'
import { useErrorMessage } from '@/composables/useErrorMessage'
import { useLocalized } from '@/composables/useLocalized'
import { useUiStore } from '@/stores/ui'
import type { Session } from '@/types/api'

const ui = useUiStore()
const { t } = useI18n()
const { errorText } = useErrorMessage()
const { dateTime } = useLocalized()

const state = useAsync(() => accountApi.sessions())
onMounted(() => void state.run())
const sessions = computed(() =>
  [...(state.data.value ?? [])].sort(
    (a, b) => Number(b.current) - Number(a.current) || b.last_seen_at.localeCompare(a.last_seen_at),
  ),
)
const others = computed(() => sessions.value.filter((s) => !s.current))

type Kind = 'phone' | 'tablet' | 'desktop'
function parseUa(ua: string): { label: string; kind: Kind } {
  if (!ua) return { label: t('account.security.unknownDevice'), kind: 'desktop' }
  const browser = /Edg\//.test(ua)
    ? 'Edge'
    : /OPR\/|Opera/.test(ua)
      ? 'Opera'
      : /SamsungBrowser/.test(ua)
        ? 'Samsung Internet'
        : /Firefox\//.test(ua)
          ? 'Firefox'
          : /Chrome\/|CriOS/.test(ua)
            ? 'Chrome'
            : /Safari\//.test(ua)
              ? 'Safari'
              : ''
  const os = /iPad/.test(ua)
    ? 'iPadOS'
    : /iPhone|iPod/.test(ua)
      ? 'iOS'
      : /Android/.test(ua)
        ? 'Android'
        : /Windows/.test(ua)
          ? 'Windows'
          : /Mac OS X|Macintosh/.test(ua)
            ? 'macOS'
            : /CrOS/.test(ua)
              ? 'ChromeOS'
              : /Linux/.test(ua)
                ? 'Linux'
                : ''
  const kind: Kind = /iPad|Tablet/.test(ua)
    ? 'tablet'
    : /Mobi|iPhone|Android/.test(ua)
      ? 'phone'
      : 'desktop'
  let label = t('account.security.unknownDevice')
  if (browser && os) label = t('account.security.on', { browser, os })
  else if (browser || os) label = browser || os
  return { label, kind }
}
const ICONS = { phone: Smartphone, tablet: Tablet, desktop: Monitor }

const busyId = ref<string | null>(null)
async function revoke(s: Session) {
  busyId.value = s.id
  try {
    await accountApi.revokeSession(s.id)
    state.set((state.data.value ?? []).filter((x) => x.id !== s.id))
    ui.success(t('account.security.revoked'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    busyId.value = null
  }
}

const confirmOpen = ref(false)
const revokingAll = ref(false)
async function revokeOthers() {
  revokingAll.value = true
  try {
    await accountApi.revokeOthers()
    confirmOpen.value = false
    ui.success(t('account.security.othersRevoked'))
    await state.run()
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    revokingAll.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="h">{{ $t('account.security.title') }}</h1>
    <p class="muted intro">{{ $t('account.security.intro') }}</p>

    <div v-if="state.loading.value && !state.data.value" class="list" aria-busy="true">
      <SkeletonBlock v-for="i in 3" :key="i" height="76px" radius="var(--radius-lg)" />
    </div>
    <ErrorState v-else-if="state.error.value" :error="state.error.value" @retry="state.retry" />
    <template v-else>
      <ul class="list">
        <li v-for="s in sessions" :key="s.id" class="card sess" :class="{ cur: s.current }">
          <span class="ic" aria-hidden="true">
            <component :is="ICONS[parseUa(s.user_agent).kind]" :size="22" />
          </span>
          <div class="info">
            <div class="top">
              <strong>{{ parseUa(s.user_agent).label }}</strong>
              <BaseBadge v-if="s.current" tone="success" size="sm" dot>
                {{ $t('account.security.current') }}
              </BaseBadge>
            </div>
            <span class="muted text-xs">
              {{ $t('account.security.lastSeen', { date: dateTime(s.last_seen_at) }) }}
              <template v-if="s.ip"> · {{ $t('account.security.ip', { ip: s.ip }) }}</template>
            </span>
            <span class="muted text-xs">
              {{ $t('account.security.signedIn', { date: dateTime(s.created_at) }) }}
            </span>
          </div>
          <BaseButton
            v-if="!s.current"
            variant="secondary"
            size="sm"
            :loading="busyId === s.id"
            :aria-label="$t('account.security.revokeAria', { device: parseUa(s.user_agent).label })"
            @click="revoke(s)"
          >
            <LogOut :size="15" aria-hidden="true" />
            <span class="rv-txt">{{ $t('account.security.revoke') }}</span>
          </BaseButton>
        </li>
      </ul>

      <div class="foot">
        <p v-if="others.length === 0" class="muted text-sm">
          {{ $t('account.security.noOthers') }}
        </p>
        <BaseButton v-else variant="danger" @click="confirmOpen = true">
          {{ $t('account.security.revokeOthers') }}
        </BaseButton>
      </div>
    </template>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('account.security.revokeOthersTitle')"
      :message="$t('account.security.revokeOthersMessage')"
      :confirm-label="$t('account.security.revokeOthers')"
      :loading="revokingAll"
      danger
      @confirm="revokeOthers"
    />
  </div>
</template>

<style scoped>
.h {
  font-size: var(--fs-xl);
  margin-bottom: var(--sp-1);
}
.intro {
  margin-bottom: var(--sp-5);
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.sess {
  display: flex;
  align-items: center;
  gap: var(--sp-3);
  padding: var(--sp-4);
}
.sess.cur {
  border-color: var(--c-primary-300);
}
.ic {
  flex: none;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: var(--c-primary-50);
  color: var(--c-primary);
  display: grid;
  place-items: center;
}
.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.top {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  flex-wrap: wrap;
}
.foot {
  margin-top: var(--sp-5);
}
@media (max-width: 479px) {
  .rv-txt {
    display: none;
  }
}
</style>
