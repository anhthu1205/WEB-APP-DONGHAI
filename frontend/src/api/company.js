import { http } from './http'

export const companyApi = {
  update: (body) => http.put('/admin/company', body),
}
