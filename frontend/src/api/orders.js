import { http } from './http'

export const ordersApi = {
  dashboard: () => http.get('/internal/dashboard'),
  list: (params) => http.get('/internal/orders', params),
  get: (id) => http.get('/internal/orders/' + id),
  create: (body) => http.post('/internal/orders', body),
  update: (id, body) => http.put('/admin/orders/' + id, body),
  remove: (id) => http.del('/admin/orders/' + id),
}
