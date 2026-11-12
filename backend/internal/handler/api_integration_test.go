//go:build integration

package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"donghai/internal/db"
	"donghai/internal/handler"
	"donghai/internal/service"
	tu "donghai/internal/testutil"
	"donghai/seed"
)

const adminPass = "Admin@12345"

type env struct {
	t   *testing.T
	srv *httptest.Server
	svc *service.Service
	ta  string // token admin
}

func setup(t *testing.T, limits handler.Limits) *env {
	t.Helper()
	pool := tu.MigratedPool(t)
	ctx := context.Background()
	if _, err := db.SeedDemo(ctx, pool, seed.DemoSQL); err != nil {
		t.Fatal(err)
	}
	svc := tu.NewService(pool, tu.FixedNow)
	if _, err := svc.BootstrapAdmin(ctx, "admin", adminPass); err != nil {
		t.Fatal(err)
	}
	static := t.TempDir()
	_ = os.WriteFile(filepath.Join(static, "index.html"), []byte("<!doctype html><title>SPA</title>"), 0o644)
	_ = os.MkdirAll(filepath.Join(static, "assets"), 0o755)
	_ = os.WriteFile(filepath.Join(static, "assets", "app.js"), []byte("console.log(1)"), 0o644)
	e := &env{t: t, srv: tu.NewServer(t, svc, limits, static), svc: svc}
	e.ta = e.login("admin", adminPass)
	return e
}

func (e *env) url(p string) string { return e.srv.URL + p }

func (e *env) do(method, path, token string, body any, headers ...string) *tu.Resp {
	e.t.Helper()
	return tu.Do(e.t, method, e.url(path), token, body, headers...)
}

func (e *env) login(u, p string) string {
	e.t.Helper()
	r := e.do("POST", "/api/auth/login", "", map[string]string{"username": u, "password": p})
	if r.Status != 200 {
		e.t.Fatalf("đăng nhập %s: %d %s", u, r.Status, r.Body)
	}
	return r.Map(e.t)["token"].(string)
}

func (e *env) expect(r *tu.Resp, status int, label string) {
	e.t.Helper()
	if r.Status != status {
		e.t.Fatalf("%s: status %d, muốn %d; body=%s", label, r.Status, status, r.Body)
	}
}

func errCode(t *testing.T, r *tu.Resp) string {
	t.Helper()
	m := r.Map(t)
	er, ok := m["error"].(map[string]any)
	if !ok {
		t.Fatalf("thiếu error: %s", r.Body)
	}
	return er["code"].(string)
}

func (e *env) createUser(username string) string {
	e.t.Helper()
	r := e.do("POST", "/api/admin/users", e.ta, map[string]string{"username": username, "full_name": "Nhân viên", "password": "nhanvien123"})
	e.expect(r, 201, "tạo user")
	return e.login(username, "nhanvien123")
}

func pngBytes(t *testing.T, w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		img.Set(x, x%h, color.RGBA{11, 94, 215, 255})
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func jpegBytes(t *testing.T) []byte {
	var b bytes.Buffer
	if err := jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 40, 30)), nil); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func webpBytes() []byte {
	b := []byte("RIFF\x24\x00\x00\x00WEBPVP8 ")
	return append(b, bytes.Repeat([]byte{0}, 40)...)
}

func (e *env) upload(token string, pid any, filename string, data []byte) *tu.Resp {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", filename)
	_, _ = fw.Write(data)
	_ = mw.Close()
	req, _ := http.NewRequest("POST", e.url(fmt.Sprintf("/api/admin/products/%v/images", pid)), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	return tu.DoRaw(e.t, req)
}

// uploadTo gửi multipart "file" tới path bất kỳ.
func (e *env) uploadTo(token, path, filename string, data []byte) *tu.Resp {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", filename)
	_, _ = fw.Write(data)
	_ = mw.Close()
	req, _ := http.NewRequest("POST", e.url(path), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	return tu.DoRaw(e.t, req)
}

func TestAPI(t *testing.T) {
	e := setup(t, handler.Limits{})
	tUser := e.createUser("nhanvien1")

	t.Run("health+headers", func(t *testing.T) {
		r := e.do("GET", "/api/health", "", nil)
		e.expect(r, 200, "health")
		if string(bytes.TrimSpace(r.Body)) != `{"status":"ok"}` {
			t.Fatalf("health body %s", r.Body)
		}
		for k, v := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Referrer-Policy": "strict-origin-when-cross-origin"} {
			if r.Header.Get(k) != v {
				t.Errorf("T-S1 header %s=%q", k, r.Header.Get(k))
			}
		}
		if r.Header.Get("Content-Security-Policy") == "" || r.Header.Get("Access-Control-Allow-Origin") != "" {
			t.Error("T-S1/T-S3 CSP/CORS sai")
		}
	})

	t.Run("T-PUB1..4 công khai", func(t *testing.T) {
		r := e.do("GET", "/api/public/products", "", nil)
		e.expect(r, 200, "products")
		for _, bad := range []string{"price_internal", "promo", "is_visible", "binh-nong-lanh-dh-wh30"} {
			if bytes.Contains(r.Body, []byte(bad)) {
				t.Errorf("T-PUB1 JSON công khai chứa %q", bad)
			}
		}
		var list []map[string]any
		r.JSON(t, &list)
		if len(list) != 7 {
			t.Fatalf("T-PUB1 phải 7 SP, được %d", len(list))
		}
		if q := e.do("GET", "/api/public/products?q=b%E1%BA%BFp", "", nil); !bytes.Contains(q.Body, []byte("bep-tu-doi")) {
			t.Errorf("tìm ?q= sai: %s", q.Body)
		}
		e.expect(e.do("GET", "/api/public/products/binh-nong-lanh-dh-wh30", "", nil), 404, "T-PUB2")
		d := e.do("GET", "/api/public/products/may-loc-nuoc-ro-dh-9", "", nil)
		e.expect(d, 200, "T-PUB3")
		m := d.Map(t)
		if m["price"].(float64) != 5990000 || m["description"] == "" || len(m["specs"].([]any)) == 0 || m["images"] == nil {
			t.Errorf("T-PUB3 chi tiết sai: %s", d.Body)
		}
		if bytes.Contains(d.Body, []byte("price_internal")) || bytes.Contains(d.Body, []byte("promo")) {
			t.Error("T-PUB3 lộ giá nội bộ")
		}
		c := e.do("GET", "/api/public/company", "", nil).Map(t)
		s := c["socials"].(map[string]any)
		for _, k := range []string{"facebook", "tiktok", "zalo", "shopee"} {
			if _, ok := s[k]; !ok {
				t.Errorf("T-PUB4 thiếu socials.%s", k)
			}
		}
	})

	t.Run("T-PUB6 404 API + SPA fallback", func(t *testing.T) {
		r := e.do("GET", "/api/khong-ton-tai", "", nil)
		e.expect(r, 404, "api lạ")
		if errCode(t, r) != "NOT_FOUND" {
			t.Error("phải JSON NOT_FOUND")
		}
		e.expect(e.do("GET", "/api/public/khong-co", "", nil), 404, "api lạ lồng")
		for _, p := range []string{"/", "/bao-hanh", "/noi-bo/bang-gia"} {
			r := e.do("GET", p, "", nil)
			if r.Status != 200 || !bytes.Contains(r.Body, []byte("SPA")) {
				t.Errorf("SPA fallback %s: %d %s", p, r.Status, r.Body)
			}
		}
		e.expect(e.do("GET", "/assets/app.js", "", nil), 200, "asset")
		e.expect(e.do("GET", "/assets/thieu.js", "", nil), 404, "asset thiếu")
	})

	t.Run("T-W5..W8 tra bảo hành", func(t *testing.T) {
		r := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "+84912345678"})
		e.expect(r, 200, "T-W5")
		var res struct {
			PhoneMasked string `json:"phone_masked"`
			NameMasked  string `json:"customer_name_masked"`
			Today       string `json:"today"`
			Orders      []struct {
				OrderCode string `json:"order_code"`
				Purchase  string `json:"purchase_date"`
				End       string `json:"warranty_end_date"`
				Status    string `json:"status"`
				DaysLeft  int    `json:"days_left"`
			} `json:"orders"`
		}
		r.JSON(t, &res)
		if res.PhoneMasked != "091****678" || res.NameMasked != "N*** V*** An" || len(res.Orders) != 2 || res.Today != "2026-11-11" {
			t.Fatalf("T-W5 sai: %s", r.Body)
		}
		if res.Orders[0].Purchase != "2026-11-05" || res.Orders[1].Purchase != "2025-03-10" {
			t.Errorf("T-W5 phải sắp purchase_date DESC: %s", r.Body)
		}
		if !strings.HasPrefix(res.Orders[0].OrderCode, "DH000") || len(res.Orders[0].OrderCode) != 8 {
			t.Errorf("order_code sai: %s", res.Orders[0].OrderCode)
		}
		if o := res.Orders[1]; o.End != "2027-03-10" || o.Status != "active" || o.DaysLeft != 119 {
			t.Errorf("bảo hành đơn 2025-03-10 sai: %+v", o)
		}
		for _, bad := range []string{"address", "price", "note", "created_by"} {
			if bytes.Contains(r.Body, []byte(bad)) {
				t.Errorf("T-W7 lộ %q", bad)
			}
		}
		b := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "0987 654 321"})
		if !bytes.Contains(b.Body, []byte(`"warranty_end_date":"2026-08-15","status":"expired"`)) {
			t.Errorf("T-W6 Bình: %s", b.Body)
		}
		c := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "0909111222"})
		if !bytes.Contains(c.Body, []byte(`"purchase_date":"2024-02-29","warranty_end_date":"2026-02-28","status":"expired"`)) {
			t.Errorf("T-W6 Cường 29/02: %s", c.Body)
		}
		n := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "0911000000"})
		e.expect(n, 200, "T-W8 không có khách")
		if !bytes.Contains(n.Body, []byte(`"orders":[]`)) {
			t.Errorf("T-W8 phải orders rỗng: %s", n.Body)
		}
		bad := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "abc"})
		e.expect(bad, 400, "T-W8 SĐT sai")
		if errCode(t, bad) != "VALIDATION_ERROR" || !bytes.Contains(bad.Body, []byte(`"phone"`)) {
			t.Errorf("T-W8 phải có fields.phone: %s", bad.Body)
		}
	})

	t.Run("T-A2..A5 phân quyền", func(t *testing.T) {
		r := e.do("GET", "/api/internal/price-list", "", nil)
		e.expect(r, 401, "T-A2")
		if errCode(t, r) != "UNAUTHORIZED" {
			t.Error("T-A2 code")
		}
		e.expect(e.do("GET", "/api/internal/price-list", "rác.rác.rác", nil), 401, "T-A3")
		for _, c := range []struct{ m, p string }{
			{"POST", "/api/admin/products"}, {"PUT", "/api/admin/company"}, {"POST", "/api/admin/users"},
			{"PUT", "/api/admin/customers/1"}, {"DELETE", "/api/admin/orders/1"}, {"GET", "/api/admin/users"},
		} {
			r := e.do(c.m, c.p, tUser, map[string]any{})
			e.expect(r, 403, "T-A4 "+c.m+" "+c.p)
			if errCode(t, r) != "FORBIDDEN" {
				t.Error("T-A4 code")
			}
		}
		e.expect(e.upload(tUser, 1, "a.png", pngBytes(t, 10, 10)), 403, "T-IMG6")
		me := e.do("GET", "/api/auth/me", tUser, nil)
		e.expect(me, 200, "me")
		if bytes.Contains(me.Body, []byte("password_hash")) || bytes.Contains(me.Body, []byte("token_version")) {
			t.Error("me lộ hash")
		}
	})

	t.Run("T-I1..I2 sản phẩm & bảng giá nội bộ", func(t *testing.T) {
		r := e.do("GET", "/api/internal/products", tUser, nil)
		e.expect(r, 200, "T-I1")
		var list []map[string]any
		r.JSON(t, &list)
		if len(list) != 8 || !bytes.Contains(r.Body, []byte("price_internal")) {
			t.Fatalf("T-I1: %d SP", len(list))
		}
		pl := e.do("GET", "/api/internal/price-list", tUser, nil)
		e.expect(pl, 200, "T-I2")
		var res struct {
			Today      string `json:"today"`
			Promotions []struct {
				Title  string `json:"title"`
				Status string `json:"status"`
				Items  []any  `json:"items"`
			} `json:"promotions"`
			Products []struct {
				ID    int64  `json:"id"`
				Promo *int64 `json:"current_promo_price"`
			} `json:"products"`
		}
		pl.JSON(t, &res)
		titles := []string{}
		for _, p := range res.Promotions {
			titles = append(titles, p.Title+"/"+p.Status)
		}
		got := strings.Join(titles, ",")
		if got != "Khuyến mãi cuối năm 2026/active,Tết Đinh Mùi 2027/upcoming" {
			t.Errorf("T-I2 KM: %s", got)
		}
		if res.Products[0].Promo == nil || *res.Products[0].Promo != 5490000 || res.Products[1].Promo != nil {
			t.Errorf("T-I2 giá KM: %s", pl.Body)
		}
		e.expect(e.do("GET", "/api/internal/promotions/1", tUser, nil), 200, "promo detail")
		e.expect(e.do("GET", "/api/internal/dashboard", tUser, nil), 200, "dashboard")
	})

	t.Run("T-I3..I8 khách & đơn", func(t *testing.T) {
		// Khách 1 do seed tạo (ngoài phạm vi nhân viên): 409 nhưng không lộ tên/id.
		r := e.do("POST", "/api/internal/customers", tUser, map[string]string{"full_name": "Ai Đó", "phone": "0912 345 678"})
		e.expect(r, 409, "T-I3")
		if bytes.Contains(r.Body, []byte(`existing_id`)) || bytes.Contains(r.Body, []byte("Nguyễn Văn An")) {
			t.Errorf("T-I3 lộ khách ngoài phạm vi: %s", r.Body)
		}
		ra := e.do("POST", "/api/internal/customers", e.ta, map[string]string{"full_name": "Ai Đó", "phone": "0912 345 678"})
		e.expect(ra, 409, "T-I3 admin")
		if !bytes.Contains(ra.Body, []byte(`"existing_id":1`)) {
			t.Errorf("T-I3 admin thiếu existing_id: %s", ra.Body)
		}
		nc := e.do("POST", "/api/internal/customers", tUser, map[string]string{"full_name": "Khách Mới", "phone": "0901000111", "address": "HP"})
		e.expect(nc, 201, "T-A5 tạo khách")
		e.expect(e.do("GET", "/api/internal/customers/by-phone?phone=%2B84987654321", tUser, nil), 404, "T-I4 ngoài phạm vi")
		bp := e.do("GET", "/api/internal/customers/by-phone?phone=%2B84987654321", e.ta, nil)
		e.expect(bp, 200, "T-I4")
		if !bytes.Contains(bp.Body, []byte("Trần Thị Bình")) {
			t.Error("T-I4 sai khách")
		}
		e.expect(e.do("GET", "/api/internal/customers/by-phone?phone=0911999888", tUser, nil), 404, "T-I4 404")
		e.expect(e.do("GET", "/api/internal/customers?q=0912", tUser, nil), 200, "list customers")

		o := e.do("POST", "/api/internal/orders", tUser, map[string]any{
			"customer":      map[string]string{"full_name": "Đỗ Văn Em", "phone": "0933 444 555", "address": "Hải Phòng"},
			"purchase_date": "2026-11-10",
			"items":         []map[string]any{{"product_id": 1, "quantity": 1, "unit_price": 5490000, "serial_number": "DH9-2611-0001"}, {"product_id": 3, "quantity": 2}},
		})
		e.expect(o, 201, "T-I5 tạo đơn khách mới")
		om := o.Map(t)
		if om["customer_reused"] != false || om["created_by_username"] != "nhanvien1" {
			t.Errorf("T-I5: %s", o.Body)
		}
		items := om["items"].([]any)
		if items[1].(map[string]any)["unit_price"].(float64) != 7690000 {
			t.Errorf("T-I6 unit_price mặc định: %v", items[1])
		}
		if om["total_amount"].(float64) != 5490000+2*7690000 {
			t.Errorf("tổng tiền sai: %v", om["total_amount"])
		}
		reuse := e.do("POST", "/api/internal/orders", tUser, map[string]any{
			"customer":      map[string]string{"full_name": "Tên Khác", "phone": "+84912345678", "address": "Khác"},
			"purchase_date": "2026-11-11", "items": []map[string]any{{"product_id": 2, "quantity": 1}},
		})
		e.expect(reuse, 201, "T-I5 dùng lại khách")
		rm := reuse.Map(t)
		cust := rm["customer"].(map[string]any)
		if rm["customer_reused"] != true || cust["full_name"] != "Nguyễn Văn An" || cust["address"] != "Số 12 Lê Lợi, Ngô Quyền, Hải Phòng" {
			t.Errorf("T-I5 không được ghi đè khách cũ: %s", reuse.Body)
		}
		base := func(items any, date string) map[string]any {
			return map[string]any{"customer_id": 1, "purchase_date": date, "items": items}
		}
		one := []map[string]any{{"product_id": 1, "quantity": 1}}
		many := make([]map[string]any, 51)
		for i := range many {
			many[i] = map[string]any{"product_id": 1, "quantity": 1}
		}
		e.expect(e.do("POST", "/api/internal/orders", tUser, base([]any{}, "2026-11-01")), 400, "T-I6 0 dòng")
		e.expect(e.do("POST", "/api/internal/orders", tUser, base(many, "2026-11-01")), 400, "T-I6 51 dòng")
		e.expect(e.do("POST", "/api/internal/orders", tUser, base(one, "2026-11-12")), 400, "T-I6 ngày mai")
		e.expect(e.do("POST", "/api/internal/orders", tUser, base(one, "1999-12-31")), 400, "T-I6 trước 2000")
		e.expect(e.do("POST", "/api/internal/orders", tUser, base(one, "2026-11-11")), 201, "customer_id")

		before := e.do("GET", "/api/internal/customers?page_size=100", tUser, nil).Map(t)["total"].(float64)
		bad := e.do("POST", "/api/internal/orders", tUser, map[string]any{
			"customer":      map[string]string{"full_name": "Mồ Côi", "phone": "0944555666"},
			"purchase_date": "2026-11-01", "items": []map[string]any{{"product_id": 99999, "quantity": 1}},
		})
		e.expect(bad, 400, "T-I7 SP không tồn tại")
		after := e.do("GET", "/api/internal/customers?page_size=100", tUser, nil).Map(t)["total"].(float64)
		if before != after {
			t.Errorf("T-I7 để lại khách mồ côi: %v → %v", before, after)
		}
		lk := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "0933444555"})
		if !bytes.Contains(lk.Body, []byte(`"purchase_date":"2026-11-10","warranty_end_date":"2028-11-10","status":"active"`)) {
			t.Errorf("T-I8 tra ngay đơn mới: %s", lk.Body)
		}
		e.expect(e.do("GET", "/api/internal/orders?q=Em&from=2026-11-01&to=2026-11-30", tUser, nil), 200, "list orders")
		e.expect(e.do("GET", "/api/internal/orders?from=abc", tUser, nil), 400, "from sai")
		id := int(om["id"].(float64))
		e.expect(e.do("GET", fmt.Sprintf("/api/internal/orders/%d", id), tUser, nil), 200, "order detail")
		e.expect(e.do("GET", "/api/internal/customers/1", tUser, nil), 200, "customer detail")
	})

	t.Run("T-SC phạm vi dữ liệu theo người nhập", func(t *testing.T) {
		u3 := e.createUser("nhanvien3")
		c3 := e.do("POST", "/api/internal/customers", u3, map[string]string{"full_name": "Khách Của NV3", "phone": "0977000111"})
		e.expect(c3, 201, "NV3 tạo khách")
		cid := int(c3.Map(t)["id"].(float64))
		o3 := e.do("POST", "/api/internal/orders", u3, map[string]any{"customer_id": cid, "purchase_date": "2026-11-05",
			"items": []map[string]any{{"product_id": 1, "quantity": 2, "unit_price": 1000000}}})
		e.expect(o3, 201, "NV3 tạo đơn")
		oid := int(o3.Map(t)["id"].(float64))

		// nhanvien1 không thấy gì của nhanvien3
		if n := e.do("GET", "/api/internal/customers?q=0977000111", tUser, nil).Map(t)["total"].(float64); n != 0 {
			t.Errorf("T-SC1 NV1 thấy khách của NV3 trong danh sách")
		}
		e.expect(e.do("GET", "/api/internal/customers/by-phone?phone=0977000111", tUser, nil), 404, "T-SC2 by-phone")
		e.expect(e.do("GET", fmt.Sprintf("/api/internal/customers/%d", cid), tUser, nil), 404, "T-SC3 chi tiết khách")
		e.expect(e.do("GET", fmt.Sprintf("/api/internal/orders/%d", oid), tUser, nil), 404, "T-SC4 chi tiết đơn")
		if n := e.do("GET", "/api/internal/orders?q=0977000111", tUser, nil).Map(t)["total"].(float64); n != 0 {
			t.Errorf("T-SC5 NV1 thấy đơn của NV3 trong danh sách")
		}
		e.expect(e.do("POST", "/api/internal/orders", tUser, map[string]any{"customer_id": cid, "purchase_date": "2026-11-06",
			"items": []map[string]any{{"product_id": 1, "quantity": 1}}}), 400, "T-SC6 đặt đơn bằng customer_id ngoài phạm vi")
		dup := e.do("POST", "/api/internal/customers", tUser, map[string]string{"full_name": "X", "phone": "0977000111"})
		e.expect(dup, 409, "T-SC7 trùng SĐT")
		if bytes.Contains(dup.Body, []byte("Khách Của NV3")) || bytes.Contains(dup.Body, []byte("existing_id")) {
			t.Errorf("T-SC7 lộ khách của NV3: %s", dup.Body)
		}

		// Danh sách đơn của nhanvien1 chỉ gồm đơn do nhanvien1 nhập
		var lst struct {
			Total int `json:"total"`
			Items []struct {
				By string `json:"created_by_username"`
			} `json:"items"`
		}
		e.do("GET", "/api/internal/orders?page_size=100", tUser, nil).JSON(t, &lst)
		for _, it := range lst.Items {
			if it.By != "nhanvien1" {
				t.Errorf("T-SC8 NV1 thấy đơn của %q", it.By)
			}
		}
		var d1 struct {
			Scope       string `json:"scope"`
			OrdersTotal int    `json:"orders_total"`
			Staff       []any  `json:"staff_stats"`
		}
		e.do("GET", "/api/internal/dashboard", tUser, nil).JSON(t, &d1)
		if d1.Scope != "own" || d1.OrdersTotal != lst.Total || d1.Staff != nil {
			t.Errorf("T-SC9 dashboard NV1: %+v (đơn của NV1: %d)", d1, lst.Total)
		}

		// Admin thấy tất cả + bảng theo nhân viên
		e.expect(e.do("GET", fmt.Sprintf("/api/internal/orders/%d", oid), e.ta, nil), 200, "T-SC10 admin xem đơn NV3")
		var da struct {
			Scope string `json:"scope"`
			Staff []struct {
				Username  string `json:"username"`
				Customers int    `json:"customers_total"`
				Orders    int    `json:"orders_total"`
				Month     int    `json:"orders_this_month"`
				Revenue   int64  `json:"revenue_this_month"`
			} `json:"staff_stats"`
		}
		e.do("GET", "/api/internal/dashboard", e.ta, nil).JSON(t, &da)
		found := false
		for _, x := range da.Staff {
			if x.Username == "nhanvien3" {
				found = true
				if x.Customers != 1 || x.Orders != 1 || x.Month != 1 || x.Revenue != 2000000 {
					t.Errorf("T-SC11 số liệu NV3: %+v", x)
				}
			}
		}
		if da.Scope != "all" || !found {
			t.Errorf("T-SC11 dashboard admin: %+v", da)
		}

		// NV1 bán cho khách đã có (của NV3) qua SĐT → dùng lại khách, NV1 chỉ thấy đơn của mình
		re := e.do("POST", "/api/internal/orders", tUser, map[string]any{
			"customer":      map[string]string{"full_name": "", "phone": "0977 000 111"},
			"purchase_date": "2026-11-07", "items": []map[string]any{{"product_id": 2, "quantity": 1}},
		})
		e.expect(re, 201, "T-SC12 dùng lại khách qua SĐT")
		var cd struct {
			OrdersCount int `json:"orders_count"`
			Orders      []struct {
				By string `json:"created_by_username"`
			} `json:"orders"`
		}
		e.do("GET", fmt.Sprintf("/api/internal/customers/%d", cid), tUser, nil).JSON(t, &cd)
		if len(cd.Orders) != 1 || cd.Orders[0].By != "nhanvien1" || cd.OrdersCount != 1 {
			t.Errorf("T-SC12 NV1 thấy đơn không phải của mình: %+v", cd)
		}
		e.do("GET", fmt.Sprintf("/api/internal/customers/%d", cid), e.ta, nil).JSON(t, &cd)
		if len(cd.Orders) != 2 || cd.OrdersCount != 2 {
			t.Errorf("T-SC13 admin phải thấy cả 2 đơn: %+v", cd)
		}
	})

	t.Run("T-LOGO logo công ty", func(t *testing.T) {
		if v := e.do("GET", "/api/public/company", "", nil).Map(t)["logo_url"]; v != nil {
			t.Fatalf("T-LOGO1 chưa có logo phải là null: %v", v)
		}
		e.expect(e.do("GET", "/api/public/logo", "", nil), 404, "T-LOGO1 chưa có logo")
		e.expect(e.uploadTo(tUser, "/api/admin/company/logo", "l.png", pngBytes(t, 64, 64)), 403, "T-LOGO2 nhân viên")
		e.expect(e.uploadTo("", "/api/admin/company/logo", "l.png", pngBytes(t, 64, 64)), 401, "T-LOGO2 không token")
		e.expect(e.uploadTo(e.ta, "/api/admin/company/logo", "l.png", []byte("hello, not an image at all")), 415, "T-LOGO3 không phải ảnh")
		big := append(pngBytes(t, 10, 10), bytes.Repeat([]byte{0}, 3*1024*1024)...)
		e.expect(e.uploadTo(e.ta, "/api/admin/company/logo", "big.png", big), 413, "T-LOGO3 quá 2MB")

		up := e.uploadTo(e.ta, "/api/admin/company/logo", "l.png", pngBytes(t, 64, 64))
		e.expect(up, 200, "T-LOGO4 tải logo")
		u, _ := up.Map(t)["logo_url"].(string)
		if !strings.HasPrefix(u, "/api/public/logo?v=") {
			t.Fatalf("T-LOGO4 logo_url: %s", up.Body)
		}
		if pu := e.do("GET", "/api/public/company", "", nil).Map(t)["logo_url"]; pu != u {
			t.Errorf("T-LOGO4 public company logo_url = %v, muốn %s", pu, u)
		}
		img := e.do("GET", u, "", nil)
		e.expect(img, 200, "T-LOGO5 lấy logo")
		if img.Header.Get("Content-Type") != "image/png" || !strings.Contains(img.Header.Get("Cache-Control"), "immutable") {
			t.Errorf("T-LOGO5 header: %v", img.Header)
		}
		e.expect(e.do("GET", u, "", nil, "If-None-Match", img.Header.Get("ETag")), 304, "T-LOGO5 304")
		if c := e.do("GET", "/api/public/logo", "", nil).Header.Get("Cache-Control"); strings.Contains(c, "immutable") {
			t.Errorf("T-LOGO5 không có v thì không được cache lâu: %s", c)
		}

		up2 := e.uploadTo(e.ta, "/api/admin/company/logo", "l2.png", pngBytes(t, 32, 32))
		e.expect(up2, 200, "T-LOGO6 thay logo")
		if u2, _ := up2.Map(t)["logo_url"].(string); u2 == u {
			t.Errorf("T-LOGO6 đổi logo phải đổi URL: %s", u2)
		}
		e.expect(e.do("DELETE", "/api/admin/company/logo", tUser, nil), 403, "T-LOGO7 nhân viên xóa")
		del := e.do("DELETE", "/api/admin/company/logo", e.ta, nil)
		e.expect(del, 200, "T-LOGO7 xóa logo")
		if v := del.Map(t)["logo_url"]; v != nil {
			t.Errorf("T-LOGO7 sau xóa logo_url phải null: %v", v)
		}
		e.expect(e.do("GET", "/api/public/logo", "", nil), 404, "T-LOGO7 sau xóa")
	})

	t.Run("T-AD1..AD3 sản phẩm admin", func(t *testing.T) {
		body := map[string]any{"name": "Máy lọc nước RO Đông Hải DH-9", "short_description": "x", "description": "y",
			"specs": []map[string]string{{"label": "A", "value": "B"}}, "unit": "cái", "price_public": 100, "price_internal": 90, "is_visible": true, "sort_order": 9}
		a := e.do("POST", "/api/admin/products", e.ta, body)
		e.expect(a, 201, "T-AD1 tạo")
		if a.Map(t)["slug"] != "may-loc-nuoc-ro-dong-hai-dh-9" {
			t.Errorf("T-AD1 slug: %s", a.Body)
		}
		b := e.do("POST", "/api/admin/products", e.ta, body)
		e.expect(b, 201, "T-AD1 trùng")
		if b.Map(t)["slug"] != "may-loc-nuoc-ro-dong-hai-dh-9-2" {
			t.Errorf("T-AD1 slug -2: %s", b.Body)
		}
		body["slug"] = "Sai Slug!"
		e.expect(e.do("POST", "/api/admin/products", e.ta, body), 400, "T-AD1 slug sai")
		body["slug"] = "may-loc-nuoc-ro-dh-9"
		e.expect(e.do("POST", "/api/admin/products", e.ta, body), 409, "slug trùng")
		delete(body, "slug")
		body["id"] = 5
		e.expect(e.do("POST", "/api/admin/products", e.ta, body), 400, "T-S2 trường lạ")
		delete(body, "id")

		id := int(b.Map(t)["id"].(float64))
		body["name"] = "Đã sửa"
		u := e.do("PUT", fmt.Sprintf("/api/admin/products/%d", id), e.ta, body)
		e.expect(u, 200, "PUT SP")
		if u.Map(t)["slug"] != "may-loc-nuoc-ro-dong-hai-dh-9-2" {
			t.Error("PUT không gửi slug phải giữ slug cũ")
		}
		e.expect(e.do("PATCH", "/api/admin/products/2/visibility", e.ta, map[string]bool{"is_visible": false}), 200, "T-AD2")
		if bytes.Contains(e.do("GET", "/api/public/products", "", nil).Body, []byte("noi-com-dien-cao-tan")) {
			t.Error("T-AD2 SP ẩn vẫn hiện công khai")
		}
		e.expect(e.do("GET", "/api/public/products/noi-com-dien-cao-tan-dh-rc18", "", nil), 404, "T-AD2 chi tiết")
		e.expect(e.do("PATCH", "/api/admin/products/2/visibility", e.ta, map[string]bool{"is_visible": true}), 200, "hiện lại")

		e.expect(e.do("DELETE", "/api/admin/products/1", e.ta, nil), 409, "T-AD3 SP có đơn")
		e.expect(e.upload(e.ta, id, "a.png", pngBytes(t, 20, 20)), 201, "upload trước khi xóa")
		e.expect(e.do("DELETE", fmt.Sprintf("/api/admin/products/%d", id), e.ta, nil), 204, "T-AD3 xóa")
		e.expect(e.do("GET", fmt.Sprintf("/api/internal/products/%d", id), e.ta, nil), 404, "đã xóa")
	})

	t.Run("T-IMG1..5 ảnh", func(t *testing.T) {
		p := e.upload(e.ta, 5, "a.png", pngBytes(t, 120, 80))
		e.expect(p, 201, "T-IMG1 png")
		pm := p.Map(t)
		if pm["is_main"] != true || pm["width"].(float64) != 120 {
			t.Errorf("T-IMG1 ảnh đầu phải là chính: %s", p.Body)
		}
		j := e.upload(e.ta, 5, "b.jpg", jpegBytes(t))
		e.expect(j, 201, "T-IMG1 jpeg")
		if j.Map(t)["is_main"] != false {
			t.Error("ảnh 2 không được là chính")
		}
		e.expect(e.upload(e.ta, 5, "c.webp", webpBytes()), 201, "T-IMG1 webp")

		big := append(pngBytes(t, 10, 10), bytes.Repeat([]byte{0}, 3*1024*1024)...)
		r := e.upload(e.ta, 5, "big.png", big)
		e.expect(r, 413, "T-IMG2 3MB")
		if errCode(t, r) != "PAYLOAD_TOO_LARGE" {
			t.Error("T-IMG2 code")
		}
		for name, data := range map[string][]byte{
			"x.jpg":   []byte("chỉ là văn bản thuần, không phải ảnh"),
			"pdf.jpg": []byte("%PDF-1.4\n%âãÏÓ\n1 0 obj\n<<>>\nendobj\n"),
			"svg.jpg": []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
			"bad.png": append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{1}, 50)...),
		} {
			r := e.upload(e.ta, 5, name, data)
			e.expect(r, 415, "T-IMG3 "+name)
		}
		e.expect(e.upload(e.ta, 99999, "a.png", pngBytes(t, 5, 5)), 404, "SP không tồn tại")

		var last *tu.Resp
		for i := 0; i < 5; i++ {
			last = e.upload(e.ta, 5, "n.png", pngBytes(t, 8, 8))
			e.expect(last, 201, "upload tới 8")
		}
		e.expect(e.upload(e.ta, 5, "9.png", pngBytes(t, 8, 8)), 400, "T-IMG4 ảnh thứ 9")

		img := e.do("GET", pm["url"].(string), "", nil)
		e.expect(img, 200, "T-PUB5 lấy ảnh")
		if img.Header.Get("Content-Type") != "image/png" || !strings.Contains(img.Header.Get("Cache-Control"), "immutable") || img.Header.Get("ETag") == "" {
			t.Errorf("T-PUB5 header: %v", img.Header)
		}
		e.expect(e.do("GET", pm["url"].(string), "", nil, "If-None-Match", img.Header.Get("ETag")), 304, "T-PUB5 304")
		e.expect(e.do("GET", "/api/images/abc", "", nil), 404, "T-PUB5 abc")
		e.expect(e.do("GET", "/api/images/00000000-0000-0000-0000-000000000000", "", nil), 404, "T-PUB5 uuid không có")

		lastID := last.Map(t)["id"].(string)
		e.expect(e.do("PATCH", "/api/admin/images/"+lastID, e.ta, map[string]bool{"is_main": true}), 200, "T-IMG5 đặt chính")
		d := e.do("GET", "/api/internal/products/5", e.ta, nil).Map(t)
		mains := 0
		for _, im := range d["images"].([]any) {
			if im.(map[string]any)["is_main"] == true {
				mains++
			}
		}
		if mains != 1 || d["images"].([]any)[0].(map[string]any)["id"] != lastID {
			t.Errorf("T-IMG5 phải đúng 1 ảnh chính = ảnh vừa đặt (%d)", mains)
		}
		e.expect(e.do("DELETE", "/api/admin/images/"+lastID, e.ta, nil), 204, "T-IMG5 xóa ảnh chính")
		d = e.do("GET", "/api/internal/products/5", e.ta, nil).Map(t)
		first := d["images"].([]any)[0].(map[string]any)
		if first["is_main"] != true || first["id"] != pm["id"] {
			t.Errorf("T-IMG5 ảnh sort nhỏ nhất phải thành chính: %v", first)
		}
		pub := e.do("GET", "/api/public/products/may-hut-bui-cam-tay-dh-v12", "", nil)
		if !bytes.Contains(pub.Body, []byte(pm["id"].(string))) {
			t.Error("trang công khai phải có ảnh")
		}
	})

	t.Run("T-AD4 công ty", func(t *testing.T) {
		body := map[string]string{"name": "Công ty Đông Hải", "slogan": "", "about": "", "address": "HP", "phone": "0225",
			"email": "a@b.vn", "working_hours": "", "facebook_url": "http://facebook.com/x", "tiktok_url": "", "zalo_url": "", "shopee_url": ""}
		e.expect(e.do("PUT", "/api/admin/company", e.ta, body), 400, "T-AD4 http")
		body["facebook_url"] = "javascript:alert(1)"
		e.expect(e.do("PUT", "/api/admin/company", e.ta, body), 400, "T-AD4 javascript")
		body["facebook_url"] = "https://facebook.com/donghai"
		e.expect(e.do("PUT", "/api/admin/company", e.ta, body), 200, "T-AD4 https")
		c := e.do("GET", "/api/public/company", "", nil)
		if !bytes.Contains(c.Body, []byte(`"facebook":"https://facebook.com/donghai"`)) || !bytes.Contains(c.Body, []byte(`"tiktok":""`)) {
			t.Errorf("T-AD4 footer: %s", c.Body)
		}
	})

	t.Run("T-AD5 khuyến mãi", func(t *testing.T) {
		body := map[string]any{"title": "KM test", "content": "", "start_date": "2026-12-10", "end_date": "2026-12-01", "is_active": true, "items": []any{}}
		e.expect(e.do("POST", "/api/admin/promotions", e.ta, body), 400, "end < start")
		body["end_date"] = "2026-12-20"
		body["items"] = []map[string]any{{"product_id": 1, "promo_price": 100, "note": ""}, {"product_id": 2, "promo_price": 200, "note": ""}}
		r := e.do("POST", "/api/admin/promotions", e.ta, body)
		e.expect(r, 201, "tạo KM")
		id := int(r.Map(t)["id"].(float64))
		body["items"] = []map[string]any{{"product_id": 3, "promo_price": 300, "note": "x"}}
		u := e.do("PUT", fmt.Sprintf("/api/admin/promotions/%d", id), e.ta, body)
		e.expect(u, 200, "PUT KM")
		if n := len(u.Map(t)["items"].([]any)); n != 1 {
			t.Errorf("PUT phải thay toàn bộ items, còn %d", n)
		}
		body["items"] = []map[string]any{{"product_id": 3, "promo_price": 1, "note": ""}, {"product_id": 3, "promo_price": 2, "note": ""}}
		e.expect(e.do("PUT", fmt.Sprintf("/api/admin/promotions/%d", id), e.ta, body), 400, "SP lặp")
		e.expect(e.do("GET", "/api/admin/promotions?status=upcoming", e.ta, nil), 200, "list KM")
		e.expect(e.do("GET", "/api/admin/promotions?status=xyz", e.ta, nil), 400, "status sai")
		e.expect(e.do("DELETE", fmt.Sprintf("/api/admin/promotions/%d", id), e.ta, nil), 204, "xóa KM")
	})

	t.Run("T-AD6..AD7 sửa/xóa khách & đơn", func(t *testing.T) {
		e.expect(e.do("PUT", "/api/admin/customers/2", e.ta, map[string]string{"full_name": "Trần Thị Bình", "phone": "0912345678", "address": "", "note": ""}), 409, "T-AD6 SĐT trùng")
		e.expect(e.do("PUT", "/api/admin/customers/2", e.ta, map[string]string{"full_name": "Trần Thị Bình", "phone": "+84987654321", "address": "HD", "note": "sửa"}), 200, "sửa khách")
		e.expect(e.do("DELETE", "/api/admin/customers/1", e.ta, nil), 409, "T-AD6 xóa khách có đơn")
		nc := e.do("POST", "/api/internal/customers", e.ta, map[string]string{"full_name": "Xóa Được", "phone": "0966000111"})
		e.expect(e.do("DELETE", fmt.Sprintf("/api/admin/customers/%d", int(nc.Map(t)["id"].(float64))), e.ta, nil), 204, "xóa khách trống")

		upd := map[string]any{"customer_id": 4, "purchase_date": "2026-10-21", "note": "sửa", "items": []map[string]any{{"product_id": 6, "quantity": 3}}}
		r := e.do("PUT", "/api/admin/orders/7", e.ta, upd)
		e.expect(r, 200, "T-AD7 sửa đơn")
		if r.Map(t)["items_count"].(float64) != 3 {
			t.Errorf("T-AD7 items: %s", r.Body)
		}
		upd["customer"] = map[string]string{"phone": "0912345678"}
		e.expect(e.do("PUT", "/api/admin/orders/7", e.ta, upd), 400, "PUT đơn không nhận customer")
		e.expect(e.do("DELETE", "/api/admin/orders/7", e.ta, nil), 204, "T-AD7 xóa đơn")
		e.expect(e.do("GET", "/api/internal/orders/7", e.ta, nil), 404, "đơn đã xóa")
	})

	t.Run("T-AD8..AD9, T-A6..A7, T-A9 tài khoản", func(t *testing.T) {
		e.expect(e.do("POST", "/api/admin/users", e.ta, map[string]string{"username": "nhanvien1", "full_name": "", "password": "12345678"}), 409, "T-AD8 trùng")
		e.expect(e.do("POST", "/api/admin/users", e.ta, map[string]string{"username": "Ab", "full_name": "", "password": "12345678"}), 400, "T-AD8 username ngắn")
		e.expect(e.do("POST", "/api/admin/users", e.ta, map[string]string{"username": "abc", "full_name": "", "password": "1234567"}), 400, "T-AD8 MK ngắn")
		e.expect(e.do("POST", "/api/admin/users", e.ta, map[string]string{"username": "abc", "full_name": "", "password": "12345678", "role": "admin"}), 400, "không nhận role")
		l := e.do("GET", "/api/admin/users", e.ta, nil)
		e.expect(l, 200, "list users")
		if bytes.Contains(l.Body, []byte("password_hash")) || bytes.Contains(l.Body, []byte("token_version")) {
			t.Error("T-AD9 lộ hash")
		}
		var users []map[string]any
		l.JSON(t, &users)
		var adminID, userID int
		for _, u := range users {
			if u["username"] == "admin" {
				adminID = int(u["id"].(float64))
			}
			if u["username"] == "nhanvien1" {
				userID = int(u["id"].(float64))
				if u["role"] != "user" {
					t.Error("T-AD8 role phải user")
				}
			}
		}
		e.expect(e.do("PATCH", fmt.Sprintf("/api/admin/users/%d", adminID), e.ta, map[string]bool{"is_active": false}), 400, "T-A9 tự khóa")
		e.expect(e.do("PATCH", fmt.Sprintf("/api/admin/users/%d", userID), e.ta, map[string]string{"role": "admin"}), 400, "T-A9 role")

		// T-A7 đổi MK
		u2 := e.createUser("nhanvien2")
		cp := e.do("POST", "/api/auth/change-password", u2, map[string]string{"current_password": "sai", "new_password": "moi123456"})
		e.expect(cp, 400, "MK hiện tại sai")
		cp = e.do("POST", "/api/auth/change-password", u2, map[string]string{"current_password": "nhanvien123", "new_password": "moi123456"})
		e.expect(cp, 200, "đổi MK")
		newTok := cp.Map(t)["token"].(string)
		e.expect(e.do("GET", "/api/auth/me", u2, nil), 401, "T-A7 token cũ")
		e.expect(e.do("GET", "/api/auth/me", newTok, nil), 200, "T-A7 token mới")

		// T-A7 admin reset MK → token cũ của user 401
		tok1 := e.login("nhanvien1", "nhanvien123")
		e.expect(e.do("POST", fmt.Sprintf("/api/admin/users/%d/reset-password", userID), e.ta, map[string]string{"new_password": "reset12345"}), 204, "reset")
		e.expect(e.do("GET", "/api/auth/me", tok1, nil), 401, "T-A7 sau reset")

		// T-A6 khóa user
		tok1 = e.login("nhanvien1", "reset12345")
		e.expect(e.do("PATCH", fmt.Sprintf("/api/admin/users/%d", userID), e.ta, map[string]bool{"is_active": false}), 200, "khóa")
		e.expect(e.do("GET", "/api/internal/dashboard", tok1, nil), 401, "T-A6 token cũ")
		lr := e.do("POST", "/api/auth/login", "", map[string]string{"username": "nhanvien1", "password": "reset12345"})
		e.expect(lr, 401, "T-A6 đăng nhập bị khóa")
		wrong := e.do("POST", "/api/auth/login", "", map[string]string{"username": "nhanvien1", "password": "sai-mat-khau"})
		if string(lr.Body) != string(wrong.Body) || !bytes.Contains(lr.Body, []byte("Sai tên đăng nhập hoặc mật khẩu")) {
			t.Errorf("T-A6 message phải giống nhau: %s vs %s", lr.Body, wrong.Body)
		}
		e.expect(e.do("PATCH", fmt.Sprintf("/api/admin/users/%d", userID), e.ta, map[string]bool{"is_active": true}), 200, "mở khóa")
		e.login("nhanvien1", "reset12345")
	})

	t.Run("T-S2 body", func(t *testing.T) {
		big := append([]byte(`{"phone":"`), bytes.Repeat([]byte("1"), 1100*1024)...)
		big = append(big, []byte(`"}`)...)
		r := e.do("POST", "/api/public/warranty/lookup", "", big)
		e.expect(r, 413, "T-S2 > 1MB")
		e.expect(e.do("POST", "/api/public/warranty/lookup", "", []byte(`{"phone":`)), 400, "T-S2 JSON hỏng")
		e.expect(e.do("POST", "/api/public/warranty/lookup", "", []byte(`{"phone":"0912345678","x":1}`)), 400, "T-S2 trường lạ")
	})
}

func TestRateLimits(t *testing.T) {
	e := setup(t, handler.DefaultLimits())
	// T-W9: lookup lần 6 → 429
	for i := 1; i <= 6; i++ {
		r := e.do("POST", "/api/public/warranty/lookup", "", map[string]string{"phone": "0912345678"},
			"X-Forwarded-For", fmt.Sprintf("1.2.3.%d, 198.51.100.9", i))
		if i <= 5 && r.Status != 200 {
			t.Fatalf("lookup lần %d: %d", i, r.Status)
		}
		if i == 6 && (r.Status != 429 || r.Header.Get("Retry-After") == "") {
			t.Fatalf("T-W9 lần 6 phải 429 + Retry-After: %d %v", r.Status, r.Header)
		}
	}
	// T-A8: login sai lần 6 (đã có 1 lần đăng nhập admin trong setup → dùng IP khác qua XFF cuối)
	for i := 1; i <= 6; i++ {
		r := e.do("POST", "/api/auth/login", "", map[string]string{"username": "admin", "password": "sai"},
			"X-Forwarded-For", fmt.Sprintf("1.2.3.%d, 198.51.100.10", i))
		if i <= 5 && r.Status != 401 {
			t.Fatalf("login lần %d: %d", i, r.Status)
		}
		if i == 6 && (r.Status != 429 || r.Header.Get("Retry-After") == "") {
			t.Fatalf("T-A8 lần 6 phải 429: %d", r.Status)
		}
	}
	// Ảnh không bị tính vào rate limit chung
	for i := 0; i < 130; i++ {
		if r := e.do("GET", "/api/images/abc", "", nil); r.Status != 404 {
			t.Fatalf("ảnh bị rate limit ở lần %d: %d", i, r.Status)
		}
	}
}

func TestBootstrapAdmin(t *testing.T) {
	pool := tu.MigratedPool(t)
	ctx := context.Background()
	svc := tu.NewService(pool, tu.FixedNow)
	if _, err := svc.BootstrapAdmin(ctx, "", ""); err != service.ErrAdminEnvMissing {
		t.Fatalf("thiếu env khi chưa có admin phải lỗi, được %v", err)
	}
	if created, err := svc.BootstrapAdmin(ctx, "Admin", "matkhau-ban-dau"); err != nil || !created {
		t.Fatalf("tạo admin: %v %v", created, err)
	}
	// T-A10: restart với mật khẩu khác → không đổi
	if created, err := svc.BootstrapAdmin(ctx, "admin", "mat-khau-khac"); err != nil || created {
		t.Fatalf("lần 2 không được tạo lại: %v %v", created, err)
	}
	if _, err := svc.BootstrapAdmin(ctx, "", ""); err != nil {
		t.Fatalf("đã có admin thì không cần env: %v", err)
	}
	if _, err := svc.Login(ctx, service.LoginRequest{Username: "admin", Password: "matkhau-ban-dau"}); err != nil {
		t.Fatalf("MK ban đầu phải còn hiệu lực: %v", err)
	}
	if _, err := svc.Login(ctx, service.LoginRequest{Username: "admin", Password: "mat-khau-khac"}); err == nil {
		t.Fatal("MK mới từ env không được ghi đè")
	}
}
