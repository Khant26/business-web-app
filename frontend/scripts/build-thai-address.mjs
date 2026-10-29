// Builds public/data/thai-address.json — a compact province → district → subdistrict tree.
//
// Source dataset: kongvut/thai-province-data (MIT)
//   https://github.com/kongvut/thai-province-data
//   https://raw.githubusercontent.com/kongvut/thai-province-data/master/api/latest/province_with_district_and_sub_district.json
//
// Usage:
//   node scripts/build-thai-address.mjs                 # downloads the source JSON
//   node scripts/build-thai-address.mjs path/to/file    # uses an already downloaded copy
//
// Output shape: [{ p_en, p_th, d: [{ en, th, s: [{ en, th, zip }] }] }]
// Province English names (p_en) are the canonical values sent to the API ("Bangkok", "Chiang Mai").

import { readFile, writeFile, mkdir } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const SOURCE_URL =
  'https://raw.githubusercontent.com/kongvut/thai-province-data/master/api/latest/province_with_district_and_sub_district.json'

const here = dirname(fileURLToPath(import.meta.url))
const outFile = resolve(here, '../public/data/thai-address.json')

async function load() {
  const local = process.argv[2]
  if (local) return JSON.parse(await readFile(local, 'utf8'))
  const res = await fetch(SOURCE_URL)
  if (!res.ok) throw new Error(`Download failed: ${res.status} ${res.statusText}`)
  return res.json()
}

const clean = (s) => String(s ?? '').trim()
const byEn = (a, b) => a.en.localeCompare(b.en, 'en')

const raw = await load()
const provinces = raw
  .filter((p) => !p.deleted_at)
  .map((p) => ({
    p_en: clean(p.name.en),
    p_th: clean(p.name.th),
    d: (p.districts ?? [])
      .filter((d) => !d.deleted_at)
      .map((d) => ({
        en: clean(d.name.en),
        th: clean(d.name.th),
        s: (d.sub_districts ?? [])
          .filter((s) => !s.deleted_at)
          .map((s) => ({
            en: clean(s.name.en),
            th: clean(s.name.th),
            zip: String(s.zip_code ?? ''),
          }))
          .sort(byEn),
      }))
      .sort(byEn),
  }))
  .sort((a, b) => a.p_en.localeCompare(b.p_en, 'en'))

await mkdir(dirname(outFile), { recursive: true })
await writeFile(outFile, JSON.stringify(provinces))

const districts = provinces.reduce((n, p) => n + p.d.length, 0)
const subs = provinces.reduce((n, p) => n + p.d.reduce((m, d) => m + d.s.length, 0), 0)
console.log(
  `Wrote ${outFile}: ${provinces.length} provinces, ${districts} districts, ${subs} subdistricts`,
)
