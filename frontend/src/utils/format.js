const vnd = new Intl.NumberFormat('vi-VN')

/** 5990000 → "5.990.000 ₫" */
export function formatVND(n) {
  if (n === null || n === undefined || n === '' || Number.isNaN(Number(n))) return '—'
  return vnd.format(Number(n)) + ' ₫'
}

/** "2026-11-11" hoặc ISO → "11/11/2026" */
export function formatDate(s) {
  if (!s) return '—'
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(s)
  if (m && s.length === 10) return `${m[3]}/${m[2]}/${m[1]}`
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  const p = (x) => String(x).padStart(2, '0')
  return `${p(d.getDate())}/${p(d.getMonth() + 1)}/${d.getFullYear()}`
}

export function formatDateTime(s) {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  const p = (x) => String(x).padStart(2, '0')
  return `${p(d.getHours())}:${p(d.getMinutes())} ${formatDate(s)}`
}

/** Hôm nay theo giờ Việt Nam dạng YYYY-MM-DD */
export function todayVN() {
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Ho_Chi_Minh' }).format(new Date())
}

/** Ngày + 2 năm (29/02 → 28/02), dùng hiển thị dự kiến phía client */
export function addWarrantyYears(iso, years = 2) {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso || '')
  if (!m) return ''
  let y = +m[1] + years
  let d = +m[3]
  const mo = +m[2]
  const leap = (yy) => (yy % 4 === 0 && yy % 100 !== 0) || yy % 400 === 0
  if (mo === 2 && d === 29 && !leap(y)) d = 28
  return `${y}-${m[2]}-${String(d).padStart(2, '0')}`
}

export function telHref(phone) {
  return 'tel:' + String(phone || '').replace(/[^\d+]/g, '')
}
