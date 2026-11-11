import { http } from './http'

export const customersApi = {
  list: (params) => http.get('/internal/customers', params),
  byPhone: (phone) => http.get('/internal/customers/by-phone', { phone }),
  get: (id) => http.get('/internal/customers/' + id),
  create: (body) => http.post('/internal/customers', body),
  update: (id, body) => http.put('/admin/customers/' + id, body),
  remove: (id) => http.del('/admin/customers/' + id),
}
