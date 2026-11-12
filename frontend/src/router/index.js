import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import PublicLayout from '../layouts/PublicLayout.vue'
import HomeView from '../views/public/HomeView.vue'

// Khu nội bộ: lazy-load để bundle công khai không chứa code nội bộ.
const InternalLayout = () => import('../layouts/InternalLayout.vue')
const v = {
  dashboard: () => import('../views/internal/DashboardView.vue'),
  products: () => import('../views/internal/ProductsView.vue'),
  priceList: () => import('../views/internal/PriceListView.vue'),
  customers: () => import('../views/internal/CustomersView.vue'),
  customer: () => import('../views/internal/CustomerDetailView.vue'),
  orders: () => import('../views/internal/OrdersView.vue'),
  orderForm: () => import('../views/internal/OrderFormView.vue'),
  order: () => import('../views/internal/OrderDetailView.vue'),
  changePassword: () => import('../views/internal/ChangePasswordView.vue'),
  productEdit: () => import('../views/internal/admin/ProductEditView.vue'),
  promotions: () => import('../views/internal/admin/PromotionsAdminView.vue'),
  promotionEdit: () => import('../views/internal/admin/PromotionEditView.vue'),
  company: () => import('../views/internal/admin/CompanyEditView.vue'),
  categories: () => import('../views/internal/admin/CategoriesView.vue'),
  users: () => import('../views/internal/admin/UsersView.vue'),
}
const admin = { requiresAuth: true, role: 'admin' }

const routes = [
  {
    path: '/',
    component: PublicLayout,
    children: [
      { path: '', name: 'home', component: HomeView, meta: { title: 'Đèn đánh cá, đèn công trình chính hãng' } },
      { path: 'danh-muc/:slug', name: 'category', component: () => import('../views/public/CategoryView.vue'), meta: { title: 'Danh mục sản phẩm' } },
      { path: 'bao-hanh', name: 'warranty', component: () => import('../views/public/WarrantyView.vue'), meta: { title: 'Tra cứu bảo hành' } },
    ],
  },
  { path: '/dang-nhap', name: 'login', component: () => import('../views/public/LoginView.vue'), meta: { title: 'Đăng nhập nội bộ', noindex: true, guestOnly: true } },
  {
    path: '/noi-bo',
    component: InternalLayout,
    meta: { requiresAuth: true, noindex: true },
    children: [
      { path: '', name: 'dashboard', component: v.dashboard, meta: { title: 'Tổng quan' } },
      { path: 'san-pham', name: 'products', component: v.products, meta: { title: 'Sản phẩm' } },
      { path: 'san-pham/moi', name: 'product-new', component: v.productEdit, meta: { ...admin, title: 'Thêm sản phẩm' } },
      { path: 'san-pham/:id(\\d+)', name: 'product-edit', component: v.productEdit, meta: { ...admin, title: 'Sửa sản phẩm' } },
      { path: 'bang-gia', name: 'price-list', component: v.priceList, meta: { title: 'Khuyến mãi & bảng giá' } },
      { path: 'khach-hang', name: 'customers', component: v.customers, meta: { title: 'Khách hàng' } },
      { path: 'khach-hang/:id(\\d+)', name: 'customer', component: v.customer, meta: { title: 'Khách hàng' } },
      { path: 'don-hang', name: 'orders', component: v.orders, meta: { title: 'Đơn hàng' } },
      { path: 'don-hang/moi', name: 'order-new', component: v.orderForm, meta: { title: 'Nhập đơn hàng' } },
      { path: 'don-hang/:id(\\d+)', name: 'order', component: v.order, meta: { title: 'Chi tiết đơn hàng' } },
      { path: 'don-hang/:id(\\d+)/sua', name: 'order-edit', component: v.orderForm, meta: { ...admin, title: 'Sửa đơn hàng' } },
      { path: 'khuyen-mai', name: 'promotions', component: v.promotions, meta: { ...admin, title: 'Quản lý khuyến mãi' } },
      { path: 'khuyen-mai/moi', name: 'promotion-new', component: v.promotionEdit, meta: { ...admin, title: 'Thêm khuyến mãi' } },
      { path: 'khuyen-mai/:id(\\d+)', name: 'promotion-edit', component: v.promotionEdit, meta: { ...admin, title: 'Sửa khuyến mãi' } },
      { path: 'danh-muc', name: 'categories', component: v.categories, meta: { ...admin, title: 'Danh mục sản phẩm' } },
      { path: 'cong-ty', name: 'company', component: v.company, meta: { ...admin, title: 'Thông tin công ty' } },
      { path: 'tai-khoan', name: 'users', component: v.users, meta: { ...admin, title: 'Tài khoản' } },
      { path: 'doi-mat-khau', name: 'change-password', component: v.changePassword, meta: { title: 'Đổi mật khẩu' } },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    component: PublicLayout,
    children: [{ path: '', name: 'not-found', component: () => import('../views/public/NotFoundView.vue'), meta: { title: 'Không tìm thấy trang' } }],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, saved) {
    if (saved) return saved
    if (to.hash) return { el: to.hash, behavior: 'smooth', top: 72 }
    if (to.path === from.path) return false // chỉ đổi query (?sp=) → giữ vị trí
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.matched.some((r) => r.meta.requiresAuth)) {
    if (!auth.isLoggedIn) return { name: 'login', query: { redirect: to.fullPath } }
    await auth.ensureUser()
    if (!auth.isLoggedIn) return { name: 'login', query: { redirect: to.fullPath } }
    if (to.meta.role === 'admin' && !auth.isAdmin) {
      useToastStore().error('Bạn không có quyền truy cập trang này')
      return { name: 'dashboard' }
    }
  }
  if (to.meta.guestOnly && auth.isLoggedIn) {
    await auth.ensureUser()
    if (auth.isLoggedIn) return { name: 'dashboard' }
  }
  return true
})

router.afterEach((to) => {
  document.title = 'Đông Hải — ' + (to.meta.title || 'Đông Hải')
  let robots = document.querySelector('meta[name="robots"]')
  if (to.matched.some((r) => r.meta.noindex)) {
    if (!robots) {
      robots = document.createElement('meta')
      robots.name = 'robots'
      document.head.appendChild(robots)
    }
    robots.content = 'noindex'
  } else if (robots) {
    robots.remove()
  }
})

export default router
