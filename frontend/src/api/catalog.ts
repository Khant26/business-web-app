import { api, seg } from './client'
import type {
  Category,
  CartItemInput,
  CartValidation,
  EventWithProducts,
  HomeData,
  Paged,
  ProductCard,
  ProductDetail,
  ProductQuery,
  Province,
  PublicSettings,
  Review,
  ShippingZone,
  ShopEvent,
} from '@/types/api'

export const catalogApi = {
  settings: () => api.get<PublicSettings>('/settings/public'),
  home: () => api.get<HomeData>('/home'),
  categories: () => api.get<Category[]>('/categories'),
  category: (slug: string) => api.get<Category>(`/categories/${seg(slug)}`),
  products: (q: ProductQuery = {}, signal?: AbortSignal) =>
    api.get<Paged<ProductCard>>('/products', { ...q }, signal),
  product: (slug: string) => api.get<ProductDetail>(`/products/${seg(slug)}`),
  reviews: (slug: string, page = 1, page_size = 10) =>
    api.get<Paged<Review>>(`/products/${seg(slug)}/reviews`, { page, page_size }),
  createReview: (slug: string, body: { rating: number; comment: string }) =>
    api.post<{ id: number }>(`/products/${seg(slug)}/reviews`, body),
  events: () => api.get<ShopEvent[]>('/events'),
  event: (slug: string) => api.get<EventWithProducts>(`/events/${seg(slug)}`),
  provinces: () => api.get<Province[]>('/provinces'),
  shippingZones: () => api.get<ShippingZone[]>('/shipping/zones'),
  validateCart: (items: CartItemInput[]) => api.post<CartValidation>('/cart/validate', { items }),
}
