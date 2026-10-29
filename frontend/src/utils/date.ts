const intlLocale = (locale: string) => (locale === 'my' ? 'my-MM' : 'en-GB')

const TZ = 'Asia/Bangkok'

export function formatDate(iso: string | null | undefined, locale = 'en'): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleDateString(intlLocale(locale), {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    timeZone: TZ,
  })
}

export function formatDateTime(iso: string | null | undefined, locale = 'en'): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(intlLocale(locale), {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    timeZone: TZ,
  })
}

/** Milliseconds until the given ISO time (negative when past). */
export function msUntil(iso: string | null | undefined, now = Date.now()): number {
  if (!iso) return 0
  return new Date(iso).getTime() - now
}

export interface Countdown {
  days: number
  hours: number
  minutes: number
  seconds: number
  total: number
}

export function splitDuration(ms: number): Countdown {
  const total = Math.max(0, ms)
  const s = Math.floor(total / 1000)
  return {
    days: Math.floor(s / 86400),
    hours: Math.floor((s % 86400) / 3600),
    minutes: Math.floor((s % 3600) / 60),
    seconds: s % 60,
    total,
  }
}

const pad = (n: number) => String(n).padStart(2, '0')

/** "1d 02:03:04" / "02:03:04" / "03:04" */
export function formatCountdown(ms: number): string {
  const c = splitDuration(ms)
  const hms =
    c.hours > 0 || c.days > 0
      ? `${pad(c.hours)}:${pad(c.minutes)}:${pad(c.seconds)}`
      : `${pad(c.minutes)}:${pad(c.seconds)}`
  return c.days > 0 ? `${c.days}d ${hms}` : hms
}

/** ISO string → value for <input type="datetime-local"> in Bangkok time. */
export function toLocalInput(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const bkk = new Date(d.getTime() + 7 * 3600 * 1000)
  return bkk.toISOString().slice(0, 16)
}

/** <input type="datetime-local"> value (Bangkok time) → ISO UTC string. */
export function fromLocalInput(value: string): string {
  if (!value) return ''
  const d = new Date(`${value}:00+07:00`)
  return Number.isNaN(d.getTime()) ? '' : d.toISOString()
}

/** Today's date in Bangkok as YYYY-MM-DD. */
export function todayISODate(): string {
  return new Date(Date.now() + 7 * 3600 * 1000).toISOString().slice(0, 10)
}
