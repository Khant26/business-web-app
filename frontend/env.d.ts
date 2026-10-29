/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** API origin when the frontend is hosted separately, e.g. https://api-shop.example.com */
  readonly VITE_API_ORIGIN?: string
}
