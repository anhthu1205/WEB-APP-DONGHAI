import { defineStore } from 'pinia'
import { authApi } from '../api/auth'
import { getToken, setToken } from '../api/http'

export const useAuthStore = defineStore('auth', {
  state: () => ({ token: getToken(), user: null, checked: false }),
  getters: {
    isLoggedIn: (s) => !!s.token,
    isAdmin: (s) => s.user?.role === 'admin',
  },
  actions: {
    async login(username, password) {
      const res = await authApi.login(username, password)
      this.token = res.token
      this.user = res.user
      this.checked = true
      setToken(res.token)
      return res.user
    },
    /** Xác thực lại token khi app khởi động (gọi 1 lần). */
    async ensureUser() {
      if (!this.token) { this.checked = true; return null }
      if (this.user && this.checked) return this.user
      try {
        this.user = await authApi.me()
      } catch {
        this.logout()
      }
      this.checked = true
      return this.user
    },
    replaceToken(token) {
      this.token = token
      setToken(token)
    },
    logout() {
      this.token = null
      this.user = null
      setToken(null)
    },
  },
})
