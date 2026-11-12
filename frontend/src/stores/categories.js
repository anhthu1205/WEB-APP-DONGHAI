import { defineStore } from 'pinia'
import { publicApi } from '../api/public'

// Danh mục công khai (cho menu đầu trang). Lỗi tải không chặn trang: menu chỉ thiếu mục danh mục.
export const useCategoriesStore = defineStore('categories', {
  state: () => ({ list: [], loaded: false, promise: null }),
  getters: {
    menu: (s) => s.list.filter((c) => c.show_in_menu),
  },
  actions: {
    load(force = false) {
      if (this.loaded && !force) return Promise.resolve(this.list)
      if (this.promise && !force) return this.promise
      this.promise = publicApi.categories()
        .then((d) => { this.list = d; this.loaded = true; return d })
        .catch(() => this.list)
        .finally(() => { this.promise = null })
      return this.promise
    },
  },
})
