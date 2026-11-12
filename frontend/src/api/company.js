import { http } from './http'

export const companyApi = {
  update: (body) => http.put('/admin/company', body),
  uploadLogo: (blob, filename) => {
    const fd = new FormData()
    fd.append('file', blob, filename)
    return http.upload('/admin/company/logo', fd)
  },
  deleteLogo: () => http.del('/admin/company/logo'),
}
