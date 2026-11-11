// Wrapper fetch duy nhất: base /api, gắn Bearer, parse lỗi {"error":{...}}, 401 → đăng xuất.
const BASE = '/api'
const TOKEN_KEY = 'dh_token'

let onUnauthorized = null
export function setUnauthorizedHandler(fn) { onUnauthorized = fn }

export function getToken() {
  try { return localStorage.getItem(TOKEN_KEY) } catch { return null }
}
export function setToken(t) {
  try { t ? localStorage.setItem(TOKEN_KEY, t) : localStorage.removeItem(TOKEN_KEY) } catch { /* bỏ qua */ }
}

export class ApiError extends Error {
  constructor(status, code, message, fields, extra) {
    super(message)
    this.status = status
    this.code = code
    this.fields = fields || {}
    this.extra = extra || {}
  }
}

function buildQuery(params) {
  if (!params) return ''
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') q.set(k, v)
  }
  const s = q.toString()
  return s ? '?' + s : ''
}

export async function request(method, path, { body, params, form, auth = true } = {}) {
  const headers = { Accept: 'application/json' }
  const token = auth ? getToken() : null
  if (token) headers.Authorization = 'Bearer ' + token
  let payload
  if (form) payload = form
  else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }
  let res
  try {
    res = await fetch(BASE + path + buildQuery(params), { method, headers, body: payload })
  } catch {
    throw new ApiError(0, 'NETWORK', 'Không kết nối được máy chủ. Kiểm tra mạng và thử lại.')
  }
  if (res.status === 204) return null
  let data = null
  const text = await res.text()
  if (text) {
    try { data = JSON.parse(text) } catch { data = null }
  }
  if (!res.ok) {
    const e = data && data.error ? data.error : {}
    let message = e.message || 'Đã có lỗi xảy ra, vui lòng thử lại.'
    if (res.status === 429) message = 'Bạn thao tác quá nhanh, vui lòng thử lại sau ít phút.'
    const err = new ApiError(res.status, e.code || 'HTTP_' + res.status, message, e.fields, {
      existingId: e.existing_id,
      retryAfter: Number(res.headers.get('Retry-After')) || 0,
    })
    if (res.status === 401 && token && onUnauthorized) onUnauthorized(err)
    throw err
  }
  return data
}

export const http = {
  get: (p, params, opt) => request('GET', p, { params, ...opt }),
  post: (p, body, opt) => request('POST', p, { body, ...opt }),
  put: (p, body) => request('PUT', p, { body }),
  patch: (p, body) => request('PATCH', p, { body }),
  del: (p) => request('DELETE', p),
  upload: (p, form) => request('POST', p, { form }),
}
