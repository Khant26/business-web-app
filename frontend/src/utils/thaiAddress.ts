export interface ThaiSubdistrict {
  en: string
  th: string
  zip: string
}
export interface ThaiDistrict {
  en: string
  th: string
  s: ThaiSubdistrict[]
}
export interface ThaiProvince {
  p_en: string
  p_th: string
  d: ThaiDistrict[]
}

let cache: Promise<ThaiProvince[]> | null = null

/** Lazily fetches public/data/thai-address.json once per session. */
export function loadThaiAddress(): Promise<ThaiProvince[]> {
  if (!cache) {
    cache = fetch('/data/thai-address.json')
      .then((r) => {
        if (!r.ok) throw new Error(`thai-address ${r.status}`)
        return r.json() as Promise<ThaiProvince[]>
      })
      .catch((e: unknown) => {
        cache = null
        throw e
      })
  }
  return cache
}
