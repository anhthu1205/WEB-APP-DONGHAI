import { http } from './http'

export const productsApi = {
  list: (q) => http.get('/internal/products', { q }),
  get: (id) => http.get('/internal/products/' + id),
  priceList: () => http.get('/internal/price-list'),
  create: (body) => http.post('/admin/products', body),
  update: (id, body) => http.put('/admin/products/' + id, body),
  setVisible: (id, isVisible) => http.patch(`/admin/products/${id}/visibility`, { is_visible: isVisible }),
  remove: (id) => http.del('/admin/products/' + id),
  duplicate: (id) => http.post(`/admin/products/${id}/duplicate`),
  uploadImage: (id, blob, filename) => {
    const fd = new FormData()
    fd.append('file', blob, filename)
    return http.upload(`/admin/products/${id}/images`, fd)
  },
  setMainImage: (uuid) => http.patch('/admin/images/' + uuid, { is_main: true }),
  removeImage: (uuid) => http.del('/admin/images/' + uuid),
}
