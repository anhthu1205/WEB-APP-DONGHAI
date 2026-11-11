package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"donghai/internal/domain"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		remote  string
		want    string
	}{
		{"RemoteAddr", nil, "10.0.0.1:5555", "10.0.0.1"},
		{"CF ưu tiên", map[string]string{"CF-Connecting-IP": "1.1.1.1", "X-Forwarded-For": "9.9.9.9, 2.2.2.2"}, "10.0.0.1:1", "1.1.1.1"},
		{"XFF lấy phần tử cuối", map[string]string{"X-Forwarded-For": "6.6.6.6, 3.3.3.3"}, "10.0.0.1:1", "3.3.3.3"},
		{"XFF rác → RemoteAddr", map[string]string{"X-Forwarded-For": "rác"}, "10.0.0.2:1", "10.0.0.2"},
		{"CF rác bị bỏ qua", map[string]string{"CF-Connecting-IP": "abc", "X-Forwarded-For": "4.4.4.4"}, "10.0.0.1:1", "4.4.4.4"},
	}
	for _, c := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = c.remote
		for k, v := range c.headers {
			r.Header.Set(k, v)
		}
		if got := ClientIP(r); got != c.want {
			t.Errorf("%s: ClientIP = %q; muốn %q", c.name, got, c.want)
		}
	}
}

func TestRateLimitSpoofXFF(t *testing.T) {
	l := NewLimiter(10, 5)
	h := RateLimit(l)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	var last *httptest.ResponseRecorder
	for i := 0; i < 6; i++ {
		r := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		// client giả phần tử đầu khác nhau, proxy thêm IP thật ở cuối
		r.Header.Set("X-Forwarded-For", "1.2.3."+string(rune('0'+i))+", 203.0.113.7")
		last = httptest.NewRecorder()
		h.ServeHTTP(last, r)
		if i < 5 && last.Code != 200 {
			t.Fatalf("lần %d phải qua, được %d", i+1, last.Code)
		}
	}
	if last.Code != http.StatusTooManyRequests || last.Header().Get("Retry-After") == "" {
		t.Fatalf("lần 6 phải 429 có Retry-After, được %d %v", last.Code, last.Header())
	}
	if !strings.Contains(last.Body.String(), `"RATE_LIMITED"`) {
		t.Fatalf("body sai: %s", last.Body.String())
	}
}

func TestGlobalLimit(t *testing.T) {
	l := NewLimiter(60, 60).WithGlobal(3600, 2)
	ok1, _ := l.Allow("a")
	ok2, _ := l.Allow("b")
	ok3, wait := l.Allow("c")
	if !ok1 || !ok2 || ok3 || wait < 1 {
		t.Fatalf("giới hạn toàn cục sai: %v %v %v %d", ok1, ok2, ok3, wait)
	}
}

func TestLimiterGC(t *testing.T) {
	l := NewLimiter(60, 1)
	now := time.Now()
	l.nowFunc = func() time.Time { return now }
	l.Allow("x")
	now = now.Add(20 * time.Minute)
	l.Allow("y")
	if _, ok := l.m["x"]; ok {
		t.Fatal("visitor cũ phải bị dọn")
	}
}

type fakeAuth struct{ users map[string]*domain.User }

func (f fakeAuth) Authenticate(_ context.Context, tok string) (*domain.User, error) {
	if u, ok := f.users[tok]; ok {
		return u, nil
	}
	return nil, domain.Unauthorized("x")
}

func TestRequireAuthAdmin(t *testing.T) {
	fa := fakeAuth{users: map[string]*domain.User{
		"tu": {ID: 2, Role: domain.RoleUser, IsActive: true},
		"ta": {ID: 1, Role: domain.RoleAdmin, IsActive: true},
	}}
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			t.Fatal("thiếu user trong context")
		}
		w.WriteHeader(200)
	})
	internal := RequireAuth(fa)(ok)
	admin := RequireAuth(fa)(RequireAdmin(ok))
	cases := []struct {
		h     http.Handler
		auth  string
		want  int
		label string
	}{
		{internal, "", 401, "không token"},
		{internal, "Bearer rac", 401, "token rác"},
		{internal, "Basic tu", 401, "sai scheme"},
		{internal, "Bearer tu", 200, "user vào nội bộ"},
		{admin, "Bearer tu", 403, "user vào admin"},
		{admin, "bearer ta", 200, "admin vào admin"},
	}
	for _, c := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if c.auth != "" {
			r.Header.Set("Authorization", c.auth)
		}
		w := httptest.NewRecorder()
		c.h.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Errorf("%s: được %d, muốn %d", c.label, w.Code, c.want)
		}
	}
}

func TestSecureHeadersAndCORS(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "http://evil.example")
	SecureHeaders(CORS(nil)(ok)).ServeHTTP(w, r)
	for k, v := range map[string]string{
		"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY",
		"Referrer-Policy": "strict-origin-when-cross-origin",
	} {
		if w.Header().Get(k) != v {
			t.Errorf("header %s = %q", k, w.Header().Get(k))
		}
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Error("thiếu CSP")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("CORS rỗng không được thêm header")
	}
	w = httptest.NewRecorder()
	r.Header.Set("Origin", "http://localhost:5173")
	CORS([]string{"http://localhost:5173"})(ok).ServeHTTP(w, r)
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Error("CORS phải cho origin hợp lệ")
	}
}
