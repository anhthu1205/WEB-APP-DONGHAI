import { http } from './http'

export const publicApi = {
  company: () => http.get('/public/company', null, { auth: false }),
  products: (q) => http.get('/public/products', { q }, { auth: false }),
  product: (slug) => http.get('/public/products/' + encodeURIComponent(slug), null, { auth: false }),
  warrantyLookup: (phone) => http.post('/public/warranty/lookup', { phone }, { auth: false }),
}
