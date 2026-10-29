import { api, seg } from '../client'
import type {
  AdminEvent,
  AdminOrder,
  AdminOrderSummary,
  AdminProduct,
  AdminProductRow,
  AdminReview,
  AdminSettings,
  AdminUser,
  AdminVariant,
  AuditLog,
  Category,
  CategoryInput,
  Coupon,
  CouponInput,
  DashboardData,
  EventInput,
  InventoryMovement,
  InventoryRow,
  OrderStatus,
  Paged,
  PaymentMethod,
  ProductImage,
  ProductInput,
  ProductStatus,
  ReviewStatus,
  Role,
  ShippingZone,
  ShippingZoneInput,
  StockReason,
  User,
  VariantInput,
} from '@/types/api'

export interface AdminOrderQuery {
  status?: OrderStatus
  q?: string
  payment_method?: PaymentMethod
  from?: string
  to?: string
  page?: number
  page_size?: number
}

export interface AdminProductQuery {
  q?: string
  status?: ProductStatus
  category?: number
  page?: number
  page_size?: number
}

export const adminApi = {
  dashboard: () => api.get<DashboardData>('/admin/dashboard'),

  // Orders
  orders: (q: AdminOrderQuery = {}) => api.get<Paged<AdminOrderSummary>>('/admin/orders', { ...q }),
  order: (code: string) => api.get<AdminOrder>(`/admin/orders/${seg(code)}`),
  setOrderStatus: (code: string, status: OrderStatus, note?: string) =>
    api.post<AdminOrder>(`/admin/orders/${seg(code)}/status`, { status, note }),
  setTracking: (
    code: string,
    body: { courier: string; tracking_number: string; tracking_url?: string },
  ) => api.put<AdminOrder>(`/admin/orders/${seg(code)}/tracking`, body),
  approveSlip: (id: number, note?: string) =>
    api.post<AdminOrder>(`/admin/slips/${id}/approve`, { note }),
  rejectSlip: (id: number, note: string) =>
    api.post<AdminOrder>(`/admin/slips/${id}/reject`, { note }),
  slipFileUrl: (id: number) => `/api/v1/admin/slips/${id}/file`,

  // Products
  products: (q: AdminProductQuery = {}) =>
    api.get<Paged<AdminProductRow>>('/admin/products', { ...q }),
  product: (id: number) => api.get<AdminProduct>(`/admin/products/${id}`),
  createProduct: (body: ProductInput) => api.post<AdminProduct>('/admin/products', body),
  updateProduct: (id: number, body: ProductInput) =>
    api.put<AdminProduct>(`/admin/products/${id}`, body),
  archiveProduct: (id: number) => api.del(`/admin/products/${id}`),
  uploadProductImage: (id: number, file: File) =>
    api.upload<ProductImage>(`/admin/products/${id}/images`, file),
  reorderProductImages: (id: number, image_ids: number[]) =>
    api.put<void>(`/admin/products/${id}/images/order`, { image_ids }),
  deleteProductImage: (id: number, imageId: number) =>
    api.del(`/admin/products/${id}/images/${imageId}`),
  createVariant: (productId: number, body: VariantInput) =>
    api.post<AdminVariant>(`/admin/products/${productId}/variants`, body),
  updateVariant: (id: number, body: Omit<VariantInput, 'stock'>) =>
    api.put<AdminVariant>(`/admin/variants/${id}`, body),
  deleteVariant: (id: number) => api.del(`/admin/variants/${id}`),

  // Inventory
  inventory: (q: { q?: string; low?: boolean; page?: number; page_size?: number } = {}) =>
    api.get<Paged<InventoryRow>>('/admin/inventory', { ...q }),
  adjustStock: (variantId: number, body: { change: number; reason: StockReason; note?: string }) =>
    api.post<InventoryRow>(`/admin/inventory/${variantId}/adjust`, body),
  movements: (q: { variant_id?: number; page?: number; page_size?: number } = {}) =>
    api.get<Paged<InventoryMovement>>('/admin/inventory/movements', { ...q }),

  // Categories
  categories: () => api.get<Category[]>('/admin/categories'),
  createCategory: (body: CategoryInput) => api.post<Category>('/admin/categories', body),
  updateCategory: (id: number, body: CategoryInput) =>
    api.put<Category>(`/admin/categories/${id}`, body),
  deleteCategory: (id: number) => api.del(`/admin/categories/${id}`),
  uploadCategoryImage: (id: number, file: File) =>
    api.upload<Category>(`/admin/categories/${id}/image`, file),

  // Events
  events: () => api.get<AdminEvent[]>('/admin/events'),
  createEvent: (body: EventInput) => api.post<AdminEvent>('/admin/events', body),
  updateEvent: (id: number, body: EventInput) => api.put<AdminEvent>(`/admin/events/${id}`, body),
  deleteEvent: (id: number) => api.del(`/admin/events/${id}`),
  uploadEventBanner: (id: number, file: File) =>
    api.upload<AdminEvent>(`/admin/events/${id}/banner`, file),

  // Coupons
  coupons: () => api.get<Coupon[]>('/admin/coupons'),
  createCoupon: (body: CouponInput) => api.post<Coupon>('/admin/coupons', body),
  updateCoupon: (id: number, body: CouponInput) => api.put<Coupon>(`/admin/coupons/${id}`, body),
  deleteCoupon: (id: number) => api.del(`/admin/coupons/${id}`),

  // Reviews
  reviews: (q: { status?: ReviewStatus; page?: number; page_size?: number } = {}) =>
    api.get<Paged<AdminReview>>('/admin/reviews', { ...q }),
  approveReview: (id: number) => api.post<void>(`/admin/reviews/${id}/approve`),
  rejectReview: (id: number) => api.post<void>(`/admin/reviews/${id}/reject`),

  // ★ Shipping zones (admin only)
  zones: () => api.get<ShippingZone[]>('/admin/shipping-zones'),
  createZone: (body: ShippingZoneInput) => api.post<ShippingZone>('/admin/shipping-zones', body),
  updateZone: (id: number, body: ShippingZoneInput) =>
    api.put<ShippingZone>(`/admin/shipping-zones/${id}`, body),
  deleteZone: (id: number) => api.del(`/admin/shipping-zones/${id}`),

  // ★ Settings
  settings: () => api.get<AdminSettings>('/admin/settings'),
  updateSettings: (body: AdminSettings) => api.put<AdminSettings>('/admin/settings', body),

  // ★ Users
  users: (q: { role?: Role; q?: string; page?: number; page_size?: number } = {}) =>
    api.get<Paged<AdminUser>>('/admin/users', { ...q }),
  createUser: (body: { email: string; name: string; role: 'staff' | 'admin' }) =>
    api.post<User>('/admin/users', body),
  updateUser: (id: number, body: { role?: Role; is_active?: boolean }) =>
    api.patch<User>(`/admin/users/${id}`, body),

  // ★ Audit log
  auditLogs: (q: { page?: number; page_size?: number } = {}) =>
    api.get<Paged<AuditLog>>('/admin/audit-logs', { ...q }),
}
