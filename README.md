# WEB-APP-DONGHAI

Web app cho Công ty Đông Hải: trang giới thiệu & sản phẩm kiểu TMĐT, tra cứu bảo hành theo SĐT, khu vực nội bộ (admin/user) quản lý sản phẩm, khách hàng, đơn hàng, khuyến mãi.

| Phần | Công nghệ | Thư mục |
|---|---|---|
| Backend | Go 1.24 (chi, pgx, JWT, bcrypt) — REST JSON `/api/...` | `backend/` |
| Frontend | Vue 3 + Vite + Vue Router + Pinia (CSS thuần) | `frontend/` |
| Database | Supabase (Postgres) — ảnh sản phẩm lưu `bytea` | `backend/migrations/` |
| Triển khai | 1 Docker image (Go phục vụ API + file tĩnh Vue) trên Render | `Dockerfile`, `render.yaml` |

Tài liệu: [`docs/01-PHUONG-AN.md`](docs/01-PHUONG-AN.md) (phương án), [`docs/02-DUYET-VA-BAN-GIAO.md`](docs/02-DUYET-VA-BAN-GIAO.md) (checklist nghiệm thu), [`docs/03-GHI-CHU-DEV.md`](docs/03-GHI-CHU-DEV.md) (ghi chú hiện thực). Ảnh giao diện: [`docs/screenshots/`](docs/screenshots/).

## Chức năng chính
- **Trang chủ** `/`: hero, lưới sản phẩm kèm giá niêm yết, bấm SP mở modal (ảnh phóng to + mô tả + thông số; chia sẻ được bằng `/?sp=<slug>`), footer thông tin công ty + 4 nút Facebook/TikTok/Zalo/Shopee (link rỗng → nút ẩn).
- **Bảo hành** `/bao-hanh`: nhập SĐT (chấp nhận `0912 345 678`, `+84…`, `84…`) → các đơn đã mua, hạn bảo hành = ngày mua + 2 năm, còn/hết hạn. Tên/SĐT được che, có rate limit.
- **Nội bộ** `/noi-bo` (bắt buộc đăng nhập):
  - user: xem sản phẩm (kèm giá nội bộ), trang ẩn **Khuyến mãi & bảng giá** `/noi-bo/bang-gia`, xem/thêm khách hàng, xem/nhập đơn hàng, đổi mật khẩu.
  - admin: toàn bộ quyền của user + sửa/ẩn/xóa sản phẩm, upload ảnh, sửa thông tin công ty & link MXH, quản lý khuyến mãi, sửa/xóa khách & đơn, tạo/khóa/đặt lại mật khẩu tài khoản.
  - Trang `/noi-bo/bang-gia` không có link công khai; API `/api/internal/price-list` trả 401 khi chưa đăng nhập.

## Chạy local

Yêu cầu: Go 1.24 (luôn đặt `GOTOOLCHAIN=local`), Node 22, PostgreSQL 16+.

```bash
# 1) Postgres local (lệnh tiện dụng; hoặc dùng Postgres sẵn có)
make pg-start                       # dữ liệu ở /var/tmp/donghai-pg, cổng 54329, tạo DB donghai_dev + donghai_test

# 2) Cấu hình
cp .env.example .env                # sửa DATABASE_URL, JWT_SECRET (≥ 32 ký tự), ADMIN_PASSWORD
#   ví dụ DATABASE_URL=postgres://postgres@localhost:54329/donghai_dev?sslmode=disable
#   đặt SEED_DEMO=true để có dữ liệu mẫu (8 SP, 3 KM, 5 khách, 8 đơn)

# 3a) Chế độ phát triển: 2 terminal
make dev-api                        # Go :8080 (tự migrate, tạo admin, seed demo)
make dev-web                        # Vite :5173, proxy /api → :8080

# 3b) Hoặc chạy như production: build frontend rồi để Go phục vụ
make build                          # frontend/dist + backend/bin/server
cd backend && STATIC_DIR=../frontend/dist ./bin/server   # http://localhost:8080 (biến khác lấy từ env)
```

Đăng nhập `/dang-nhap` bằng `ADMIN_USERNAME`/`ADMIN_PASSWORD`; vào **Tài khoản** để tạo user nhân viên (vd `nhanvien1`).

## Biến môi trường

| Biến | Bắt buộc | Mặc định | Ghi chú |
|---|---|---|---|
| `DATABASE_URL` | Có | — | Supabase: dùng **Session pooler cổng 5432** (`postgresql://postgres.<ref>:<MẬT_KHẨU>@aws-0-<region>.pooler.supabase.com:5432/postgres?sslmode=require`). Nếu buộc dùng cổng 6543, thêm `default_query_exec_mode=simple_protocol` |
| `JWT_SECRET` | Có | — | ≥ 32 ký tự; thiếu/ngắn → server không khởi động |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | Lần đầu | — | Chỉ dùng khi DB chưa có admin; về sau đổi env **không** ghi đè mật khẩu |
| `PORT` | Không | `8080` | Render tự đặt |
| `STATIC_DIR` | Không | rỗng (chỉ API) | Docker đặt `/app/web` |
| `JWT_TTL_HOURS` | Không | `12` | |
| `MIGRATE_ON_START` | Không | `true` | Migration idempotent, có advisory lock |
| `SEED_DEMO` | Không | `false` | Seed dữ liệu mẫu khi bảng `products` rỗng |
| `ALLOWED_ORIGINS` | Không | rỗng | Chỉ cần khi frontend khác origin (dev); production cùng origin |
| `DB_MAX_CONNS` | Không | `5` | |
| `LOG_LEVEL` | Không | `info` | `debug` / `info` / `warn` / `error` |

Không commit `.env`. Trên Render, bí mật đặt `sync: false` / `generateValue`.

## Kiểm tra & test

```bash
cd backend
gofmt -l .                          # phải rỗng
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go test ./...     # unit test: SĐT, bảo hành, JWT/bcrypt, slug, middleware, config
# tích hợp với Postgres thật (tự bỏ qua nếu thiếu TEST_DATABASE_URL); mỗi test dùng schema riêng
TEST_DATABASE_URL='postgres://postgres@localhost:54329/donghai_test?sslmode=disable' \
  GOTOOLCHAIN=local go test -count=1 -tags=integration ./...
cd ../frontend && npm ci && npm run build
```
Hoặc gọn: `make check`.

## Triển khai (Render + Supabase)

1. **Database**: apply `backend/migrations/0001_init.sql`, `0002_rls.sql` lên Supabase (SQL Editor hoặc MCP) — không bắt buộc vì backend tự migrate khi khởi động. RLS bật trên mọi bảng, không tạo policy ⇒ anon key/PostgREST không đọc được dữ liệu; backend kết nối bằng user `postgres` (owner) nên không bị chặn.
2. **Render**: tạo Web Service từ repo (Blueprint `render.yaml`: Docker, gói free, Singapore, health check `/api/health`). Điền `DATABASE_URL`, `ADMIN_PASSWORD`; `JWT_SECRET` tự sinh.
3. Kiểm tra `https://<service>.onrender.com/api/health` → `{"status":"ok"}`.

Build image thủ công: `docker build -t web-app-donghai .` rồi
`docker run -p 8080:8080 -e DATABASE_URL=... -e JWT_SECRET=... -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=... web-app-donghai`.

Lưu ý gói miễn phí: Render ngủ sau ~15 phút không có truy cập (mở lại lần đầu chậm 30–60 giây); Supabase free tạm dừng project sau 7 ngày không hoạt động (vào Dashboard bấm Restore).

## Vận hành

- **Thay link mạng xã hội / thông tin công ty**: đăng nhập admin → **Thông tin công ty**. Link phải bắt đầu bằng `https://`; để trống thì nút tương ứng ẩn ở footer. Dữ liệu demo dùng link mẫu (`https://facebook.com/`, `https://www.tiktok.com/`, `https://zalo.me/`, `https://shopee.vn/`) — cần thay link thật.
- **Ảnh sản phẩm**: admin → Sản phẩm → Sửa → tải ảnh (JPEG/PNG/WebP, ≤ 2 MB, ≤ 8 ảnh/SP; trình duyệt tự thu nhỏ về 1600px).
- **Khôi phục mật khẩu admin** (chỉ có 1 admin, không có API khôi phục) — chạy trên Supabase SQL Editor:
  ```sql
  UPDATE users SET password_hash = extensions.crypt('<MẬT_KHẨU_MỚI>', extensions.gen_salt('bf', 12)),
                   token_version = token_version + 1
  WHERE username = 'admin';
  ```
  (Postgres local: `CREATE EXTENSION pgcrypto` rồi dùng `crypt(...)`/`gen_salt(...)`.)
- User quên mật khẩu: admin vào **Tài khoản** → Đặt lại MK. Khóa tài khoản / đổi mật khẩu có hiệu lực ngay (token cũ bị từ chối).
