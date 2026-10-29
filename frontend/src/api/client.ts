import type { ApiErrorBody } from '@/types/api'

/**
 * Where the API lives. Empty (default) = same origin, which is what the Vite dev proxy and the
 * all-in-one Docker image use. Set VITE_API_ORIGIN (e.g. https://api-shop.example.com) when the
 * frontend is hosted separately, e.g. on Vercel.
 */
export const API_ORIGIN = (import.meta.env.VITE_API_ORIGIN ?? '').replace(/\/+$/, '')
export const API_BASE = `${API_ORIGIN}/api/v1`

const ASSET_KEYS = new Set(['url', 'image_url', 'banner_url', 'first_item_image'])

/** Rewrites uploaded-file paths in API responses (in place) so <img> tags load them from the API. */
function absolutizeAssets(value: unknown): void {
  if (!API_ORIGIN || typeof value !== 'object' || value === null) return
  if (Array.isArray(value)) {
    value.forEach(absolutizeAssets)
    return
  }
  const obj = value as Record<string, unknown>
  for (const [k, v] of Object.entries(obj)) {
    if (typeof v === 'string') {
      if (ASSET_KEYS.has(k) && v.startsWith('/uploads/')) obj[k] = API_ORIGIN + v
    } else {
      absolutizeAssets(v)
    }
  }
}
export const CSRF_HEADER = 'X-Requested-With'
export const CSRF_VALUE = 'sawyun'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
export type QueryValue = string | number | boolean | null | undefined
export type Query = Record<string, QueryValue>

/** Error thrown for every non-2xx response (and for network failures, with code `network`). */
export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly fields: Record<string, string>
  readonly retryAfter: number | null

  constructor(
    code: string,
    message: string,
    status: number,
    fields: Record<string, string> = {},
    retryAfter: number | null = null,
  ) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.fields = fields
    this.retryAfter = retryAfter
  }
}

export function isApiError(e: unknown): e is ApiError {
  return e instanceof ApiError
}

type UnauthorizedHandler = (path: string) => void
const unauthorizedHandlers = new Set<UnauthorizedHandler>()

/** Registers a callback fired whenever a request returns 401. Returns an unsubscribe function. */
export function onUnauthorized(handler: UnauthorizedHandler): () => void {
  unauthorizedHandlers.add(handler)
  return () => unauthorizedHandlers.delete(handler)
}

export function buildQuery(query?: Query): string {
  if (!query) return ''
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === '' || value === false) continue
    params.set(key, value === true ? '1' : String(value))
  }
  const s = params.toString()
  return s ? `?${s}` : ''
}

function isErrorBody(v: unknown): v is ApiErrorBody {
  if (typeof v !== 'object' || v === null || !('error' in v)) return false
  const err = (v as { error: unknown }).error
  return typeof err === 'object' && err !== null && 'code' in err
}

function parseRetryAfter(value: string | null): number | null {
  if (!value) return null
  const secs = Number(value)
  if (Number.isFinite(secs)) return Math.max(0, Math.ceil(secs))
  const date = Date.parse(value)
  if (Number.isNaN(date)) return null
  return Math.max(0, Math.ceil((date - Date.now()) / 1000))
}

const STATUS_CODES: Record<number, string> = {
  400: 'bad_request',
  401: 'unauthorized',
  403: 'forbidden',
  404: 'not_found',
  409: 'conflict',
  413: 'too_large',
  422: 'validation',
  429: 'rate_limited',
}

export interface RequestOptions {
  query?: Query
  body?: unknown
  signal?: AbortSignal
  /** Do not notify unauthorized handlers on 401 (e.g. for /auth/me probing). */
  silent401?: boolean
}

export async function request<T>(
  method: HttpMethod,
  path: string,
  opts: RequestOptions = {},
): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' }
  if (method !== 'GET') headers[CSRF_HEADER] = CSRF_VALUE

  let body: BodyInit | undefined
  if (opts.body instanceof FormData) {
    body = opts.body
  } else if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(opts.body)
  }

  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}${buildQuery(opts.query)}`, {
      method,
      headers,
      body,
      credentials: 'include',
      signal: opts.signal,
    })
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw e
    throw new ApiError('network', 'Network error', 0)
  }

  if (res.status === 204) return undefined as T

  const text = await res.text()
  let data: unknown = undefined
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = undefined
    }
  }

  if (!res.ok) {
    const retryAfter = parseRetryAfter(res.headers.get('Retry-After'))
    let err: ApiError
    if (isErrorBody(data)) {
      err = new ApiError(
        data.error.code,
        data.error.message ?? '',
        res.status,
        data.error.fields ?? {},
        retryAfter,
      )
    } else {
      err = new ApiError(
        STATUS_CODES[res.status] ?? 'internal',
        res.statusText,
        res.status,
        {},
        retryAfter,
      )
    }
    if (res.status === 401 && !opts.silent401) {
      unauthorizedHandlers.forEach((h) => h(path))
    }
    throw err
  }

  absolutizeAssets(data)
  return data as T
}

export const api = {
  get: <T>(path: string, query?: Query, signal?: AbortSignal) =>
    request<T>('GET', path, { query, signal }),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, { body }),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, { body }),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, { body }),
  del: <T = void>(path: string) => request<T>('DELETE', path),
  upload: <T>(path: string, file: File | Blob, method: 'POST' | 'PUT' = 'POST') => {
    const fd = new FormData()
    fd.append('file', file)
    return request<T>(method, path, { body: fd })
  },
}

/** Encode a path segment (slugs, codes). */
export const seg = (v: string | number) => encodeURIComponent(String(v))
