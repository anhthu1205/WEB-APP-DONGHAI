import { defineStore } from 'pinia'

let seq = 0
export const useToastStore = defineStore('toast', {
  state: () => ({ items: [] }),
  actions: {
    show(message, type = 'success', timeout = 3500) {
      const id = ++seq
      this.items.push({ id, message, type })
      setTimeout(() => this.dismiss(id), timeout)
    },
    success(m) { this.show(m, 'success') },
    error(m) { this.show(m, 'error', 5000) },
    dismiss(id) { this.items = this.items.filter((t) => t.id !== id) },
  },
})
