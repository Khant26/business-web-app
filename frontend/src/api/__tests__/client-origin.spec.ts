import { afterEach, describe, expect, it, vi } from 'vitest'

afterEach(() => {
  vi.unstubAllEnvs()
  vi.unstubAllGlobals()
  vi.resetModules()
})

describe('api client with a separately hosted API (VITE_API_ORIGIN)', () => {
  it('calls the API origin and makes uploaded-file paths absolute', async () => {
    vi.stubEnv('VITE_API_ORIGIN', 'https://api-shop.example.com/')
    const body = {
      items: [{ image_url: '/uploads/p/a.jpg', category: { image_url: '' } }],
      banner_url: 'https://cdn.example.com/b.jpg',
      images: [{ url: '/uploads/p/b.webp' }],
      description_en: '/uploads/not-an-asset-key',
    }
    const fetch = vi.fn(async (_url: string) => new Response(JSON.stringify(body), { status: 200 }))
    vi.stubGlobal('fetch', fetch)

    const { api, API_BASE } = await import('../client')
    const data = await api.get<typeof body>('/products')

    expect(API_BASE).toBe('https://api-shop.example.com/api/v1')
    expect(fetch.mock.calls[0]![0]).toBe('https://api-shop.example.com/api/v1/products')
    expect(data.items[0]!.image_url).toBe('https://api-shop.example.com/uploads/p/a.jpg')
    expect(data.items[0]!.category.image_url).toBe('')
    expect(data.images[0]!.url).toBe('https://api-shop.example.com/uploads/p/b.webp')
    expect(data.banner_url).toBe('https://cdn.example.com/b.jpg')
    expect(data.description_en).toBe('/uploads/not-an-asset-key')
  })
})
