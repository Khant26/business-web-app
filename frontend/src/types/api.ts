// Types mirror docs/API.md exactly. Money is integer satang; times are RFC 3339 UTC strings.

export type Role = 'customer' | 'staff' | 'admin'

export interface User {
  id: number
  email: string
  name: string
  phone: string
  role: Role
  created_at: string
}

export interface Paged<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface ApiErrorBody {
  error: {
    code: string
    message: string
    fields?: Record<string, string>
  }
}

export interface Category {
  id: number
  slug: string
  name_en: string
  name_my: string
  image_url: string
  sort_order: number
  is_active: boolean
  product_count: number
}

export interface Variant {
  id: number
  sku: string
  name_en: string
  name_my: string
  price_satang: number
  compare_at_satang: number | null
  stock: number
  is_active: boolean
  sort_order: number
}

export interface AdminVariant extends Variant {
  low_stock_threshold: number
}

export interface CategoryRef {
  slug: string
  name_en: string
  name_my: string
}

export interface ProductCard {
  id: number
  slug: string
  name_en: string
  name_my: string
  image_url: string
  price_satang: number
  compare_at_satang: number | null
  rating_avg: number
  rating_count: number
  sold_count: number
  in_stock: boolean
  featured: boolean
  category: CategoryRef | null
}

export interface ProductImage {
  id: number
  url: string
  sort_order: number
}

export interface ProductDetail extends ProductCard {
  description_en: string
  description_my: string
  images: ProductImage[]
  variants: Variant[]
}

export interface ShopEvent {
  id: number
  slug: string
  title_en: string
  title_my: string
  description_en: string
  description_my: string
  banner_url: string
  starts_at: string
  ends_at: string
  is_active: boolean
  sort_order: number
}

export interface EventWithProducts extends ShopEvent {
  products: ProductCard[]
}

export interface ShippingZone {
  id: number
  name_en: string
  name_my: string
  fee_satang: number
  free_over_satang: number | null
  is_default: boolean
  sort_order: number
  provinces: string[]
}

export interface Address {
  id: number
  label: string
  recipient_name: string
  phone: string
  line1: string
  subdistrict: string
  district: string
  province: string
  postcode: string
  is_default: boolean
}

export type AddressInput = Omit<Address, 'id'>

export type ShippingAddress = Pick<
  Address,
  'recipient_name' | 'phone' | 'line1' | 'subdistrict' | 'district' | 'province' | 'postcode'
>

export const ORDER_STATUSES = [
  'pending_payment',
  'payment_review',
  'paid',
  'confirmed',
  'packing',
  'shipped',
  'delivered',
  'cancelled',
  'expired',
] as const
export type OrderStatus = (typeof ORDER_STATUSES)[number]

export type PaymentMethod = 'promptpay' | 'cod'

export interface OrderItem {
  id: number
  product_id: number | null
  product_slug: string
  variant_id: number | null
  product_name_en: string
  product_name_my: string
  variant_name_en: string
  variant_name_my: string
  sku: string
  image_url: string
  unit_price_satang: number
  quantity: number
  line_total_satang: number
  can_review: boolean
}

export interface OrderSummary {
  code: string
  status: OrderStatus
  payment_method: PaymentMethod
  total_satang: number
  item_count: number
  created_at: string
  first_item_image: string
}

export type SlipStatus = 'pending' | 'approved' | 'rejected'

export interface Slip {
  id: number
  status: SlipStatus
  note: string
  uploaded_at: string
  reviewed_at: string | null
}

export interface AdminSlip extends Slip {
  image_url: string
}

export interface OrderHistoryEntry {
  status: OrderStatus
  note: string
  created_at: string
}

export interface Order extends OrderSummary {
  subtotal_satang: number
  discount_satang: number
  shipping_satang: number
  vat_rate_bp: number
  vat_satang: number
  coupon_code: string
  shipping_address: ShippingAddress
  email: string
  note: string
  courier: string
  tracking_number: string
  tracking_url: string
  payment_due_at: string | null
  paid_at: string | null
  shipped_at: string | null
  delivered_at: string | null
  cancelled_at: string | null
  cancel_reason: string
  items: OrderItem[]
  history: OrderHistoryEntry[]
  slips: Slip[]
  can_cancel: boolean
  can_upload_slip: boolean
}

export interface QuoteLine {
  variant_id: number
  product_slug: string
  product_name_en: string
  product_name_my: string
  variant_name_en: string
  variant_name_my: string
  image_url: string
  unit_price_satang: number
  quantity: number
  line_total_satang: number
  available: number
}

export type ProblemCode = 'out_of_stock' | 'insufficient_stock' | 'unavailable'

export interface CartProblem {
  variant_id: number
  code: ProblemCode
  available: number
}

export interface Quote {
  lines: QuoteLine[]
  subtotal_satang: number
  discount_satang: number
  shipping_satang: number
  vat_rate_bp: number
  vat_satang: number
  total_satang: number
  coupon: { code: string; valid: boolean; message_code: string } | null
  zone: { name_en: string; name_my: string; free_over_satang: number | null } | null
  problems: CartProblem[]
  cod_allowed: boolean
}

export interface CartItemInput {
  variant_id: number
  quantity: number
}

export interface CartValidation {
  lines: QuoteLine[]
  problems: CartProblem[]
}

export interface Review {
  id: number
  rating: number
  comment: string
  user_name: string
  created_at: string
}

export interface PublicSettings {
  shop_name: string
  vat_rate_bp: number
  payment_window_minutes: number
  cod_max_satang: number
  support_email: string
  support_phone: string
}

export interface HomeData {
  events: ShopEvent[]
  categories: Category[]
  popular: ProductCard[]
  featured: ProductCard[]
  new_arrivals: ProductCard[]
}

export interface Province {
  name_en: string
  name_th: string
}

export type ProductSort = 'popular' | 'newest' | 'price_asc' | 'price_desc' | 'rating'

export interface ProductQuery {
  q?: string
  category?: string
  event?: string
  sort?: ProductSort
  min_price?: number
  max_price?: number
  in_stock?: boolean
  page?: number
  page_size?: number
}

export interface Session {
  id: string
  ip: string
  user_agent: string
  created_at: string
  last_seen_at: string
  current: boolean
}

export interface CheckoutQuoteInput {
  items: CartItemInput[]
  province: string
  coupon_code?: string
}

export interface PlaceOrderInput {
  items: CartItemInput[]
  address_id?: number
  address?: AddressInput
  payment_method: PaymentMethod
  coupon_code?: string
  note?: string
  save_address?: boolean
}

export interface PromptPayInfo {
  payload: string
  amount_satang: number
  promptpay_name: string
  due_at: string
}

// ---------- Admin ----------

export interface AdminOrderSummary extends OrderSummary {
  customer_email: string
  recipient_name: string
  has_pending_slip: boolean
}

export interface AdminOrder extends Omit<Order, 'slips'> {
  customer: User
  slips: AdminSlip[]
}

export interface InventoryRow {
  variant_id: number
  sku: string
  product_id: number
  product_name_en: string
  variant_name_en: string
  stock: number
  low_stock_threshold: number
  is_low: boolean
  image_url: string
}

export interface DashboardData {
  today_sales_satang: number
  today_orders: number
  pending_slips: number
  to_pack: number
  low_stock_count: number
  sales_7d: { date: string; total_satang: number; orders: number }[]
  recent_orders: AdminOrderSummary[]
  low_stock: InventoryRow[]
}

export type ProductStatus = 'draft' | 'active' | 'archived'

export interface AdminProductRow {
  id: number
  slug: string
  name_en: string
  name_my: string
  status: ProductStatus
  featured: boolean
  category_name_en: string
  image_url: string
  variant_count: number
  total_stock: number
  min_price_satang: number
  updated_at: string
}

export interface AdminProduct extends Omit<ProductDetail, 'variants'> {
  variants: AdminVariant[]
  status: ProductStatus
  category_id: number | null
  event_ids: number[]
}

export interface ProductInput {
  slug: string
  category_id: number | null
  name_en: string
  name_my: string
  description_en: string
  description_my: string
  status: ProductStatus
  featured: boolean
  event_ids?: number[]
}

export interface VariantInput {
  sku: string
  name_en: string
  name_my: string
  price_satang: number
  compare_at_satang: number | null
  stock?: number
  low_stock_threshold: number
  is_active: boolean
  sort_order: number
}

export type StockReason = 'restock' | 'adjust' | 'damage'
export type MovementReason = 'order' | 'cancel' | 'expire' | StockReason

export interface InventoryMovement {
  id: number
  variant_id: number
  sku: string
  product_name_en: string
  change: number
  reason: MovementReason
  note: string
  order_code: string
  user_email: string
  stock_after: number
  created_at: string
}

export interface CategoryInput {
  slug: string
  name_en: string
  name_my: string
  sort_order: number
  is_active: boolean
}

export interface AdminEvent extends ShopEvent {
  product_ids: number[]
}

export interface EventInput {
  slug: string
  title_en: string
  title_my: string
  description_en: string
  description_my: string
  starts_at: string
  ends_at: string
  is_active: boolean
  sort_order: number
  product_ids: number[]
}

export type CouponType = 'percent' | 'fixed'

export interface Coupon {
  id: number
  code: string
  type: CouponType
  value: number
  min_order_satang: number
  max_discount_satang: number | null
  starts_at: string | null
  ends_at: string | null
  usage_limit: number | null
  per_user_limit: number
  used_count: number
  is_active: boolean
  created_at: string
}

export type CouponInput = Omit<Coupon, 'id' | 'used_count' | 'created_at'>

export type ReviewStatus = 'pending' | 'approved' | 'rejected'

export interface AdminReview {
  id: number
  product_id: number
  product_name_en: string
  user_email: string
  rating: number
  comment: string
  status: ReviewStatus
  created_at: string
}

export type ShippingZoneInput = Omit<ShippingZone, 'id'>

export interface AdminSettings {
  shop_name: string
  promptpay_id: string
  promptpay_name: string
  vat_rate_bp: number
  payment_window_minutes: number
  cod_max_satang: number
  support_email: string
  support_phone: string
}

export interface AdminUser extends User {
  is_active: boolean
  order_count: number
}

export interface AuditLog {
  id: number
  user_email: string
  action: string
  entity: string
  entity_id: string
  details: unknown
  ip: string
  created_at: string
}
