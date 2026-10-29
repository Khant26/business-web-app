import { api, seg } from './client'
import type {
  CheckoutQuoteInput,
  Order,
  OrderStatus,
  OrderSummary,
  Paged,
  PlaceOrderInput,
  PromptPayInfo,
  Quote,
} from '@/types/api'

export const checkoutApi = {
  quote: (body: CheckoutQuoteInput) => api.post<Quote>('/checkout/quote', body),
  placeOrder: (body: PlaceOrderInput) => api.post<Order>('/orders', body),
}

export const ordersApi = {
  list: (q: { status?: OrderStatus; page?: number; page_size?: number } = {}) =>
    api.get<Paged<OrderSummary>>('/me/orders', { ...q }),
  get: (code: string) => api.get<Order>(`/me/orders/${seg(code)}`),
  cancel: (code: string, reason?: string) =>
    api.post<Order>(`/me/orders/${seg(code)}/cancel`, reason ? { reason } : {}),
  promptpay: (code: string) => api.get<PromptPayInfo>(`/me/orders/${seg(code)}/promptpay`),
  uploadSlip: (code: string, file: File) => api.upload<Order>(`/me/orders/${seg(code)}/slip`, file),
}
