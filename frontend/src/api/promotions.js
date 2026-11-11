import { http } from './http'

export const promotionsApi = {
  get: (id) => http.get('/internal/promotions/' + id),
  list: (params) => http.get('/admin/promotions', params),
  create: (body) => http.post('/admin/promotions', body),
  update: (id, body) => http.put('/admin/promotions/' + id, body),
  remove: (id) => http.del('/admin/promotions/' + id),
}
