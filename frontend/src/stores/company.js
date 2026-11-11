import { defineStore } from 'pinia'
import { publicApi } from '../api/public'

export const useCompanyStore = defineStore('company', {
  state: () => ({ data: null, loading: false, error: null, promise: null }),
  actions: {
    load(force = false) {
      if (this.data && !force) return Promise.resolve(this.data)
      if (this.promise && !force) return this.promise
      this.loading = true
      this.error = null
      this.promise = publicApi.company()
        .then((d) => { this.data = d; return d })
        .catch((e) => { this.error = e; return null })
        .finally(() => { this.loading = false; this.promise = null })
      return this.promise
    },
  },
})
