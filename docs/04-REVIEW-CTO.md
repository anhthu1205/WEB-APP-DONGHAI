# 04 — REVIEW CTO/BA: NGHIỆM THU & TRIỂN KHAI WEB-APP-DONGHAI

> Người viết: Agent 4 (CTO/BA). Ngày: 11/11/2026.
> Đầu vào: yêu cầu gốc của khách, `01-PHUONG-AN.md`, checklist mục 5 của `02-DUYET-VA-BAN-GIAO.md`, `03-GHI-CHU-DEV.md`, toàn bộ code `backend/` + `frontend/`.
> Nguyên tắc: không tin báo cáo DEV, tự đọc code và tự chạy lại mọi kiểm tra.

---

## 1. Kết luận

**ĐẠT CÓ ĐIỀU KIỆN.**

- Code đáp ứng đủ yêu cầu khách: trang chủ kiểu TMĐT + modal sản phẩm, footer 4 nút Facebook/TikTok/Zalo/Shopee, tra bảo hành theo SĐT (2 năm kể từ ngày mua), khu nội bộ bắt buộc đăng nhập, phân quyền admin/user, trang ẩn "Khuyến mãi & bảng giá". Go + Vue 3 + Supabase (Postgres), FE/BE tách bạch rõ ràng.
- Mọi ca **Chặn** kiểm được trong sandbox đều PASS. Đã sửa 1 lỗi bảo mật mức nhẹ (mục 3).
- Supabase: đã áp migration, RLS bật trên 10/10 bảng, anon/authenticated không có quyền, advisor không có cảnh báo mức ERROR.
- Render: đã tạo service https://led-donghai.onrender.com — **build Docker trên Render THÀNH CÔNG**; app dừng ở bước khởi động vì chưa có `DATABASE_URL` (đúng dự kiến).
- **Điều kiện còn lại** (khách phải làm, xem mục 6): nhập `DATABASE_URL` (chuỗi Session pooler Supabase) vào Render rồi Manual Deploy. Sau đó kiểm tra D3 (`/api/health` 200 trên onrender.com, IP client thật trong log).

---

## 2. Bảng checklist đã kiểm (02 §5)

Môi trường kiểm: Go 1.24.7 (`GOTOOLCHAIN=local`), Node 22, PostgreSQL 16 local cổng 54329, binary chạy thật + `frontend/dist`, curl. Ký hiệu: ✅ PASS · ❌ FAIL · ⏳ chưa kiểm được (có lý do).

### 2.1 Build & chất lượng
| Ca | Kết quả | Ghi chú |
|---|---|---|
| Q1 gofmt / vet / test / test tích hợp | ✅ | `gofmt -l .` rỗng; `go vet ./...` và `-tags=integration` sạch; `go test ./...` xanh; `go test -tags=integration ./...` xanh (db, handler, auth, config, middleware, phone, slug, warranty) |
| Q2 phiên bản `go.mod`, `npm ci && npm run build` | ✅ | `go 1.24.0`, chi v5.5.0, pgx v5.8.0, jwt v5.3.1, x/crypto v0.48.0, x/time v0.14.0; build Vite OK |
| Q3 `docker build` | ✅ (trên Render) | Sandbox không build nguyên trạng được (Docker Hub 429 + CA proxy). Dockerfile gốc build thành công trên Render (5.2). `/api/health` trong container: chờ D3 |
| Q4 không secret, `.gitignore`, handler không có SQL | ✅ | `git ls-files` không có `.env`/binary; `grep SELECT/INSERT/UPDATE` trong `handler/` rỗng |

### 2.2 Nghiệp vụ đơn vị
| Ca | Kết quả | Ghi chú |
|---|---|---|
| T-P1…T-P4 (SĐT, che tên, slug) | ✅ | unit test + curl (`+84 912 345 678` → `091****678`, `N*** V*** An`) |
| T-W1…T-W4 (bảo hành 2 năm) | ✅ | Đọc `internal/warranty`: cộng 2 năm theo ngày dân sự, 29/02 → 28/02, ngày cuối vẫn còn hạn, "hôm nay" theo giờ VN; tzdata nhúng vào binary (distroless) |
| T-A1 JWT/bcrypt | ✅ | Chỉ chấp nhận HS256, bắt buộc `exp`; bcrypt cost 12 |

### 2.3 Auth & phân quyền
| Ca | Kết quả | Ghi chú |
|---|---|---|
| T-A2 price-list không token → 401 | ✅ | curl 401 JSON `UNAUTHORIZED` |
| T-A3 token rác → 401 | ✅ | |
| T-A4 user gọi API admin → 403 | ✅ | 7 endpoint admin (gồm `GET/PATCH /api/admin/users`) đều 403. Phân quyền ở backend: cả nhóm `/api/admin` qua `RequireAuth` + `RequireAdmin`, role đọc lại từ DB (không tin claim) |
| T-A5 user nhập khách/đơn → 201 | ✅ | test tích hợp |
| T-A6 khóa user → token cũ 401 ngay | ✅ | curl: khóa xong token cũ của user → 401 (`token_version` + `is_active` kiểm mỗi request) |
| T-A7 đổi/reset MK vô hiệu token cũ | ✅ | test tích hợp |
| T-A8 login 429, XFF đầu chuỗi không né được | ✅ | curl: lần 6 → 429 có `Retry-After` dù đổi `X-Forwarded-For: 1.2.3.n, 5.5.5.5`. Xem lưu ý 4.2 |
| T-A9 không tự khóa; `role` trong PATCH → 400 | ✅ | curl 400 |
| T-A10 bootstrap admin từ env, không ghi đè | ✅ | test tích hợp + log khởi động |

### 2.4 API công khai
| Ca | Kết quả | Ghi chú |
|---|---|---|
| T-PUB1 không lộ `price_internal`/`promo`/`is_visible`, 7 SP | ✅ | DTO công khai riêng (`PublicProductDTO`), grep JSON = 0 |
| T-PUB2 SP ẩn → 404 | ✅ | |
| T-PUB3, T-PUB4 | ✅ | test tích hợp |
| T-PUB5 ảnh: cache immutable, ETag/304, `abc` → 404 | ✅ | curl `abc` → 404 |
| T-PUB6 `/api/lạ` → 404 JSON; SPA fallback | ✅ | `/bao-hanh`, `/noi-bo/bang-gia` → 200 `text/html` |
| T-W5 tra `+84 912 345 678` → 2 đơn An, che SĐT/tên | ✅ | DH000008 (05/11/2026 → 05/11/2028), DH000001 (10/03/2025 → 10/03/2027, còn 119 ngày) |
| T-W6 Bình 2024-08-15 → 2026-08-15 expired | ✅ | curl |
| T-W7 không có địa chỉ/giá/ghi chú/người nhập | ✅ | `WarrantyItemDTO` chỉ có tên SP, SL, serial |
| T-W8 SĐT `abc` → 400; không có khách → `orders: []` | ✅ | |
| T-W9 lần 6/phút cùng IP → 429 | ✅ | curl |

### 2.5 API nội bộ / admin / ảnh
| Ca | Kết quả | Ghi chú |
|---|---|---|
| T-I1…T-I8 | ✅ | test tích hợp; curl T-I2: KM "cuối năm 2026" + "Tết Đinh Mùi 2027", không có "Hè 2026"; SP1 `current_promo_price`=5490000, SP2 null |
| T-AD1…T-AD10 | ✅ | test tích hợp; curl `facebook_url: "javascript:..."` → 400 |
| T-IMG1, T-IMG4, T-IMG5 | ✅ | test tích hợp |
| T-IMG2 3 MB → 413 | ✅ | curl |
| T-IMG3 text giả `.jpg` → 415 | ✅ | curl; kiểm theo bytes thật (`DetectContentType` + `DecodeConfig`) |
| T-IMG6 user upload → 403 | ✅ | curl |

### 2.6 Bảo mật chung
| Ca | Kết quả | Ghi chú |
|---|---|---|
| T-S1 header bảo mật | ✅ | nosniff, `X-Frame-Options: DENY`, Referrer-Policy, CSP `default-src 'self'` |
| T-S2 body > 1 MB / JSON hỏng / trường lạ | ✅ | `DisallowUnknownFields`, `MaxBytesReader` |
| T-S3 không CORS khi `ALLOWED_ORIGINS` rỗng | ✅ | không có header `Access-Control-*` |
| T-S4 thiếu env → không khởi động | ✅ | unit test config, thoát mã 2 |
| T-S5 log không có mật khẩu/token/DSN | ✅ | grep log chạy thật: 0 dòng chứa mật khẩu; chỉ log path, không log query |
| T-S6 bundle công khai không có `price_internal` | ✅ | `grep` `dist/assets/index-*.js` rỗng; view nội bộ lazy-load |
| T-S7 Supabase RLS, owner, advisors | ✅ | mục 5.1 |
| SQL injection | ✅ | Toàn bộ truy vấn tham số hóa (`$1…`); chuỗi tìm kiếm escape `%`/`_`; thử `q=' OR 1=1--` → 200 bình thường |
| XSS | ✅ | không có `v-html`; link MXH bắt buộc `https://` (chặn `javascript:`), `rel="noopener noreferrer"` |

### 2.7 Giao diện (xem `docs/screenshots/`)
| Ca | Kết quả | Ghi chú |
|---|---|---|
| T-FE1 trang chủ TMĐT, giá `5.990.000 ₫`, không link bảng giá | ✅ | hero + ô tra bảo hành nhanh, lưới 4 cột, card có giá đỏ nổi bật |
| T-FE2 modal ảnh lớn + chi tiết + thông số | ✅ | có nút "Liên hệ mua", "Chat Zalo"; `?sp=slug` |
| T-FE3 footer + 4 nút MXH | ✅ | Facebook/TikTok/Zalo/Shopee đúng màu thương hiệu, mở tab mới |
| T-FE4 trang bảo hành | ✅ | ngày `dd/MM/yyyy`, badge "Còn bảo hành · còn N ngày" |
| T-FE5, T-FE6 guard đăng nhập / quyền | ✅ | đọc `router/index.js`: `requiresAuth` → `/dang-nhap?redirect=`; `role: 'admin'` → về Tổng quan + toast |
| T-FE7, T-FE8 | ✅ (theo ảnh + code) | form 2 bước, mobile 375px không tràn ngang. Chưa kiểm thủ công bằng trình duyệt ở vòng này; Lighthouse (NÊN) chưa chạy |

Đánh giá UI: đúng tinh thần "web thương mại điện tử", sạch, nhất quán, tiếng Việt hiển thị dấu tốt. Ảnh sản phẩm hiện là placeholder vì dữ liệu demo chưa có ảnh — khách cần tải ảnh thật (mục 6).

### 2.8 Triển khai
| Ca | Kết quả | Ghi chú |
|---|---|---|
| D1 migration Supabase, `schema_migrations` 2 version | ✅ | mục 5.1 |
| D2 Render Docker free Singapore, repo/main | ✅ | mục 5.2 |
| D3 `/api/health` 200 trên onrender.com, IP client thật | ⏳ | Chờ khách nhập `DATABASE_URL` (không có mật khẩu DB trong phiên) |
| D4 README đủ mục | ✅ | chạy local, env, test, MXH, khôi phục MK admin, lưu ý gói free |

---

## 3. Lỗi đã sửa

| # | Mức | Vấn đề | Sửa |
|---|---|---|---|
| F1 | Bảo mật (nhẹ) | `BurnPasswordCheck` so khớp với hash giả cost `MinCost` (4) trong khi hash thật cost 12 → đăng nhập username không tồn tại trả lời nhanh hơn rõ rệt ⇒ dò được username qua thời gian phản hồi | `internal/auth/password.go`: hash giả sinh lười (`sync.Once`) với đúng `BcryptCost`. Chạy lại gofmt/vet/test |

Không phát hiện lỗi chặn nào khác.

---

## 4. Đánh giá điểm lệch của DEV (03 §1) & lưu ý

### 4.1 Điểm lệch L1–L15
| # | Kết luận | Nhận xét |
|---|---|---|
| L1 font Be Vietnam Pro tự host | Chấp nhận | Cùng origin, đúng CSP, cải thiện hiển thị tiếng Việt |
| L2, L3, L6 tên module/package/component | Chấp nhận | Quy ước hợp lý |
| L4 schema riêng mỗi test | Chấp nhận | Tốt hơn phương án |
| L5 cổng 54329 | Chấp nhận | Chỉ ảnh hưởng local |
| L7 route sửa đơn cho admin | Chấp nhận | Bổ sung UI còn thiếu cho API sẵn có |
| L8 bảng giá chỉ SP đang hiển thị | Chấp nhận | Hợp lý nghiệp vụ; admin vẫn thấy SP ẩn ở trang Sản phẩm |
| L9 sai MK hiện tại → 400 | Chấp nhận | Tránh FE tự đăng xuất |
| L10 bỏ trống tên khi SĐT đã có | Chấp nhận | Đúng tinh thần "dùng lại khách" |
| L11 `items_count` = tổng số lượng | Chấp nhận | |
| L12 giữ slug khi PUT không gửi slug | Chấp nhận | Giữ link chia sẻ |
| L13 ô tra bảo hành ở hero, SĐT qua `history.state` | Chấp nhận | Không đưa SĐT lên URL — tốt |
| L14 bỏ dòng `# syntax=` | Chấp nhận | Không ảnh hưởng Render |
| L15 log chỉ path | Chấp nhận | Tốt cho riêng tư |

### 4.2 Lưu ý kỹ thuật (không chặn)
- **IP client sau proxy (D3)**: `ClientIP` ưu tiên `CF-Connecting-IP`, rồi phần tử **cuối** của `X-Forwarded-For`. Đúng khi chạy sau proxy của Render (Cloudflare ghi đè `CF-Connecting-IP`). Nếu chạy **trực tiếp** ra Internet không qua proxy thì client tự gửi XFF đổi được IP (đã thử local: đổi `X-Forwarded-For` mỗi lần → không bị 429). Không áp dụng cho Render, nhưng cần xác nhận trường `ip` trong log Render khi app chạy (D3).
- Advisor Supabase `function_search_path_mutable` (WARN) cho `public.set_updated_at` — rủi ro thấp (hàm trigger không `SECURITY DEFINER`). Có thể thêm migration `0003` đặt `SET search_path = public` ở đợt sau.
- Advisor `rls_enabled_no_policy` (INFO) — **cố ý**: không tạo policy để anon/authenticated (PostgREST) không đọc được gì; backend dùng role `postgres` (owner, `rolbypassrls = true`).
- JWT lưu `localStorage`: chấp nhận được nhờ CSP chặt + không `v-html`; TTL 12 giờ.
- Ảnh của SP ẩn vẫn tải được nếu biết uuid — đúng thiết kế (uuid ngẫu nhiên, không dò được).

---

## 5. Trạng thái triển khai

### 5.1 Supabase (project `lzmhasrctmekzlkngrah`)
- Trước khi áp: schema `public` **trống**, chưa có migration → không xung đột.
- Đã áp **y nguyên** `backend/migrations/0001_init.sql` và `0002_rls.sql` (qua MCP `apply_migration`, tên `0001_init`, `0002_rls`).
- Đồng bộ với runner của app: chính 2 file SQL tự `INSERT INTO schema_migrations (version) VALUES ('0001_init' | '0002_rls')` — trùng đúng tên version mà runner (`internal/db/migrate.go`, version = tên file bỏ `.sql`) dùng để bỏ qua. Đã xác nhận `schema_migrations` có đúng 2 dòng. ⇒ Khi app khởi động trên Render với `MIGRATE_ON_START=true`, runner thấy đủ 2 version và **không chạy lại**; kể cả nếu chạy lại thì SQL cũng idempotent (đã kiểm chạy 2 lần trên PG16). Cách này an toàn hơn để app tự migrate lần đầu vì có thể kiểm RLS/advisor ngay, trước khi có dữ liệu.
- RLS: 10/10 bảng `relrowsecurity = true`; owner `postgres`; `postgres.rolbypassrls = true`; `anon`/`authenticated` **không** có quyền SELECT/INSERT/UPDATE/DELETE trên bảng nào.
- `get_advisors` (security): 0 ERROR; 1 WARN, 10 INFO như 4.2.
- **Không** nạp dữ liệu demo bằng tay. Dữ liệu demo sẽ do app tự seed ở lần khởi động đầu (`SEED_DEMO=true`, chỉ seed khi bảng `products` rỗng).

### 5.2 Render
| Mục | Giá trị |
|---|---|
| Service | `web-app-donghai` (id `srv-dbq4hjpsrm7s739g61jg`) |
| URL | **https://led-donghai.onrender.com** |
| Dashboard | https://dashboard.render.com/web/srv-dbq4hjpsrm7s739g61jg |
| Cấu hình | Web Service, runtime Docker (`./Dockerfile`), gói **free**, region **singapore**, repo `anhthu1205/WEB-APP-DONGHAI` nhánh `main`, autoDeploy bật, health check `/api/health` |
| Biến môi trường | `JWT_SECRET` (64 ký tự ngẫu nhiên), `ADMIN_USERNAME=admin`, `ADMIN_PASSWORD` (ngẫu nhiên, đã giao riêng cho khách — **không** ghi trong repo), `JWT_TTL_HOURS=12`, `MIGRATE_ON_START=true`, `SEED_DEMO=true`, `DB_MAX_CONNS=5`, `LOG_LEVEL=info`, `DATABASE_URL` = **rỗng (khách tự nhập)** |
| Deploy đầu (`dep-dbq4hkhsrm7s739g639g`) | **Build Docker: THÀNH CÔNG** (sự kiện `build_ended: succeeded`, 3 stage node/golang/distroless, ~1 phút). **Khởi động: thất bại như dự kiến** — log: `cấu hình không hợp lệ: thiếu DATABASE_URL`, `Exited with status 2` |

Deploy đầu **dự kiến thất bại ở bước khởi động** vì `DATABASE_URL` rỗng (server in `cấu hình không hợp lệ: thiếu DATABASE_URL` rồi thoát mã 2). Đây là điều đã lường trước, không phải lỗi code.

---

## 6. Hướng dẫn cho khách hoàn tất (từng bước)

1. **Lấy chuỗi kết nối Supabase (Session pooler)**
   - Vào https://supabase.com/dashboard → chọn project **WEB-APP-DONGHAI**.
   - Bấm nút **Connect** (đầu trang) → tab **Connection String** → chọn **Session pooler** (cổng **5432**). *Không* dùng "Direct connection" (Render gói free không hỗ trợ IPv6).
   - Chuỗi có dạng `postgresql://postgres.lzmhasrctmekzlkngrah:[YOUR-PASSWORD]@aws-0-ap-northeast-2.pooler.supabase.com:5432/postgres`.
   - Thay `[YOUR-PASSWORD]` bằng mật khẩu database (quên thì: **Project Settings → Database → Reset database password**). Nếu mật khẩu có ký tự đặc biệt (`@ : / ? #`…) thì nên đặt lại mật khẩu chỉ gồm chữ + số cho dễ.
   - Thêm `?sslmode=require` vào cuối chuỗi.
2. **Dán vào Render**: https://dashboard.render.com/web/srv-dbq4hjpsrm7s739g61jg → **Environment** → dòng `DATABASE_URL` → **Edit** → dán chuỗi → **Save Changes**.
3. **Deploy**: chọn **Save, rebuild, and deploy** (hoặc **Manual Deploy → Deploy latest commit**). Đợi 3–5 phút đến khi trạng thái **Live**. Kiểm tra https://led-donghai.onrender.com/api/health trả `{"status":"ok"}`.
4. **Đăng nhập admin**: mở https://led-donghai.onrender.com/dang-nhap (link "Đăng nhập nội bộ" ở chân trang), tên đăng nhập `admin`, mật khẩu ban đầu được gửi riêng (không nằm trong tài liệu/repo; cũng xem được ở Render → Environment → `ADMIN_PASSWORD`).
5. **Đổi mật khẩu ngay**: menu **Đổi mật khẩu**. Sau đó có thể xóa biến `ADMIN_PASSWORD` trên Render (chỉ dùng ở lần đầu; đổi env về sau **không** ghi đè mật khẩu).
6. **Dữ liệu demo**: lần đầu app nạp 8 SP, 3 KM, 5 khách, 8 đơn mẫu để xem thử. Muốn dùng thật: đặt `SEED_DEMO=false` trên Render (để không seed lại), rồi xóa/sửa dữ liệu mẫu trong khu nội bộ (hoặc nhờ kỹ thuật xóa bằng SQL trên Supabase). Lưu ý: seed chỉ chạy khi bảng sản phẩm rỗng.
7. **Thông tin công ty & mạng xã hội**: admin → **Thông tin công ty** → nhập tên, địa chỉ, SĐT, email, giờ làm việc và link thật Facebook/TikTok/Zalo/Shopee (bắt buộc bắt đầu bằng `https://`; để trống thì nút ẩn).
8. **Sản phẩm & ảnh thật**: admin → **Sản phẩm** → Sửa → tải ảnh (JPEG/PNG/WebP ≤ 2 MB, tối đa 8 ảnh/SP, chọn ảnh chính), sửa giá niêm yết/giá nội bộ, ẩn/hiện SP.
9. **Tạo tài khoản nhân viên**: admin → **Tài khoản** → Tạo (role user: xem + nhập khách/đơn, xem trang Khuyến mãi & bảng giá).

---

## 7. Việc nên làm tiếp
- (Kỹ thuật, sau khi có `DATABASE_URL`) Kiểm D3: health 200, trường `ip` trong log Render là IP thật của người truy cập; thử 6 lần tra bảo hành/phút → 429.
- Gói free: Render ngủ sau ~15 phút (lần mở đầu chậm 30–60 s); Supabase free tạm dừng sau 7 ngày không hoạt động. Cân nhắc gói trả phí hoặc ping định kỳ khi chạy thật.
- Migration `0003`: `ALTER FUNCTION set_updated_at() SET search_path = public` để hết cảnh báo advisor.
- Sao lưu: bật/tải backup Supabase định kỳ (dữ liệu khách hàng + ảnh nằm trong DB).
- Vitest cho `formatVND`/`phone.js`, Lighthouse a11y (mục NÊN chưa làm).
- Tên miền riêng (vd `donghai.vn`) trỏ về Render khi sẵn sàng.

## Cập nhật 2026-11-12: đổi địa chỉ web
- Địa chỉ chính thức: **https://led-donghai.onrender.com** (service Render `led-donghai`, id `srv-dbqh933rjlhs7389oo10`, cùng repo/nhánh `main`, cùng database Supabase và cùng biến môi trường với service cũ).
- Render không đổi được địa chỉ `*.onrender.com` của service đã tạo, nên đã tạo service mới. Service cũ `web-app-donghai` (https://web-app-donghai.onrender.com) đã được xóa theo yêu cầu của khách (2026-11-12).
