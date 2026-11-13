// v-reveal: phần tử hiện dần (mờ → rõ, trượt lên) khi cuộn tới.
// Dùng: v-reveal hoặc v-reveal="i" (i = thứ tự để so le độ trễ, tối đa 8 bước).
// Người dùng bật "giảm chuyển động" hoặc trình duyệt không có IntersectionObserver → hiện ngay.
const reduced = () => typeof window !== 'undefined' && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

let io = null
function observer() {
  if (io) return io
  io = new IntersectionObserver((entries) => {
    for (const e of entries) {
      if (e.isIntersecting) {
        e.target.classList.add('is-in')
        io.unobserve(e.target)
      }
    }
  }, { rootMargin: '0px 0px -8% 0px', threshold: 0.08 })
  return io
}

export const reveal = {
  mounted(el, binding) {
    if (reduced() || typeof IntersectionObserver === 'undefined') return
    const i = Number(binding.value) || 0
    el.style.setProperty('--reveal-delay', `${Math.min(i % 8, 8) * 60}ms`)
    el.classList.add('reveal')
    observer().observe(el)
  },
  beforeUnmount(el) {
    io?.unobserve(el)
  },
}
