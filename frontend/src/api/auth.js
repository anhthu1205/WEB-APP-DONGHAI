import { http } from './http'

export const authApi = {
  login: (username, password) => http.post('/auth/login', { username, password }, { auth: false }),
  me: () => http.get('/auth/me'),
  changePassword: (current_password, new_password) => http.post('/auth/change-password', { current_password, new_password }),
}
