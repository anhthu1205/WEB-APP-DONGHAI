// Chuẩn hóa SĐT phía client — chỉ để hỗ trợ UX; backend mới là nơi kiểm tra thật.
export function normalizePhone(s) {
  let v = String(s || '').replace(/[\s.\-()]/g, '')
  if (!v) return null
  if (v.startsWith('+')) {
    if (!v.startsWith('+84')) return null
    v = '0' + v.slice(3)
  }
  if (!/^\d+$/.test(v)) return null
  if (v.startsWith('0084')) v = '0' + v.slice(4)
  else if (v.startsWith('84') && (v.length === 11 || v.length === 12)) v = '0' + v.slice(2)
  return /^0\d{9,10}$/.test(v) ? v : null
}

export function prettyPhone(p) {
  if (!p) return ''
  if (p.length === 10) return `${p.slice(0, 4)} ${p.slice(4, 7)} ${p.slice(7)}`
  if (p.length === 11) return `${p.slice(0, 3)} ${p.slice(3, 7)} ${p.slice(7)}`
  return p
}
