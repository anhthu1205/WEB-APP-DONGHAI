package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ClientIP xác định IP client, KHÔNG tin header đầu chuỗi do client tự gửi:
// CF-Connecting-IP (Cloudflare ghi đè) → phần tử CUỐI của X-Forwarded-For (do proxy gần nhất thêm) → RemoteAddr.
func ClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" && net.ParseIP(ip) != nil {
		return ip
	}
	if xff := r.Header.Values("X-Forwarded-For"); len(xff) > 0 {
		last := xff[len(xff)-1]
		parts := strings.Split(last, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(ip) != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type visitor struct {
	lim  *rate.Limiter
	seen time.Time
}

// Limiter là token bucket theo IP (map + dọn định kỳ), kèm giới hạn toàn cục tùy chọn.
type Limiter struct {
	mu      sync.Mutex
	perIP   rate.Limit
	burst   int
	m       map[string]*visitor
	global  *rate.Limiter
	lastGC  time.Time
	ttl     time.Duration
	nowFunc func() time.Time
}

// NewLimiter: perMinute request/phút mỗi IP, burst.
func NewLimiter(perMinute float64, burst int) *Limiter {
	return &Limiter{
		perIP: rate.Limit(perMinute / 60), burst: burst, m: map[string]*visitor{},
		ttl: 10 * time.Minute, nowFunc: time.Now,
	}
}

// WithGlobal thêm giới hạn toàn cục perHour request/giờ (chống dò phân tán).
func (l *Limiter) WithGlobal(perHour float64, burst int) *Limiter {
	l.global = rate.NewLimiter(rate.Limit(perHour/3600), burst)
	return l
}

// Allow trả (cho phép, số giây phải chờ).
func (l *Limiter) Allow(key string) (bool, int) {
	now := l.nowFunc()
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.lastGC) > time.Minute {
		for k, v := range l.m {
			if now.Sub(v.seen) > l.ttl {
				delete(l.m, k)
			}
		}
		l.lastGC = now
	}
	v, ok := l.m[key]
	if !ok {
		v = &visitor{lim: rate.NewLimiter(l.perIP, l.burst)}
		l.m[key] = v
	}
	v.seen = now
	res := v.lim.ReserveN(now, 1)
	if d := res.DelayFrom(now); d > 0 {
		res.CancelAt(now)
		return false, retrySeconds(d)
	}
	if l.global != nil {
		g := l.global.ReserveN(now, 1)
		if d := g.DelayFrom(now); d > 0 {
			g.CancelAt(now)
			res.CancelAt(now)
			return false, retrySeconds(d)
		}
	}
	return true, 0
}

func retrySeconds(d time.Duration) int {
	s := int(math.Ceil(d.Seconds()))
	if s < 1 {
		s = 1
	}
	return s
}

// RateLimit middleware theo ClientIP; vượt → 429 + Retry-After.
func RateLimit(l *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if l != nil {
				if ok, wait := l.Allow(ClientIP(r)); !ok {
					w.Header().Set("Retry-After", strconv.Itoa(wait))
					WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Bạn thao tác quá nhanh, vui lòng thử lại sau ít phút")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
