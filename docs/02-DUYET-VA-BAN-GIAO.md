# 02 — DUYỆT PHƯƠNG ÁN & BÀN GIAO DEV: WEB-APP-DONGHAI

> Người duyệt: Agent 2. Đầu vào: `docs/01-PHUONG-AN.md` (bản Agent 1) + yêu cầu gốc của khách.
> Người nhận: Agent 3 (DEV) — code theo **01 (đã sửa) + tài liệu này**. Agent 4 (CTO/BA) nghiệm thu theo **mục 5**.
> Khi 01 và 02 mâu thuẫn: **02 thắng**.

---

## 1. Kết luận duyệt

**DUYỆT CÓ ĐIỀU CHỈNH.** Phương án bao phủ đủ yêu cầu khách, kiến trúc gọn (1 container Go phục vụ API + file tĩnh Vue; FE/BE là 2 project độc lập, chỉ nói chuyện qua hợp đồng API mục 5 của 01), bảo mật đúng hướng (DTO công khai riêng, phân quyền ở backend, RLS chặn PostgREST). Đã sửa trực tiếp trong 01 các điểm dưới đây (xem "Lịch sử sửa đổi" cuối 01). Không còn điểm nào DEV phải đoán.

### 1.1 Đối chiếu yêu cầu gốc
| # | Yêu cầu khách | Mục trong 01 | Kết quả |
|---|---|---|---|
| 1 | Trang chủ kiểu TMĐT, SP kèm giá, thông tin công ty | F01, 5.2, 7.2 | Đủ |
| 2 | Bấm SP → modal ảnh phóng to + chi tiết | F02, 7.2 (gallery + lightbox) | Đủ |
| 3 | Footer: chi tiết công ty, địa chỉ, SĐT + nút Facebook/TikTok/Zalo/Shopee | F03, 7.2, R7 | Đủ — làm **4 nút** |
| 4 | Trang bảo hành: nhập SĐT → đơn đã mua, bảo hành 2 năm | F04, 4.3, 5.2 | Đủ — có che dữ liệu + rate limit |
| 5 | Nội bộ bắt buộc đăng nhập | 5.3, 6, 7.1 | Đủ |
| 6 | 1 admin full quyền, admin tạo user | F24, F26, 5.5 | Đủ (+ khóa/đặt lại MK) |
| 7 | User chỉ xem + nhập thông tin khách | F13, F14, F25, R8 | Đủ — **chốt mục 2.1** |
| 8 | Admin sửa chi tiết SP, đổi ảnh | F20, F21 | Đủ |
| 9 | Trang ẩn KM & giá cho người đăng nhập | F12, `/api/internal/price-list`, `/noi-bo/bang-gia` | Đủ — API 401 khi chưa đăng nhập, không link công khai, view lazy-load |
| 10 | Go + Vue + Supabase | 2.1–2.3 | Đủ — phiên bản đã chốt và kiểm chứng |
| 11 | FE/BE tách bạch, không làm ẩu | 2, 3 (quy tắc phân lớp) | Đủ — bổ sung quy ước code mục 4 |

---

## 2. Các điểm đã chốt / đã sửa

### 2.1 Quyết định nghiệp vụ
1. **User = xem + THÊM khách hàng/đơn hàng; sửa/xóa khách & đơn là của admin.** Đúng nguyên văn khách: user "chỉ có quyền xem, nhập thông tin khách". User xem được giá nội bộ và trang KM (đó là mục đích trang ẩn). User **không** sửa SP/ảnh/công ty/KM/tài khoản.
2. Chỉ **1 admin** (tạo từ env lần đầu); API tạo tài khoản chỉ tạo role `user`; không có trường `role` trong PATCH user.
3. Bảo hành 2 năm tính theo **ngày mua của đơn**, áp dụng cho mọi dòng; ngày hết hạn tính cả ngày cuối (còn bảo hành đến hết ngày `end`); 29/02 → 28/02.
4. Phạm vi gọn — **cắt bỏ**: API sắp xếp ảnh, thư viện `x/image/webp`, vuốt cảm ứng trong lightbox (không bắt buộc), ESLint/Prettier (không bắt buộc). Vitest frontend: **NÊN**, không chặn nghiệm thu.

### 2.2 Sửa kỹ thuật (đã ghi vào 01)
| # | Vấn đề phát hiện | Mức | Đã chốt |
|---|---|---|---|
| K1 | `go get @latest` sẽ kéo pgx v5.12 / x/crypto v0.58 / x/time v0.16 đòi **Go ≥ 1.25/1.26** → vỡ build trên Go 1.24.7 và image `golang:1.24` | Chặn | Ghim: chi v5.5.0, pgx v5.8.0, jwt v5.3.1, x/crypto v0.48.0, x/time v0.14.0; `GOTOOLCHAIN=local`. Đã kiểm `go vet` + `go run` OK |
| K2 | Runner migration đọc `schema_migrations` trước khi 0001 tạo bảng → lỗi lần chạy đầu; advisory lock trên pool có thể lock/unlock ở 2 kết nối khác nhau | Chặn | Runner tự `CREATE TABLE IF NOT EXISTS schema_migrations` trước; lock/unlock trên 1 `pool.Acquire()` |
| K3 | `chi middleware.RealIP` tin header client gửi → né được rate limit tra bảo hành/đăng nhập | Bảo mật | Tự viết `clientIP`: `CF-Connecting-IP` → phần tử cuối `X-Forwarded-For` → `RemoteAddr` |
| K4 | So sánh ngày `date` (UTC) với "hôm nay" giờ VN dễ lệch 1 ngày | Đúng/sai | Quy về ngày dân sự UTC 00:00; `today` inject được |
| K5 | `/api/images/abc` → lỗi Postgres 22P02 → 500 | Nhỏ | Validate uuid → 404; hỗ trợ `If-None-Match` → 304 |
| K6 | Ảnh trên trang chủ có thể ăn hạn mức 120 req/phút | Nhỏ | Loại `/api/images/*` khỏi rate limit chung |
| K7 | `DisallowUnknownFields` + FE gửi nguyên object từ GET → 400 | Nhỏ | FE gửi đúng tập trường request |
| K8 | Sai tên repo ở 9.2 | Nhỏ | `anhthu1205/WEB-APP-DONGHAI` |
| K9 | Supabase: backend phải không bị RLS chặn | Triển khai | Agent 4 kiểm `rolbypassrls`/owner sau khi apply |
| K10 | Admin duy nhất quên mật khẩu | Vận hành | R13: thủ tục SQL `extensions.crypt(...)` trong README |

Đã kiểm chứng: SQL `0001_init` + `0002_rls` chạy **2 lần liên tiếp** trên PostgreSQL 16 local không lỗi (idempotent); Postgres `date '2024-02-29' + interval '2 years'` = `2026-02-28` còn Go `AddDate(2,0,0)` = `2026-03-01` ⇒ bắt buộc clamp như 4.3; `http.DetectContentType` của Go 1.24 trả `image/webp`; npm: vite 6.4 / plugin-vue 5.2 / vue-router 4.6 / pinia 2.3 hỗ trợ Node 22.

---

## 3. Thứ tự triển khai (milestone) cho DEV

Làm tuần tự; **mỗi milestone xong → chạy tiêu chí → commit** (có thể nhiều commit/milestone). Không bắt đầu milestone sau khi tiêu chí milestone trước chưa đạt.

Chuẩn bị Postgres local cho test tích hợp (chạy với user `postgres`, thư mục data phải để user này truy cập được):
```bash
B=/usr/lib/postgresql/16/bin; D=/var/tmp/donghai-pg
mkdir -p $D && chown postgres $D
su postgres -s /bin/sh -c "$B/initdb -D $D -U postgres -A trust >/dev/null; $B/pg_ctl -D $D -o '-p 5432 -k /tmp' -l $D/log start -w"
psql -h /tmp -U postgres -c 'CREATE DATABASE donghai_test' -c 'CREATE DATABASE donghai_dev'
export TEST_DATABASE_URL='postgres://postgres@localhost:5432/donghai_test?sslmode=disable'
```

| M | Nội dung | Tiêu chí hoàn thành (đo được) |
|---|---|---|
| **M1** Khung dự án | `backend/go.mod` (phiên bản K1), `cmd/server/main.go` tối thiểu, `internal/config` (validate env fail-fast: thiếu `DATABASE_URL`/`JWT_SECRET` hoặc `JWT_SECRET` < 32 ký tự → thoát mã ≠ 0 với log rõ); `frontend/` Vite + Vue 3 + Router + Pinia, `vite.config.js` proxy `/api` → `:8080`; `.dockerignore`, `.env.example` (chỉ tên biến, giá trị giả) | `GOTOOLCHAIN=local go build ./... && go vet ./...` sạch; `gofmt -l .` rỗng; `npm ci && npm run build` OK; `package-lock.json` & `go.sum` được commit |
| **M2** Lõi nghiệp vụ thuần | `phone` (Normalize, Mask, MaskName), `warranty` (EndDate, Status), `auth` (JWT HS256, bcrypt), slugify tiếng Việt, `domain` errors | `go test ./internal/phone/... ./internal/warranty/... ./internal/auth/...` pass, đủ mọi ca ở 01 §9.1 + ca T-W1…T-W4, T-P1…T-P4 (mục 5) |
| **M3** CSDL | `migrations/0001_init.sql`, `0002_rls.sql` (y nguyên 01 §4.1–4.2), `embed.go`, runner (K2), `db.go` (pgxpool, retry 5 lần, `DB_MAX_CONNS`), bootstrap admin (F26), `seed/demo.sql` (01 §8), `testutil` | Test tích hợp: migrate DB trống 2 lần không lỗi, `schema_migrations` có đúng 2 dòng; seed → 8 SP (1 ẩn), 3 KM, 5 khách, 8 đơn; khởi động thiếu `ADMIN_*` khi chưa có admin → thoát; đã có admin → không ghi đè MK |
| **M4** Nền HTTP + Auth | `handler/router.go`, `respond.go` (định dạng lỗi 01 §5.1, map 23505/23503 → 409), middleware (RequestID, Recoverer, logger slog JSON có `request_id`+IP, secure headers, CORS khi có `ALLOWED_ORIGINS`, body limit 1 MB, rate limit + `clientIP` K3, RequireAuth/RequireAdmin), `/api/health`, `/api/auth/*`, 404 JSON cho `/api/*` lạ, static + SPA fallback, graceful shutdown | Ca T-A1…T-A8, T-S1…T-S5 pass (httptest/tích hợp) |
| **M5** API công khai | `/api/public/company`, `/products`, `/products/{slug}`, `/api/images/{uuid}`, `/api/public/warranty/lookup` | Ca T-PUB1…T-PUB6, T-W5…T-W9 pass |
| **M6** API nội bộ (user) | dashboard, products, price-list, promotions/{id}, customers (list, by-phone, detail, create), orders (list, detail, create — transaction) | Ca T-I1…T-I8 pass |
| **M7** API admin | products CRUD + visibility, images (upload/main/delete), company, promotions CRUD, sửa/xóa khách & đơn, users | Ca T-AD1…T-AD10, T-IMG1…T-IMG6 pass; `go test -tags=integration ./...` toàn bộ xanh |
| **M8** Frontend | Công khai: Home (lưới + modal + lightbox + `?sp=`), Footer 4 nút MXH, Bảo hành, Đăng nhập. Nội bộ: layout + menu theo quyền, Dashboard, Sản phẩm, Bảng giá (trang ẩn), Khách hàng, Đơn hàng (form 2 bước), Đổi MK; admin: Sửa SP + upload ảnh, KM, Công ty, Tài khoản | `npm run build` OK; T-FE1…T-FE8 kiểm thủ công trên `go run` + `vite`/bản build; không có `v-html` (`grep -r "v-html" frontend/src` rỗng) |
| **M9** Đóng gói & tài liệu | `Dockerfile` multi-stage (01 §2.3), `render.yaml`, README (chạy local, env, test, thay link MXH, R13 khôi phục MK admin) | `docker build` OK; `docker run` với Postgres local: `/api/health` 200, `/` và refresh `/bao-hanh` trả HTML, smoke 01 §9.1 chạy hết; ảnh image < 50 MB |

---

## 4. Quy ước code & commit

### 4.1 Backend Go
- Format/kiểm tra bắt buộc trước mỗi commit: `gofmt -l .` rỗng, `GOTOOLCHAIN=local go vet ./...` sạch, `go test ./...` xanh (tích hợp: `-tags=integration` với `TEST_DATABASE_URL`).
- Phân lớp: `handler` (HTTP, không SQL) → `service` (nghiệp vụ, validate, quyền nghiệp vụ) → `repository` (chỉ SQL tham số hóa). `ctx context.Context` là tham số đầu mọi hàm service/repository. Không biến toàn cục ngoài config bất biến.
- Đặt tên: package chữ thường một từ; file `snake_case.go`; kiểu xuất `PascalCase`; DTO request/response hậu tố `Request`/`DTO` (vd `CreateOrderRequest`, `PublicProductDTO`); JSON tag `snake_case`. Lỗi sentinel `ErrXxx` trong `domain`; `ValidationError{Fields map[string]string}`.
- Lỗi: luôn `{"error":{"code","message","fields"?}}` (01 §5.1); message tiếng Việt cho người dùng; không trả lỗi SQL/stack; wrap lỗi bằng `fmt.Errorf("...: %w", err)`; log lỗi 500 bằng `slog` kèm `request_id`. Không `panic` trong luồng xử lý request.
- Không log mật khẩu, token, `DATABASE_URL`, SĐT đầy đủ ở route tra bảo hành.
- Truy vấn danh sách SP/ảnh **không** SELECT cột `data`.

### 4.2 Frontend Vue
- Vue 3 `<script setup>`, Composition API; component `PascalCase.vue`; store Pinia `useXxxStore`; route path tiếng Việt không dấu như 01 §7.1.
- Mọi gọi API qua `src/api/http.js`; không gọi `fetch` trực tiếp trong view. Không hard-code URL backend (luôn `/api`).
- View nội bộ & admin **lazy-load**. Cấm `v-html`. Tiền: `formatVND`; ngày: `dd/MM/yyyy`.
- Request body chỉ chứa đúng các trường API định nghĩa (K7).
- `npm run build` không lỗi; ESLint không bắt buộc.

### 4.3 Commit
- Commit nhỏ theo milestone, nhánh `main`, message **tiếng Việt** dạng Conventional Commits: `<loại>(<phạm vi>): <mô tả>` với loại ∈ `feat|fix|docs|test|refactor|chore|build`, phạm vi ∈ `backend|frontend|db|deploy|docs`. Vd: `feat(backend): API tra cứu bảo hành theo SĐT (M5)`.
- Thân commit kết thúc bằng 2 dòng:
  ```
  Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_01TuE52fG4kBwoNBXhBncdXA
  ```
- Không commit: `.env*` (trừ `.env.example`), `node_modules`, `frontend/dist`, binary. Trước push: `git status` sạch, `git grep -nE "postgres(ql)?://[^:]+:[^@<]+@"` không ra chuỗi kết nối có mật khẩu thật.

---

## 5. Checklist nghiệm thu (Definition of Done) — Agent 4 dùng để review

Đánh dấu từng ca **PASS/FAIL**; mọi ca "Chặn" phải PASS mới được deploy. Lệnh mẫu dùng `B=http://localhost:8080`, `TU`/`TA` = token user/admin.

### 5.1 Build & chất lượng (Chặn)
- [ ] Q1 `cd backend && gofmt -l . ` rỗng; `GOTOOLCHAIN=local go vet ./...` sạch; `go test ./...` xanh; `go test -tags=integration ./...` xanh trên PG16 local.
- [ ] Q2 `go.mod` đúng phiên bản K1, `go 1.24.0`; `cd frontend && npm ci && npm run build` OK.
- [ ] Q3 `docker build .` OK; container chạy, `/api/health` → 200 `{"status":"ok"}`.
- [ ] Q4 Không secret trong repo; `.gitignore` chặn `.env*`; cấu trúc thư mục khớp 01 §3; handler không chứa SQL (`grep -rn "SELECT\|INSERT\|UPDATE " backend/internal/handler` rỗng).

### 5.2 Đơn vị nghiệp vụ (Chặn)
- [ ] T-P1 `Normalize("+84 912 345 678")`, `"84912345678"`, `"0084 912.345.678"`, `"0912-345-678"` → `0912345678`.
- [ ] T-P2 `Normalize("(024) 3826 1234")` → `02438261234`; lỗi với `""`, `"abc"`, `"12345"`, `"09123456789012"`, `"+84+912"`.
- [ ] T-P3 `Mask("0912345678")` = `091****678`; `MaskName("Nguyễn Văn An")` = `N*** V*** An`; `MaskName("An")` = `An`.
- [ ] T-P4 Slug: `"Máy lọc nước RO Đông Hải DH-9"` → `may-loc-nuoc-ro-dong-hai-dh-9`.
- [ ] T-W1 `EndDate(2025-03-10)` = `2027-03-10`; `EndDate(2024-12-31)` = `2026-12-31`.
- [ ] T-W2 `EndDate(2024-02-29)` = **`2026-02-28`**.
- [ ] T-W3 Mua 2024-11-11, today 2026-11-11 → `active`, `days_left=0` (**hết hạn đúng ngày: ngày cuối vẫn còn**).
- [ ] T-W4 Mua 2024-11-10, today 2026-11-11 → `expired`, `days_left=0`; mua 2025-03-10, today 2026-11-11 → `active`, `days_left=119`.
- [ ] T-A1 JWT hết hạn → lỗi; token ký `alg=none` hoặc HS256 khác secret → lỗi; bcrypt verify đúng/sai.

### 5.3 Auth & phân quyền (Chặn)
- [ ] T-A2 `GET /api/internal/price-list` **không token → 401** `{"error":{"code":"UNAUTHORIZED",...}}`.
- [ ] T-A3 Token rác / hết hạn → 401.
- [ ] T-A4 **User gọi API admin → 403** `FORBIDDEN`: `POST /api/admin/products`, `PUT /api/admin/company`, `POST /api/admin/users`, `PUT /api/admin/customers/1`, `DELETE /api/admin/orders/1`.
- [ ] T-A5 User gọi `POST /api/internal/customers` và `POST /api/internal/orders` → 201 (user được nhập).
- [ ] T-A6 Admin khóa user → token cũ của user → 401 ngay; đăng nhập user bị khóa → 401 cùng message "Sai tên đăng nhập hoặc mật khẩu".
- [ ] T-A7 Đổi MK → token cũ 401, token mới dùng được; admin reset MK user → token cũ user 401.
- [ ] T-A8 Đăng nhập sai lần thứ 6 liên tiếp (burst 5) trong 1 phút từ cùng IP → 429 có `Retry-After`; gửi `X-Forwarded-For: 1.2.3.<n>` khác nhau ở đầu chuỗi **không** né được limit.
- [ ] T-A9 Admin không thể khóa chính mình (400); `PATCH /api/admin/users/{id}` với `{"role":"admin"}` → 400 (trường lạ).
- [ ] T-A10 Khởi động lần đầu tạo admin từ env; restart với `ADMIN_PASSWORD` khác → MK admin **không** đổi.

### 5.4 API công khai (Chặn)
- [ ] T-PUB1 `GET /api/public/products` → JSON **không chứa** chuỗi `price_internal`, `promo`, `is_visible`; chỉ SP hiển thị (seed: 7 SP, không có `binh-nong-lanh-dh-wh30`).
- [ ] T-PUB2 `GET /api/public/products/binh-nong-lanh-dh-wh30` (SP ẩn) → 404.
- [ ] T-PUB3 `GET /api/public/products/may-loc-nuoc-ro-dh-9` có `description`, `specs`, `images`, `price` = 5990000; không có giá nội bộ/KM.
- [ ] T-PUB4 `GET /api/public/company` có `socials.facebook|tiktok|zalo|shopee`.
- [ ] T-PUB5 `GET /api/images/<uuid hợp lệ>` → 200, đúng `Content-Type`, `Cache-Control` có `immutable`, có `ETag`; gửi lại `If-None-Match` → 304; `GET /api/images/abc` → 404 (không 500); uuid không tồn tại → 404.
- [ ] T-PUB6 `GET /api/khong-ton-tai` → 404 JSON (không trả HTML); `GET /bao-hanh` và `/noi-bo/bang-gia` → trả `index.html` (SPA fallback).
- [ ] T-W5 **`POST /api/public/warranty/lookup {"phone":"+84912345678"}` → ra đơn của Nguyễn Văn An** (2 đơn seed), `phone_masked`=`091****678`, `customer_name_masked`=`N*** V*** An`, đơn sắp `purchase_date DESC`, `order_code` dạng `DH000xxx`.
- [ ] T-W6 Đơn seed của Trần Thị Bình 2024-08-15 → `warranty_end_date`=`2026-08-15`, `status`=`expired`; Lê Văn Cường 2024-02-29 → `2026-02-28`, `expired`.
- [ ] T-W7 Response tra bảo hành **không có** địa chỉ, ghi chú, giá, người nhập (`grep` JSON không chứa `address`, `price`, `note`, `created_by`).
- [ ] T-W8 SĐT hợp lệ nhưng không có khách → 200 `orders: []`; SĐT sai (`"abc"`) → 400 `VALIDATION_ERROR` có `fields.phone`.
- [ ] T-W9 Lần gọi tra cứu thứ 6 trong 1 phút từ cùng IP → 429 kèm `Retry-After`.

### 5.5 API nội bộ (Chặn)
- [ ] T-I1 `GET /api/internal/products` (TU) có `price_internal`, có cả SP ẩn.
- [ ] T-I2 `GET /api/internal/price-list` (TU) với today=2026-11-11: `promotions` gồm "Khuyến mãi cuối năm 2026" (đang chạy) + "Tết Đinh Mùi 2027" (sắp tới), **không** có "Hè 2026"; SP 1 có `current_promo_price`=5490000, SP 2 có `current_promo_price`=null.
- [ ] T-I3 Tạo khách SĐT `0912 345 678` (đã có) → 409 có `existing_id`.
- [ ] T-I4 `GET /api/internal/customers/by-phone?phone=%2B84987654321` → khách Trần Thị Bình; SĐT không có → 404.
- [ ] T-I5 Tạo đơn kèm `customer` mới → 201, khách mới được tạo, `created_by` = user hiện tại; tạo đơn với `customer.phone` đã tồn tại → `customer_reused:true`, tên/địa chỉ khách cũ **không** bị ghi đè.
- [ ] T-I6 Đơn có `unit_price` bỏ trống → lấy `price_public`; đơn 0 dòng hoặc > 50 dòng → 400; `purchase_date` ngày mai → 400; trước 2000-01-01 → 400.
- [ ] T-I7 Lỗi giữa chừng khi tạo đơn (vd `product_id` không tồn tại) → không để lại khách mới mồ côi (transaction).
- [ ] T-I8 Đơn vừa nhập tra được ngay qua `/api/public/warranty/lookup`.

### 5.6 API admin & ảnh (Chặn)
- [ ] T-AD1 Tạo SP không gửi slug → slug tự sinh; trùng tên → `-2`; slug sai định dạng → 400.
- [ ] T-AD2 `PATCH /products/{id}/visibility {"is_visible":false}` → biến mất khỏi API công khai.
- [ ] T-AD3 Xóa SP đã có đơn → 409; SP chưa có đơn → 204 (ảnh bị xóa theo).
- [ ] T-AD4 `PUT /api/admin/company` với `facebook_url:"http://..."` hoặc `"javascript:..."` → 400; `https://...` hoặc rỗng → 200, footer cập nhật.
- [ ] T-AD5 CRUD KM: `end_date < start_date` → 400; PUT thay toàn bộ items.
- [ ] T-AD6 Sửa khách đổi SĐT trùng khách khác → 409; xóa khách có đơn → 409.
- [ ] T-AD7 Sửa đơn thay items; xóa đơn → 204.
- [ ] T-AD8 Tạo user → role luôn `user`; username trùng → 409; username `"Ab"` hoặc MK < 8 ký tự → 400.
- [ ] T-AD9 Danh sách user không chứa `password_hash`/`token_version`.
- [ ] T-AD10 Mọi lỗi 4xx/5xx đúng định dạng `{"error":{...}}`, không lộ text lỗi SQL.
- [ ] T-IMG1 Upload JPEG/PNG/WebP hợp lệ ≤ 2 MB → 201; ảnh đầu tiên tự `is_main=true`.
- [ ] T-IMG2 **Upload file 3 MB → 413** `PAYLOAD_TOO_LARGE`.
- [ ] T-IMG3 File `.txt`/`.pdf`/SVG đổi đuôi thành `.jpg` (hoặc `Content-Type: image/jpeg` giả) → 415.
- [ ] T-IMG4 Ảnh thứ 9 của 1 SP → 400 (giới hạn 8).
- [ ] T-IMG5 `PATCH /api/admin/images/{uuid} {"is_main":true}` → chỉ còn đúng 1 ảnh chính; xóa ảnh chính → ảnh còn lại có `sort_order` nhỏ nhất thành chính.
- [ ] T-IMG6 User (không phải admin) upload ảnh → 403.

### 5.7 Bảo mật chung (Chặn)
- [ ] T-S1 Header trên mọi response: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy`, `Content-Security-Policy` như 01 §6.
- [ ] T-S2 Body JSON > 1 MB → 413; JSON hỏng → 400; trường lạ → 400.
- [ ] T-S3 Không có CORS header khi `ALLOWED_ORIGINS` rỗng.
- [ ] T-S4 Server không khởi động khi thiếu `DATABASE_URL`/`JWT_SECRET` hoặc `JWT_SECRET` < 32 ký tự.
- [ ] T-S5 Log không chứa mật khẩu/token/`DATABASE_URL`; lỗi 500 có `request_id`.
- [ ] T-S6 Bundle công khai: `grep -l "price_internal\|current_promo_price" frontend/dist/assets/index-*.js` rỗng (code trang ẩn và API nội bộ nằm ở chunk lazy riêng; chuỗi path route trong router được phép).
- [ ] T-S7 Supabase (sau apply): mọi bảng `relrowsecurity = true`; `postgres` bypass RLS hoặc là owner (K9); `get_advisors` không có cảnh báo mức ERROR.

### 5.8 Giao diện (kiểm thủ công — Chặn trừ khi ghi NÊN)
- [ ] T-FE1 Trang chủ: hero + lưới SP có ảnh/placeholder + giá `5.990.000 ₫`; không có giá nội bộ; không có link tới `/noi-bo/bang-gia`.
- [ ] T-FE2 Bấm SP → modal ảnh lớn + thumbnail + chi tiết + thông số; bấm ảnh → phóng to toàn màn hình; đóng bằng ×/Esc/bấm nền; URL `?sp=slug` mở lại đúng modal.
- [ ] T-FE3 Footer: tên, giới thiệu, địa chỉ, SĐT, email, giờ làm việc + **4 nút Facebook/TikTok/Zalo/Shopee** mở tab mới (`rel="noopener noreferrer"`); link rỗng → nút ẩn.
- [ ] T-FE4 Trang Bảo hành: nhập `+84 912 345 678` → hiện đơn, ngày hết hạn `dd/MM/yyyy`, badge còn/hết hạn; thông báo khi không có đơn và khi 429.
- [ ] T-FE5 Chưa đăng nhập vào `/noi-bo/bang-gia` → chuyển `/dang-nhap?redirect=...`; đăng nhập xong quay lại đúng trang.
- [ ] T-FE6 User: menu không có nhóm "Quản trị"; vào `/noi-bo/tai-khoan` bằng URL → bị đẩy về `/noi-bo` + toast; không thấy nút Sửa/Xóa khách, đơn, SP.
- [ ] T-FE7 Admin: sửa SP + upload/đặt ảnh chính/xóa ảnh → trang chủ cập nhật; sửa công ty + link MXH → footer cập nhật; tạo/khóa/reset MK user.
- [ ] T-FE8 Form nhập đơn 2 bước (SĐT → khách có sẵn/tạo mới → dòng SP, tổng tiền, hạn bảo hành dự kiến); responsive ở 375px không vỡ layout. (NÊN) Lighthouse Accessibility ≥ 90 trang chủ.

### 5.9 Triển khai (Agent 4)
- [ ] D1 Migration apply lên Supabase `lzmhasrctmekzlkngrah`, `schema_migrations` có 2 version; backend khởi động không chạy lại.
- [ ] D2 Render service Docker free Singapore từ `anhthu1205/WEB-APP-DONGHAI`/`main`; env theo 01 §9.2, secret `sync: false`/`generateValue`.
- [ ] D3 Khi có `DATABASE_URL`: `/api/health` 200 trên `*.onrender.com`; log request hiển thị IP client thật (K3) — nếu không, điều chỉnh `clientIP` và ghi lại trong 04.
- [ ] D4 README có: chạy local, biến môi trường, chạy test, thay link MXH, khôi phục MK admin (R13), lưu ý Render ngủ/Supabase tạm dừng.

---

## 6. Ngoài phạm vi (không làm ở đợt này)
Nhiều admin; phân quyền chi tiết theo chức năng; bảo hành khác nhau theo SP; giỏ hàng/thanh toán online; xuất Excel; Supabase Storage/Auth; refresh token; đa ngôn ngữ; domain riêng. Ghi nhận để đề xuất đợt sau.
