import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError, buildQuery, onUnauthorized } from '../client'

function mockFetch(status: number, body?: unknown, headers: Record<string, string> = {}) {
  const fn = vi.fn(
    async (_url: string, _init?: RequestInit) =>
      new Response(body === undefined ? null : JSON.stringify(body), {
        status,
        headers: { 'Content-Type': 'application/json', ...headers },
      }),
  )
  vi.stubGlobal('fetch', fn)
  return fn
}

const headersOf = (init: RequestInit | undefined) => (init?.headers ?? {}) as Record<string, string>

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('api client', () => {
  it('sends credentials and no CSRF header on GET', async () => {
    const fetch = mockFetch(200, { ok: true })
    await api.get('/home', { q: 'tea', page: 2, empty: '', in_stock: true, off: false })
    const [url, init] = fetch.mock.calls[0]!
    expect(url).toBe('/api/v1/home?q=tea&page=2&in_stock=1')
    expect(init?.credentials).toBe('include')
    expect(headersOf(init)['X-Requested-With']).toBeUndefined()
  })

  it('adds X-Requested-With: sawyun and a JSON body on non-GET', async () => {
    const fetch = mockFetch(200, { id: 1 })
    await api.post('/orders', { a: 1 })
    await api.put('/x', {})
    await api.patch('/me', {})
    await api.del('/y')
    expect(fetch).toHaveBeenCalledTimes(4)
    for (const [, init] of fetch.mock.calls) {
      expect(headersOf(init)['X-Requested-With']).toBe('sawyun')
      expect(init?.credentials).toBe('include')
    }
    const first = fetch.mock.calls[0]![1]
    expect(headersOf(first)['Content-Type']).toBe('application/json')
    expect(first?.body).toBe('{"a":1}')
    expect(fetch.mock.calls[3]![1]?.method).toBe('DELETE')
  })

  it('sends multipart without forcing Content-Type', async () => {
    const fetch = mockFetch(200, {})
    await api.upload('/me/orders/SY-1/slip', new Blob(['x'], { type: 'image/png' }))
    const init = fetch.mock.calls[0]![1]
    expect(init?.body).toBeInstanceOf(FormData)
    expect((init?.body as FormData).get('file')).toBeTruthy()
    expect(headersOf(init)['Content-Type']).toBeUndefined()
    expect(headersOf(init)['X-Requested-With']).toBe('sawyun')
  })

  it('returns undefined for 204', async () => {
    mockFetch(204)
    await expect(api.post('/auth/logout')).resolves.toBeUndefined()
  })

  it('parses API error bodies', async () => {
    mockFetch(422, { error: { code: 'validation', message: 'Bad', fields: { phone: 'invalid' } } })
    const err = await api.post('/me/addresses', {}).catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect(err).toMatchObject({ code: 'validation', status: 422, fields: { phone: 'invalid' } })
  })

  it('reads Retry-After on 429', async () => {
    mockFetch(
      429,
      { error: { code: 'rate_limited', message: 'slow down' } },
      { 'Retry-After': '42' },
    )
    const err = (await api.post('/auth/otp/request', {}).catch((e: unknown) => e)) as ApiError
    expect(err.code).toBe('rate_limited')
    expect(err.retryAfter).toBe(42)
  })

  it('falls back to status-derived codes for non-JSON errors', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('<html>oops</html>', { status: 404 })),
    )
    const err = (await api.get('/nope').catch((e: unknown) => e)) as ApiError
    expect(err.code).toBe('not_found')
    expect(err.status).toBe(404)
  })

  it('maps network failures to code "network"', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Promise.reject(new TypeError('Failed to fetch'))),
    )
    const err = (await api.get('/home').catch((e: unknown) => e)) as ApiError
    expect(err.code).toBe('network')
  })

  it('notifies unauthorized handlers on 401', async () => {
    mockFetch(401, { error: { code: 'unauthorized', message: 'no' } })
    const handler = vi.fn()
    const off = onUnauthorized(handler)
    await api.get('/me/orders').catch(() => undefined)
    expect(handler).toHaveBeenCalledWith('/me/orders')
    off()
  })

  it('buildQuery skips empty values', () => {
    expect(buildQuery({ a: undefined, b: null, c: '', d: 0 })).toBe('?d=0')
    expect(buildQuery({})).toBe('')
  })
})
