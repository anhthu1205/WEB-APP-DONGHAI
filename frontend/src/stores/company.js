import { defineStore } from 'pinia'
import { publicApi } from '../api/public'

export const DEFAULT_LOGO = '/favicon.svg'

function applyFavicon(href) {
  if (typeof document === 'undefined') return
  let link = document.querySelector('link[rel="icon"]')
  if (!link) { link = document.createElement('link'); link.rel = 'icon'; document.head.appendChild(link) }
  if (link.getAttribute('href') === href) return
  link.removeAttribute('type')
  link.setAttribute('href', href)
}

export const useCompanyStore = defineStore('company', {
  state: () => ({ data: null, loading: false, error: null, promise: null }),
  getters: {
    // Logo riêng do admin tải lên, chưa có thì dùng logo mặc định
    logoUrl: (s) => s.data?.logo_url || DEFAULT_LOGO,
  },
  actions: {
    // Gán thông tin công ty (sau khi tải hoặc admin sửa) và cập nhật favicon theo logo
    set(d) {
      this.data = d
      applyFavicon(d?.logo_url || DEFAULT_LOGO)
    },
    load(force = false) {
      if (this.data && !force) return Promise.resolve(this.data)
      if (this.promise && !force) return this.promise
      this.loading = true
      this.error = null
      this.promise = publicApi.company()
        .then((d) => { this.set(d); return d })
        .catch((e) => { this.error = e; return null })
        .finally(() => { this.loading = false; this.promise = null })
      return this.promise
    },
  },
})
