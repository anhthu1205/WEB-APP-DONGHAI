import { http } from './http'

export const usersApi = {
  list: () => http.get('/admin/users'),
  create: (body) => http.post('/admin/users', body),
  update: (id, body) => http.patch('/admin/users/' + id, body),
  resetPassword: (id, newPassword) => http.post(`/admin/users/${id}/reset-password`, { new_password: newPassword }),
}
