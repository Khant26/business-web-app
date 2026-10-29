import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { i18n, loadAdminMessages } from '@/i18n'

declare module 'vue-router' {
  interface RouteMeta {
    /** i18n key for document.title */
    title?: string
    requiresAuth?: boolean
    requiresStaff?: boolean
    /** ★ admin-only */
    requiresAdmin?: boolean
    guestOnly?: boolean
    /** Do not scroll to top when only the query string changes. */
    keepScroll?: boolean
  }
}

const StoreLayout = () => import('@/layouts/StoreLayout.vue')
const AdminLayout = () => import('@/layouts/AdminLayout.vue')

const storeRoutes: RouteRecordRaw[] = [
  { path: '', name: 'home', component: () => import('@/pages/store/HomePage.vue') },
  {
    path: 'products',
    name: 'products',
    component: () => import('@/pages/store/ProductsPage.vue'),
    meta: { title: 'catalog.allProducts', keepScroll: true },
  },
  {
    path: 'products/:slug',
    name: 'product',
    component: () => import('@/pages/store/ProductDetailPage.vue'),
    props: true,
  },
  {
    path: 'categories',
    name: 'categories',
    component: () => import('@/pages/store/CategoriesPage.vue'),
    meta: { title: 'catalog.categoriesTitle' },
  },
  {
    path: 'categories/:slug',
    name: 'category',
    component: () => import('@/pages/store/CategoryPage.vue'),
    props: true,
    meta: { keepScroll: true },
  },
  {
    path: 'events',
    name: 'events',
    component: () => import('@/pages/store/EventsPage.vue'),
    meta: { title: 'catalog.eventsTitle' },
  },
  {
    path: 'events/:slug',
    name: 'event',
    component: () => import('@/pages/store/EventPage.vue'),
    props: true,
  },
  {
    path: 'cart',
    name: 'cart',
    component: () => import('@/pages/store/CartPage.vue'),
    meta: { title: 'cart.title' },
  },
  {
    path: 'login',
    name: 'login',
    component: () => import('@/pages/store/LoginPage.vue'),
    meta: { title: 'nav.login', guestOnly: true },
  },
  {
    path: 'checkout',
    name: 'checkout',
    component: () => import('@/pages/store/CheckoutPage.vue'),
    meta: { title: 'checkout.title', requiresAuth: true },
  },
  {
    path: 'account',
    component: () => import('@/pages/store/account/AccountLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'account-profile',
        component: () => import('@/pages/store/account/ProfilePage.vue'),
        meta: { title: 'account.profile.title' },
      },
      {
        path: 'orders',
        name: 'account-orders',
        component: () => import('@/pages/store/account/OrdersPage.vue'),
        meta: { title: 'account.orders.title', keepScroll: true },
      },
      {
        path: 'orders/:code',
        name: 'account-order',
        component: () => import('@/pages/store/account/OrderDetailPage.vue'),
        props: true,
      },
      {
        path: 'orders/:code/pay',
        name: 'account-pay',
        component: () => import('@/pages/store/account/PayPage.vue'),
        props: true,
        meta: { title: 'pay.title' },
      },
      {
        path: 'addresses',
        name: 'account-addresses',
        component: () => import('@/pages/store/account/AddressesPage.vue'),
        meta: { title: 'account.addresses.title' },
      },
      {
        path: 'wishlist',
        name: 'account-wishlist',
        component: () => import('@/pages/store/account/WishlistPage.vue'),
        meta: { title: 'nav.wishlist' },
      },
      {
        path: 'security',
        name: 'account-security',
        component: () => import('@/pages/store/account/SecurityPage.vue'),
        meta: { title: 'account.security.title' },
      },
    ],
  },
]

const adminRoutes: RouteRecordRaw[] = [
  {
    path: '',
    name: 'admin-dashboard',
    component: () => import('@/pages/admin/DashboardPage.vue'),
    meta: { title: 'admin.nav.dashboard' },
  },
  {
    path: 'orders',
    name: 'admin-orders',
    component: () => import('@/pages/admin/OrdersPage.vue'),
    meta: { title: 'admin.nav.orders', keepScroll: true },
  },
  {
    path: 'orders/:code',
    name: 'admin-order',
    component: () => import('@/pages/admin/OrderDetailPage.vue'),
    props: true,
  },
  {
    path: 'products',
    name: 'admin-products',
    component: () => import('@/pages/admin/ProductsPage.vue'),
    meta: { title: 'admin.nav.products', keepScroll: true },
  },
  {
    path: 'products/new',
    name: 'admin-product-new',
    component: () => import('@/pages/admin/ProductEditPage.vue'),
    meta: { title: 'admin.products.new' },
  },
  {
    path: 'products/:id(\\d+)',
    name: 'admin-product-edit',
    component: () => import('@/pages/admin/ProductEditPage.vue'),
    props: (r) => ({ id: Number(r.params.id) }),
  },
  {
    path: 'inventory',
    name: 'admin-inventory',
    component: () => import('@/pages/admin/InventoryPage.vue'),
    meta: { title: 'admin.nav.inventory', keepScroll: true },
  },
  {
    path: 'categories',
    name: 'admin-categories',
    component: () => import('@/pages/admin/CategoriesPage.vue'),
    meta: { title: 'admin.nav.categories' },
  },
  {
    path: 'events',
    name: 'admin-events',
    component: () => import('@/pages/admin/EventsPage.vue'),
    meta: { title: 'admin.nav.events' },
  },
  {
    path: 'coupons',
    name: 'admin-coupons',
    component: () => import('@/pages/admin/CouponsPage.vue'),
    meta: { title: 'admin.nav.coupons' },
  },
  {
    path: 'reviews',
    name: 'admin-reviews',
    component: () => import('@/pages/admin/ReviewsPage.vue'),
    meta: { title: 'admin.nav.reviews', keepScroll: true },
  },
  {
    path: 'shipping-zones',
    name: 'admin-zones',
    component: () => import('@/pages/admin/ShippingZonesPage.vue'),
    meta: { title: 'admin.nav.zones', requiresAdmin: true },
  },
  {
    path: 'settings',
    name: 'admin-settings',
    component: () => import('@/pages/admin/SettingsPage.vue'),
    meta: { title: 'admin.nav.settings', requiresAdmin: true },
  },
  {
    path: 'users',
    name: 'admin-users',
    component: () => import('@/pages/admin/UsersPage.vue'),
    meta: { title: 'admin.nav.users', requiresAdmin: true, keepScroll: true },
  },
  {
    path: 'audit-log',
    name: 'admin-audit',
    component: () => import('@/pages/admin/AuditLogPage.vue'),
    meta: { title: 'admin.nav.audit', requiresAdmin: true, keepScroll: true },
  },
]

export const routes: RouteRecordRaw[] = [
  {
    path: '/admin',
    component: AdminLayout,
    meta: { requiresAuth: true, requiresStaff: true },
    beforeEnter: () => loadAdminMessages(),
    children: adminRoutes,
  },
  {
    path: '/',
    component: StoreLayout,
    children: [
      ...storeRoutes,
      {
        path: ':pathMatch(.*)*',
        name: 'not-found',
        component: () => import('@/pages/store/NotFoundPage.vue'),
        meta: { title: 'notFound.title' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, saved) {
    if (saved) return saved
    if (to.hash) return { el: to.hash, top: 80 }
    if (to.path === from.path && to.meta.keepScroll) return false
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.init()

  const needsAuth = to.matched.some((r) => r.meta.requiresAuth)
  const needsStaff = to.matched.some((r) => r.meta.requiresStaff)
  const needsAdmin = to.matched.some((r) => r.meta.requiresAdmin)

  if (needsAuth && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (needsStaff && !auth.isStaff) {
    return { name: 'home' }
  }
  if (needsAdmin && !auth.isAdmin) {
    return { name: 'admin-dashboard' }
  }
  if (to.meta.guestOnly && auth.isLoggedIn) {
    const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : '/'
    return redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/'
  }
  return true
})

router.afterEach((to) => {
  const shop = i18n.global.t('app.name')
  const key = [...to.matched].reverse().find((r) => r.meta.title)?.meta.title
  if (key) document.title = `${i18n.global.t(key)} · ${shop}`
  else if (to.name === 'home') document.title = shop
})

export default router
