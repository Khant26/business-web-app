import { api, seg } from './client'
import type { Address, AddressInput, ProductCard, Session, User } from '@/types/api'

export const accountApi = {
  updateProfile: (body: { name: string; phone: string }) => api.patch<{ user: User }>('/me', body),
  sessions: () => api.get<Session[]>('/me/sessions'),
  revokeSession: (id: string) => api.del(`/me/sessions/${seg(id)}`),
  revokeOthers: () => api.post<void>('/me/sessions/revoke-others'),
  addresses: () => api.get<Address[]>('/me/addresses'),
  createAddress: (body: AddressInput) => api.post<Address>('/me/addresses', body),
  updateAddress: (id: number, body: AddressInput) => api.put<Address>(`/me/addresses/${id}`, body),
  deleteAddress: (id: number) => api.del(`/me/addresses/${id}`),
  wishlist: () => api.get<ProductCard[]>('/me/wishlist'),
  wishlistIds: () => api.get<number[]>('/me/wishlist/ids'),
  addWishlist: (productId: number) => api.put<void>(`/me/wishlist/${productId}`),
  removeWishlist: (productId: number) => api.del(`/me/wishlist/${productId}`),
}
