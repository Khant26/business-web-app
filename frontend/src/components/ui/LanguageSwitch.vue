<script setup lang="ts">
import { Languages } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { setLocale } from '@/i18n'
import type { AppLocale } from '@/utils/localized'

withDefaults(defineProps<{ compact?: boolean }>(), {})
const { locale } = useI18n()

function choose(l: AppLocale) {
  setLocale(l)
}
</script>

<template>
  <div class="lang" role="group" :aria-label="$t('lang.switchTo')">
    <Languages v-if="!compact" :size="16" class="ic" aria-hidden="true" />
    <button
      type="button"
      :class="{ on: locale === 'en' }"
      :aria-pressed="locale === 'en'"
      lang="en"
      @click="choose('en')"
    >
      EN
    </button>
    <button
      type="button"
      :class="{ on: locale === 'my' }"
      :aria-pressed="locale === 'my'"
      lang="my"
      @click="choose('my')"
    >
      မြန်မာ
    </button>
  </div>
</template>

<style scoped>
.lang {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 3px;
  border-radius: var(--radius-full);
  background: var(--c-surface-2);
  border: 1px solid var(--c-border);
}
.ic {
  color: var(--c-muted);
  margin: 0 4px 0 6px;
}
button {
  border: 0;
  background: transparent;
  border-radius: var(--radius-full);
  padding: 4px 10px;
  font-size: var(--fs-xs);
  font-weight: var(--fw-semibold);
  color: var(--c-muted);
  cursor: pointer;
  line-height: 1.4;
}
button[lang='my'] {
  font-family: var(--font-my);
}
button.on {
  background: var(--c-surface);
  color: var(--c-primary);
  box-shadow: var(--shadow-xs);
}
button:focus-visible {
  outline: none;
  box-shadow: var(--ring);
}
</style>
