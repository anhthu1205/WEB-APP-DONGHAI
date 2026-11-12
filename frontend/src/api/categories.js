import { http } from './http'

export const categoriesApi = {
  list: () => http.get('/internal/categories'),
  create: (body) => http.post('/admin/categories', body),
  update: (id, body) => http.put('/admin/categories/' + id, body),
  remove: (id) => http.del('/admin/categories/' + id),
  uploadCover: (id, blob, filename) => {
    const fd = new FormData()
    fd.append('file', blob, filename)
    return http.upload(`/admin/categories/${id}/cover`, fd)
  },
  removeCover: (id) => http.del(`/admin/categories/${id}/cover`),
}
