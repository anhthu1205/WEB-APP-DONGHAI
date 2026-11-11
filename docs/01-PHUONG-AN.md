# 01 — PHƯƠNG ÁN KỸ THUẬT: WEB-APP-DONGHAI

> Người lập: Agent 1 (Kiến trúc). Người duyệt: Agent 2. Tài liệu đủ chi tiết để DEV code thẳng.
> Quy ước: **PHẢI** = bắt buộc; **NÊN** = khuyến nghị; tiền = số nguyên VND (`bigint`), ngày = `YYYY-MM-DD`, múi giờ nghiệp vụ = `Asia/Ho_Chi_Minh`.

---

## 1. Tóm tắt yêu cầu → danh sách chức năng

### 1.1 Công khai (không đăng nhập)
| Mã | Chức năng | Ghi chú |
|---|---|---|
| F01 | Trang chủ kiểu TMĐT: hero (tên/slogan công ty), lưới sản phẩm đang hiển thị kèm **giá niêm yết** | Không bao giờ hiện giá nội bộ / KM |
| F02 | Bấm sản phẩm → modal: ảnh lớn (gallery, phóng to), mô tả chi tiết, thông số, giá niêm yết | Có thể mở trực tiếp qua `/?sp=<slug>` |
| F03 | Footer: tên, giới thiệu, địa chỉ, SĐT, email, giờ làm việc + **4 nút MXH**: Facebook, TikTok, Zalo, Shopee | Nút nào không có link thì ẩn |
| F04 | Trang Bảo hành: nhập SĐT → danh sách đơn đã mua, ngày mua, sản phẩm, **ngày hết hạn = ngày mua + 2 năm**, trạng thái Còn/Hết hạn, số ngày còn lại | Che bớt tên/SĐT, không trả địa chỉ; rate limit |
| F05 | Đăng nhập (username + mật khẩu) | Link "Đăng nhập nội bộ" nhỏ ở footer được phép; KHÔNG có link tới trang khuyến mãi |

### 1.2 User (đã đăng nhập, role `user` — admin cũng có toàn bộ quyền này)
| Mã | Chức năng |
|---|---|
| F10 | Dashboard nội bộ: số khách, số đơn tháng này, đơn gần đây |
| F11 | Xem danh sách sản phẩm (kể cả giá nội bộ), không sửa |
| F12 | **Trang ẩn "Khuyến mãi / Bảng giá nội bộ"**: CTKM đang/sắp diễn ra + bảng giá (niêm yết, nội bộ, giá KM hiện hành) |
| F13 | Khách hàng: tìm kiếm (tên/SĐT), xem chi tiết + lịch sử đơn, **thêm mới** |
| F14 | Đơn hàng: tìm kiếm/lọc theo ngày, xem chi tiết, **nhập đơn mới** (chọn khách theo SĐT hoặc tạo khách mới ngay trong form; nhiều dòng sản phẩm; serial tùy chọn) |
| F15 | Đổi mật khẩu của chính mình |

### 1.3 Admin (role `admin`)
| Mã | Chức năng |
|---|---|
| F20 | Sản phẩm: thêm, sửa (tên, slug, mô tả ngắn/chi tiết, thông số, giá niêm yết, giá nội bộ, đơn vị, thứ tự), **ẩn/hiện**, xóa (chỉ khi chưa có đơn tham chiếu) |
| F21 | Ảnh sản phẩm: upload (JPEG/PNG/WebP ≤ 2 MB, ≤ 8 ảnh/SP), đặt ảnh chính, sắp xếp, xóa |
| F22 | Sửa thông tin công ty (1 bản ghi) gồm 4 link MXH |
| F23 | Quản lý khuyến mãi: CRUD chương trình (tiêu đề, nội dung, từ ngày–đến ngày, bật/tắt) + danh sách SP áp dụng với giá KM |
| F24 | Quản lý tài khoản: liệt kê, tạo user, khóa/mở khóa, đặt lại mật khẩu |
| F25 | Sửa / xóa khách hàng và đơn hàng (user chỉ thêm + xem) |
| F26 | Admin ban đầu tự tạo từ env `ADMIN_USERNAME`/`ADMIN_PASSWORD` khi DB chưa có admin |

---

## 2. Kiến trúc tổng thể

```
Trình duyệt ──HTTPS──► Render Web Service (1 container Docker, gói free)
                         └─ Go binary (:PORT)
                              ├─ /api/*        → REST JSON (chi router)
                              ├─ /api/images/* → ảnh bytea từ DB (cache header)
                              └─ /*            → file tĩnh Vue (dist) + SPA fallback index.html
                                      │ pgxpool (TLS)
                                      ▼
                         Supabase Postgres 17 (project lzmhasrctmekzlkngrah, qua Session Pooler)
```

- **Monorepo, tách bạch**: `backend/` (Go module độc lập, không biết gì về Vue) và `frontend/` (Vite project độc lập, chỉ gọi `/api/...`). Hợp đồng duy nhất giữa hai bên = mục 5 của tài liệu này.
- Production: cùng origin → **không cần CORS**. Dev: Vite (`:5173`) proxy `/api` → Go (`:8080`); CORS chỉ bật khi đặt `ALLOWED_ORIGINS`.
- Không dùng Supabase Auth/Storage/PostgREST; Supabase chỉ là Postgres. Mọi truy cập dữ liệu qua backend Go.

### 2.1 Thư viện Go (go 1.24)
| Mục đích | Thư viện | Lý do |
|---|---|---|
| Router | `github.com/go-chi/chi/v5` + `chi/v5/middleware` (RequestID, RealIP, Recoverer, Timeout) | Nhẹ, tương thích `net/http`, group route theo quyền gọn |
| DB | `github.com/jackc/pgx/v5` (`pgxpool`) | Driver Postgres nhanh nhất, hỗ trợ `bytea`, `jsonb`, `date`, transaction |
| JWT | `github.com/golang-jwt/jwt/v5` | Chuẩn, hỗ trợ validate `exp`, ép thuật toán |
| Mật khẩu | `golang.org/x/crypto/bcrypt` (cost 12) | Chuẩn công nghiệp |
| Rate limit | `golang.org/x/time/rate` (token bucket theo IP, map + dọn định kỳ) | Không cần Redis cho 1 instance |
| Migration | **Tự viết runner nhỏ** + `embed` các file `migrations/*.sql`, bảng `schema_migrations` | 0 phụ thuộc thêm; file SQL thuần nên Agent 4 cũng apply được bằng Supabase MCP (xem 4.4) |
| Múi giờ | `import _ "time/tzdata"` | Image distroless không có tzdata |
| Test | `testing` chuẩn + `net/http/httptest`; test tích hợp dùng Postgres local qua `TEST_DATABASE_URL` | Không cần thêm framework |

Không dùng ORM (SQL viết tay trong repository, tham số hóa `$1..$n`).

### 2.2 Frontend
Vue 3 (Composition API, `<script setup>`), Vite 5/6, Vue Router 4, Pinia 2. **Không dùng UI framework** — CSS thuần với biến CSS (`src/assets/main.css`) để bundle nhỏ. HTTP: wrapper `fetch` tự viết (không cần axios). Các view nội bộ **lazy-load** (`() => import(...)`) để bundle công khai không chứa code nội bộ.

### 2.3 Triển khai
- `Dockerfile` multi-stage ở gốc repo:
  1. `node:22-alpine` → `cd frontend && npm ci && npm run build` → `frontend/dist`
  2. `golang:1.24-alpine` → `cd backend && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server`
  3. `gcr.io/distroless/static-debian12:nonroot` → copy `/out/server` → `/app/server`, `frontend/dist` → `/app/web`; `ENV STATIC_DIR=/app/web PORT=8080`; `USER nonroot`; `CMD ["/app/server"]`.
- `.dockerignore`: `**/node_modules`, `frontend/dist`, `.git`, `docs`, `*.env`.
- `render.yaml` (Blueprint): `type: web`, `runtime: docker`, `plan: free`, `region: singapore`, `healthCheckPath: /api/health`, env xem mục 9.2.
- Khi khởi động, server: đọc config → kết nối DB (retry 5 lần, backoff) → chạy migration (nếu `MIGRATE_ON_START=true`, mặc định true) → seed admin (F26) → seed demo (nếu `SEED_DEMO=true` và bảng `products` rỗng) → listen.

---

## 3. Cấu trúc thư mục

```
web-app-donghai/
├── README.md  .gitignore  .dockerignore  Dockerfile  render.yaml
├── docs/                     01-PHUONG-AN.md, 02-..., 04-...
├── backend/
│   ├── go.mod  go.sum
│   ├── cmd/server/main.go            # wiring: config → db → migrate → seed → router → ListenAndServe + graceful shutdown
│   ├── migrations/                   # SQL thuần, embed
│   │   ├── embed.go                  # //go:embed *.sql
│   │   ├── 0001_init.sql
│   │   └── 0002_rls.sql
│   ├── seed/demo.sql (embed)         # dữ liệu mẫu (mục 8)
│   └── internal/
│       ├── config/config.go          # đọc & validate env, fail-fast
│       ├── db/db.go  db/migrate.go   # pgxpool, runner migration
│       ├── domain/                   # struct thuần + lỗi nghiệp vụ (ErrNotFound, ErrConflict, ValidationError)
│       ├── phone/phone.go (+_test)   # Normalize, Mask
│       ├── warranty/warranty.go (+_test) # EndDate, Status
│       ├── auth/jwt.go  auth/password.go (+_test)
│       ├── repository/               # SQL: user_repo.go, company_repo.go, product_repo.go, image_repo.go,
│       │                             #      promotion_repo.go, customer_repo.go, order_repo.go
│       ├── service/                  # nghiệp vụ + validate: auth, user, company, product, image,
│       │                             #      promotion, customer, order, warranty_lookup, bootstrap(seed admin)
│       ├── handler/                  # HTTP: decode/encode JSON, gọi service; 1 file / nhóm route
│       │   ├── router.go             # khai báo toàn bộ route (mục 5)
│       │   ├── respond.go            # writeJSON, writeError (map lỗi domain → HTTP)
│       │   └── public.go auth.go internal_*.go admin_*.go image.go static.go
│       ├── middleware/               # auth.go (RequireAuth, RequireAdmin), ratelimit.go, cors.go,
│       │                             # bodylimit.go, secure_headers.go, logger.go (slog JSON)
│       └── testutil/                 # tạo DB test, reset schema
└── frontend/
    ├── package.json  vite.config.js  index.html
    ├── public/ favicon.svg  placeholder-product.svg
    └── src/
        ├── main.js  App.vue
        ├── api/        http.js (fetch wrapper: base '/api', gắn Bearer, parse lỗi, 401 → logout)
        │               public.js auth.js products.js customers.js orders.js promotions.js users.js company.js
        ├── stores/     auth.js (token, user, isAdmin; lưu localStorage) company.js (cache thông tin công ty)
        ├── router/     index.js (routes + guard beforeEach theo meta.requiresAuth / meta.role)
        ├── layouts/    PublicLayout.vue (header + footer)  InternalLayout.vue (sidebar menu theo quyền)
        ├── views/
        │   ├── public/   HomeView.vue  WarrantyView.vue  LoginView.vue  NotFoundView.vue
        │   └── internal/ DashboardView.vue  ProductsView.vue  PriceListView.vue  CustomersView.vue
        │                 CustomerDetailView.vue  OrdersView.vue  OrderFormView.vue  OrderDetailView.vue
        │                 ChangePasswordView.vue
        │                 admin/ ProductEditView.vue  PromotionsAdminView.vue  PromotionEditView.vue
        │                        CompanyEditView.vue  UsersView.vue
        ├── components/ ProductCard.vue  ProductModal.vue  ImageGallery.vue  SiteFooter.vue  SocialButtons.vue
        │               SiteHeader.vue  PhoneInput.vue  MoneyInput.vue  DataTable.vue  Pagination.vue
        │               ConfirmDialog.vue  ToastHost.vue  ImageUploader.vue  SpecsEditor.vue
        └── utils/      format.js (formatVND, formatDate dd/MM/yyyy)  phone.js (normalize phía client, chỉ để UX)
```

**Quy tắc phân lớp backend (PHẢI)**: handler không chứa SQL; repository không chứa logic quyền; service nhận `context.Context` + struct đầu vào, trả domain struct/lỗi domain. Handler đọc user hiện tại từ context (`middleware.UserFromContext`).

---

## 4. Thiết kế CSDL

### 4.1 `migrations/0001_init.sql`
```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
  version    text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at := now(); RETURN NEW; END $$;

-- Người dùng nội bộ
CREATE TABLE IF NOT EXISTS users (
  id             bigserial PRIMARY KEY,
  username       text NOT NULL UNIQUE CHECK (username ~ '^[a-z0-9_.]{3,32}$'),  -- lưu lowercase
  full_name      text NOT NULL DEFAULT '' CHECK (char_length(full_name) <= 100),
  password_hash  text NOT NULL,
  role           text NOT NULL CHECK (role IN ('admin','user')),
  is_active      boolean NOT NULL DEFAULT true,
  token_version  integer NOT NULL DEFAULT 0,      -- tăng khi khóa/đổi mật khẩu → vô hiệu JWT cũ
  last_login_at  timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Thông tin công ty: đúng 1 dòng
CREATE TABLE IF NOT EXISTS company_info (
  id            smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  name          text NOT NULL,
  slogan        text NOT NULL DEFAULT '',
  about         text NOT NULL DEFAULT '',
  address       text NOT NULL DEFAULT '',
  phone         text NOT NULL DEFAULT '',
  email         text NOT NULL DEFAULT '',
  working_hours text NOT NULL DEFAULT '',
  facebook_url  text NOT NULL DEFAULT '',
  tiktok_url    text NOT NULL DEFAULT '',
  zalo_url      text NOT NULL DEFAULT '',
  shopee_url    text NOT NULL DEFAULT '',
  updated_by    bigint REFERENCES users(id) ON DELETE SET NULL,
  updated_at    timestamptz NOT NULL DEFAULT now()
);
INSERT INTO company_info (id, name) VALUES (1, 'Công ty Đông Hải') ON CONFLICT (id) DO NOTHING;

-- Sản phẩm
CREATE TABLE IF NOT EXISTS products (
  id                bigserial PRIMARY KEY,
  name              text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
  slug              text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
  sku               text UNIQUE,
  short_description text NOT NULL DEFAULT '' CHECK (char_length(short_description) <= 500),
  description       text NOT NULL DEFAULT '' CHECK (char_length(description) <= 20000),
  specs             jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(specs) = 'array'), -- [{"label":"Công suất","value":"1500W"}]
  unit              text NOT NULL DEFAULT 'cái',
  price_public      bigint NOT NULL CHECK (price_public >= 0),     -- giá niêm yết công khai
  price_internal    bigint CHECK (price_internal >= 0),            -- giá nội bộ (NULL = chưa đặt)
  is_visible        boolean NOT NULL DEFAULT true,
  sort_order        integer NOT NULL DEFAULT 0,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_products_visible_order ON products (is_visible, sort_order, id);

-- Ảnh sản phẩm (lưu trong DB)
CREATE TABLE IF NOT EXISTS product_images (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),   -- uuid: không dò được ảnh của SP ẩn
  product_id  bigint NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  data        bytea NOT NULL,
  mime_type   text NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/webp')),
  size_bytes  integer NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 2097152),
  width       integer,     -- nếu decode được
  height      integer,
  is_main     boolean NOT NULL DEFAULT false,
  sort_order  integer NOT NULL DEFAULT 0,
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_product_images_product ON product_images (product_id, sort_order);
CREATE UNIQUE INDEX IF NOT EXISTS uq_product_images_main ON product_images (product_id) WHERE is_main;

-- Khuyến mãi
CREATE TABLE IF NOT EXISTS promotions (
  id          bigserial PRIMARY KEY,
  title       text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
  content     text NOT NULL DEFAULT '',
  start_date  date NOT NULL,
  end_date    date NOT NULL,
  is_active   boolean NOT NULL DEFAULT true,
  created_by  bigint REFERENCES users(id) ON DELETE SET NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  CHECK (end_date >= start_date)
);
CREATE INDEX IF NOT EXISTS idx_promotions_dates ON promotions (start_date, end_date);

CREATE TABLE IF NOT EXISTS promotion_items (
  promotion_id bigint NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
  product_id   bigint NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  promo_price  bigint NOT NULL CHECK (promo_price >= 0),
  note         text NOT NULL DEFAULT '',
  PRIMARY KEY (promotion_id, product_id)
);
CREATE INDEX IF NOT EXISTS idx_promotion_items_product ON promotion_items (product_id);

-- Khách hàng
CREATE TABLE IF NOT EXISTS customers (
  id          bigserial PRIMARY KEY,
  full_name   text NOT NULL CHECK (char_length(full_name) BETWEEN 1 AND 100),
  phone       text NOT NULL UNIQUE CHECK (phone ~ '^0[0-9]{9,10}$'),   -- đã chuẩn hóa
  address     text NOT NULL DEFAULT '' CHECK (char_length(address) <= 300),
  note        text NOT NULL DEFAULT '',
  created_by  bigint REFERENCES users(id) ON DELETE SET NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers (lower(full_name));

-- Đơn hàng
CREATE TABLE IF NOT EXISTS orders (
  id             bigserial PRIMARY KEY,
  customer_id    bigint NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
  purchase_date  date NOT NULL,
  note           text NOT NULL DEFAULT '' CHECK (char_length(note) <= 1000),
  created_by     bigint REFERENCES users(id) ON DELETE SET NULL,   -- người nhập
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders (customer_id, purchase_date DESC);
CREATE INDEX IF NOT EXISTS idx_orders_purchase_date ON orders (purchase_date DESC);

CREATE TABLE IF NOT EXISTS order_items (
  id            bigserial PRIMARY KEY,
  order_id      bigint NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id    bigint NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
  product_name  text NOT NULL,                 -- snapshot tên SP tại thời điểm bán
  quantity      integer NOT NULL CHECK (quantity BETWEEN 1 AND 10000),
  unit_price    bigint NOT NULL CHECK (unit_price >= 0),
  serial_number text CHECK (serial_number IS NULL OR char_length(serial_number) <= 100)
);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items (product_id);

-- Trigger updated_at
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['users','company_info','products','promotions','customers','orders'] LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS trg_%1$s_updated_at ON %1$s', t);
    EXECUTE format('CREATE TRIGGER trg_%1$s_updated_at BEFORE UPDATE ON %1$s FOR EACH ROW EXECUTE FUNCTION set_updated_at()', t);
  END LOOP;
END $$;

INSERT INTO schema_migrations (version) VALUES ('0001_init') ON CONFLICT DO NOTHING;
```

### 4.2 `migrations/0002_rls.sql`
```sql
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['schema_migrations','users','company_info','products','product_images',
                           'promotions','promotion_items','customers','orders','order_items'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
      EXECUTE format('REVOKE ALL ON TABLE %I FROM anon, authenticated', t);
    END IF;
  END LOOP;
END $$;
INSERT INTO schema_migrations (version) VALUES ('0002_rls') ON CONFLICT DO NOTHING;
```
**Giải thích RLS**: Supabase tự mở schema `public` qua REST (PostgREST) cho role `anon`/`authenticated` bằng anon key vốn là công khai. Bật RLS và **không tạo policy nào** ⇒ mọi truy vấn qua PostgREST trả 0 dòng / bị từ chối; `REVOKE` là lớp phòng thủ thứ hai. Backend Go kết nối bằng user `postgres` (chủ sở hữu bảng, không bật `FORCE RLS`) nên **không bị RLS chặn**. Mọi phân quyền nghiệp vụ nằm ở backend. Supabase Advisor sẽ báo mức INFO "RLS enabled, no policy" — đây là chủ đích. Khối `IF EXISTS pg_roles 'anon'` giúp chạy được cả trên Postgres local (không có role anon).

### 4.3 Quy tắc nghiệp vụ dữ liệu
**Chuẩn hóa SĐT VN** — `phone.Normalize(s string) (string, error)`:
1. Bỏ khoảng trắng, `.`, `-`, `(`, `)`; nếu còn ký tự khác ngoài chữ số và một dấu `+` đầu chuỗi → lỗi.
2. `+84xxxxxxxxx` → `0xxxxxxxxx`; `84xxxxxxxxx` (độ dài 11–12) → `0` + phần sau; `0084...` → `0...`.
3. Hợp lệ khi khớp `^0[0-9]{9,10}$` (di động 10 số, cố định 11 số). Ngược lại → `ErrInvalidPhone`.
4. Mọi nơi ghi/tìm theo SĐT (tạo khách, tìm khách, tra bảo hành) **PHẢI** gọi Normalize.

`phone.Mask("0912345678")` → `"091****678"` (giữ 3 đầu + 3 cuối). `MaskName("Nguyễn Văn An")` → `"N*** V*** An"` (giữ tên cuối, các từ khác chỉ giữ ký tự đầu; xử lý theo rune UTF-8).

**Bảo hành** — `warranty.EndDate(purchase time.Time) time.Time`: `purchase.AddDate(2,0,0)`; nếu ngày mua là 29/02 và năm đích không nhuận → **28/02** (không để Go tự trôi sang 01/03). `Status(end, today)`: `active` nếu `today <= end` (bảo hành hết ngày end), ngược lại `expired`; `days_left = max(0, end - today)` theo ngày. `today` lấy theo `Asia/Ho_Chi_Minh`. Hằng số `WarrantyYears = 2` trong package. Không lưu ngày hết hạn trong DB (tính ở backend).

**Khác**: `purchase_date` không được ở tương lai (> hôm nay VN) và không trước 2000-01-01. Đơn phải có ≥ 1 dòng, ≤ 50 dòng. `product_name`/`unit_price` mặc định lấy từ SP (unit_price mặc định = `price_public`, cho phép sửa). Slug tự sinh từ tên (bỏ dấu tiếng Việt, `đ`→`d`) nếu không gửi; trùng → thêm `-2`, `-3`. Ảnh chính: khi upload ảnh đầu tiên của SP → tự `is_main=true`; khi xóa ảnh chính → ảnh có `sort_order` nhỏ nhất thành chính (trong 1 transaction).

### 4.4 Migration runner
- Bảng `schema_migrations(version)`; runner đọc file embed theo thứ tự tên, bỏ qua version đã có, chạy mỗi file trong 1 transaction. Các file viết **idempotent** (`IF NOT EXISTS`, `ON CONFLICT`) và tự `INSERT` version ở cuối ⇒ Agent 4 có thể apply nội dung y nguyên bằng `mcp__Supabase__apply_migration`, sau đó backend khởi động sẽ không chạy lại.
- Dùng `pg_advisory_lock(727001)` quanh quá trình migrate để tránh chạy song song.

---

## 5. Đặc tả API

### 5.1 Quy ước chung
- Base `/api`, JSON UTF-8, `snake_case`. Thời gian `timestamptz` → RFC3339; ngày → `"2026-11-11"`; tiền → số nguyên VND.
- Auth: header `Authorization: Bearer <jwt>`.
- Phân trang: `?page=1&page_size=20` (tối đa 100) → `{"items":[...],"total":123,"page":1,"page_size":20}`.
- Lỗi (luôn dạng này, không bao giờ lộ lỗi SQL/stack):
```json
{"error":{"code":"VALIDATION_ERROR","message":"Dữ liệu không hợp lệ","fields":{"phone":"Số điện thoại không hợp lệ"}}}
```
| HTTP | code | Khi nào |
|---|---|---|
| 400 | `VALIDATION_ERROR` / `BAD_REQUEST` | Sai dữ liệu / JSON hỏng |
| 401 | `UNAUTHORIZED` | Thiếu/sai/hết hạn token, sai mật khẩu, user bị khóa |
| 403 | `FORBIDDEN` | User gọi API admin |
| 404 | `NOT_FOUND` | Không có tài nguyên (SP ẩn với public cũng 404) |
| 409 | `CONFLICT` | Trùng username/slug/SĐT, xóa SP đã có đơn |
| 413 | `PAYLOAD_TOO_LARGE` | Body/ảnh quá lớn |
| 415 | `UNSUPPORTED_MEDIA_TYPE` | Ảnh sai loại |
| 429 | `RATE_LIMITED` | Vượt rate limit (kèm header `Retry-After`) |
| 500 | `INTERNAL` | Lỗi khác — message chung "Đã có lỗi xảy ra", chi tiết chỉ ghi log kèm request_id |

- Nhóm quyền theo tiền tố: `/api/public/*` (ai cũng gọi), `/api/auth/*`, `/api/internal/*` (**RequireAuth**: user + admin), `/api/admin/*` (**RequireAuth + RequireAdmin**). Mọi route không khớp dưới `/api/` → 404 JSON (không rơi vào SPA fallback).

### 5.2 Công khai
| Method | Path | Mô tả |
|---|---|---|
| GET | `/api/health` | `{"status":"ok"}` (ping DB với timeout 2s; lỗi → 503) |
| GET | `/api/public/company` | Thông tin công ty |
| GET | `/api/public/products` | SP `is_visible=true`, sắp theo `sort_order, id`; hỗ trợ `?q=` |
| GET | `/api/public/products/{slug}` | Chi tiết SP hiển thị; SP ẩn → 404 |
| GET | `/api/images/{uuid}` | Trả bytes ảnh, `Content-Type` theo mime, `Cache-Control: public, max-age=31536000, immutable`, `ETag` = uuid |
| POST | `/api/public/warranty/lookup` | Tra bảo hành theo SĐT (rate limit) |

```jsonc
// GET /api/public/company
{"name":"Công ty TNHH Đông Hải","slogan":"...","about":"...","address":"...","phone":"0901234567",
 "email":"...","working_hours":"8:00 - 17:30 (T2 - T7)",
 "socials":{"facebook":"https://facebook.com/...","tiktok":"...","zalo":"https://zalo.me/0901234567","shopee":"..."}}

// GET /api/public/products  → mảng (không phân trang, số SP ít)
[{"id":1,"slug":"may-loc-nuoc-ro-dh-9","name":"Máy lọc nước RO DH-9","short_description":"...",
  "price":5990000,"unit":"cái","main_image_url":"/api/images/7c1e...-..."}]   // main_image_url null nếu chưa có ảnh

// GET /api/public/products/may-loc-nuoc-ro-dh-9
{"id":1,"slug":"...","name":"...","short_description":"...","description":"...",
 "specs":[{"label":"Công suất lọc","value":"10 lít/giờ"}],"price":5990000,"unit":"cái",
 "images":[{"url":"/api/images/7c1e...","is_main":true,"width":1200,"height":1200}]}
```
**PHẢI**: DTO công khai là struct riêng (`PublicProductDTO`) **không có** trường `price_internal`, `is_visible`, KM. Có test khẳng định JSON không chứa chuỗi `price_internal`/`promo`.

```jsonc
// POST /api/public/warranty/lookup   body: {"phone":"+84 912 345 678"}
// 200 (luôn 200 khi SĐT hợp lệ, kể cả không có đơn → orders: [])
{"phone_masked":"091****678","customer_name_masked":"N*** V*** An","warranty_years":2,"today":"2026-11-11",
 "orders":[{"order_code":"DH000012","purchase_date":"2025-03-10","warranty_end_date":"2027-03-10",
   "status":"active","days_left":484,
   "items":[{"product_name":"Máy lọc nước RO DH-9","quantity":1,"serial_number":"DH9-2503-0012"}]}]}
// 400 VALIDATION_ERROR nếu SĐT sai định dạng; 429 nếu vượt rate limit
```
Không trả: địa chỉ, ghi chú, giá, người nhập. `order_code` = `"DH" + lpad(id,6,'0')`. Đơn sắp theo `purchase_date DESC`.

### 5.3 Auth
| Method | Path | Quyền | Mô tả |
|---|---|---|---|
| POST | `/api/auth/login` | public (rate limit) | `{"username","password"}` → `{"token","expires_at","user":{"id","username","full_name","role"}}`; sai → 401 message chung "Sai tên đăng nhập hoặc mật khẩu" (kể cả user bị khóa — không tiết lộ) |
| GET | `/api/auth/me` | user | Thông tin user hiện tại |
| POST | `/api/auth/change-password` | user | `{"current_password","new_password"}` → 200 `{"token","expires_at"}`; tăng `token_version` (token cũ mất hiệu lực), client thay bằng token mới |

Đăng xuất: phía client xóa token (stateless). Không có refresh token; hết hạn → đăng nhập lại.

### 5.4 Nội bộ (user + admin) — `/api/internal`
| Method | Path | Mô tả |
|---|---|---|
| GET | `/dashboard` | `{"customers_total","orders_this_month","orders_total","recent_orders":[...5]}` |
| GET | `/products` | Tất cả SP (cả ẩn) kèm `price_public`, `price_internal`, `is_visible`, `main_image_url`; `?q=` |
| GET | `/products/{id}` | Chi tiết đầy đủ + images (`id`,`url`,`is_main`,`sort_order`) |
| GET | `/price-list` | **Trang ẩn**: `{"today":"...","promotions":[KM đang chạy + sắp tới, is_active=true, end_date>=today, kèm items],"products":[{"id","name","unit","price_public","price_internal","current_promo_price","current_promotion_title"}]}`. `current_promo_price` = giá KM thấp nhất trong các KM đang hiệu lực hôm nay, null nếu không có |
| GET | `/promotions/{id}` | Chi tiết 1 KM (kèm items + tên SP) |
| GET | `/customers` | `?q=` (tên ILIKE hoặc SĐT — nếu q chuẩn hóa được thành SĐT thì tìm khớp, ngược lại tìm phần số), phân trang |
| GET | `/customers/by-phone?phone=` | Tìm chính xác theo SĐT chuẩn hóa → 200 khách / 404 (dùng cho form nhập đơn) |
| GET | `/customers/{id}` | Khách + danh sách đơn (kèm bảo hành) |
| POST | `/customers` | Tạo khách `{"full_name","phone","address","note"}` → 201; trùng SĐT → 409 kèm `{"error":{..., "existing_id":12}}` |
| GET | `/orders` | `?q=` (tên/SĐT khách), `?from=&to=` (ngày mua), phân trang; mỗi dòng: `id, order_code, purchase_date, customer{id,full_name,phone}, total_amount, items_count, warranty_end_date, status, created_by_username` |
| GET | `/orders/{id}` | Chi tiết + items + bảo hành + người nhập |
| POST | `/orders` | Tạo đơn (dưới đây) → 201 |

```jsonc
// POST /api/internal/orders  — chọn 1 trong 2: customer_id HOẶC customer (tạo mới / dùng lại theo SĐT)
{"customer_id":null,
 "customer":{"full_name":"Nguyễn Văn An","phone":"0912 345 678","address":"Hải Phòng"},
 "purchase_date":"2026-11-10","note":"",
 "items":[{"product_id":1,"quantity":1,"unit_price":5990000,"serial_number":"DH9-2611-0001"},
          {"product_id":3,"quantity":2}]}           // unit_price bỏ trống → price_public
// Nếu "customer.phone" đã tồn tại → dùng khách đó (KHÔNG ghi đè tên/địa chỉ), response có "customer_reused":true
// 201 → {"id":45,"order_code":"DH000045","customer_reused":false, ...chi tiết đơn}
```
Tạo đơn + (nếu cần) khách trong **1 transaction**. `created_by` = user hiện tại.

### 5.5 Admin — `/api/admin`
| Method | Path | Mô tả |
|---|---|---|
| POST | `/products` | Tạo SP `{"name","slug"?,"sku"?,"short_description","description","specs":[{label,value}],"unit","price_public","price_internal","is_visible","sort_order"}` → 201 |
| PUT | `/products/{id}` | Cập nhật toàn bộ trường trên |
| PATCH | `/products/{id}/visibility` | `{"is_visible":false}` |
| DELETE | `/products/{id}` | 204; có `order_items` tham chiếu → 409 "Sản phẩm đã có đơn hàng, chỉ có thể ẩn" |
| POST | `/products/{id}/images` | `multipart/form-data`, field `file` (1 ảnh/lần) → 201 `{"id","url","is_main","sort_order"}` |
| PATCH | `/images/{uuid}` | `{"is_main":true}` và/hoặc `{"sort_order":2}` |
| DELETE | `/images/{uuid}` | 204 |
| PUT | `/company` | Cập nhật toàn bộ thông tin công ty (URL MXH phải `https://` hoặc rỗng) |
| GET | `/promotions` | Tất cả KM (phân trang, `?status=active|upcoming|ended`) |
| POST | `/promotions` | `{"title","content","start_date","end_date","is_active","items":[{"product_id","promo_price","note"}]}` → 201 |
| PUT | `/promotions/{id}` | Thay toàn bộ (items thay thế trong 1 transaction) |
| DELETE | `/promotions/{id}` | 204 |
| PUT | `/customers/{id}` | Sửa khách (SĐT chuẩn hóa, trùng → 409) |
| DELETE | `/customers/{id}` | Có đơn → 409 |
| PUT | `/orders/{id}` | Sửa đơn (cùng body như POST, không có `customer`), thay items |
| DELETE | `/orders/{id}` | 204 |
| GET | `/users` | Danh sách `{"id","username","full_name","role","is_active","last_login_at","created_at"}` |
| POST | `/users` | `{"username","full_name","password"}` → tạo **role `user`** → 201; trùng → 409 |
| PATCH | `/users/{id}` | `{"full_name"?, "is_active"?}`; khóa → tăng `token_version`. Không cho tự khóa mình / khóa admin → 400 |
| POST | `/users/{id}/reset-password` | `{"new_password"}` → 204; tăng `token_version` |

Admin cũng gọi được toàn bộ `/api/internal/*`.

---

## 6. Bảo mật
| Hạng mục | Quyết định |
|---|---|
| Mật khẩu | bcrypt cost 12; độ dài 8–72 ký tự; không log mật khẩu |
| JWT | HS256, secret `JWT_SECRET` (≥ 32 ký tự, thiếu → server không khởi động); claims `sub`(user id), `role`, `tv`(token_version), `iat`, `exp` = 12h (env `JWT_TTL_HOURS`); parser ép `WithValidMethods([]string{"HS256"})` |
| Middleware RequireAuth | Validate JWT → load user theo `sub` (1 query PK) → từ chối nếu `!is_active` hoặc `tv != token_version` ⇒ khóa/đổi mật khẩu có hiệu lực ngay. Role lấy **từ DB**, không tin claim |
| Lưu token phía client | `localStorage` (key `dh_token`) + Bearer header ⇒ miễn nhiễm CSRF. Giảm rủi ro XSS: Vue tự escape, **cấm `v-html`** với dữ liệu người dùng (mô tả SP hiển thị dạng text với `white-space: pre-line`), CSP header |
| Admin ban đầu | Lúc khởi động: nếu `SELECT count(*) FROM users WHERE role='admin'` = 0 → bắt buộc có `ADMIN_USERNAME`/`ADMIN_PASSWORD` (thiếu → log lỗi và thoát), tạo admin. Đã có admin → bỏ qua (không ghi đè mật khẩu) |
| Rate limit (theo IP, `X-Forwarded-For` đầu tiên do Render đặt — dùng chi `RealIP`) | `/api/public/warranty/lookup`: 5 req/phút, burst 5; `/api/auth/login`: 10 req/phút, burst 5; toàn bộ `/api/*`: 120 req/phút. Ngoài ra lookup có thêm giới hạn toàn cục 300 req/giờ để chống dò phân tán |
| Chống dò SĐT | Rate limit + che tên/SĐT + không trả địa chỉ + POST body (SĐT không nằm trong URL/log) + phản hồi đồng nhất khi không có đơn |
| Upload ảnh | `http.MaxBytesReader` 2 MB + 64 KB overhead; xác định loại bằng `http.DetectContentType` trên bytes thực (không tin header/đuôi file); chỉ `image/jpeg|png|webp`; tối đa 8 ảnh/SP (đếm trước khi insert); đọc kích thước bằng `image.DecodeConfig` (jpeg/png; webp qua `golang.org/x/image/webp` — NÊN) — decode lỗi → 415. Frontend NÊN resize về cạnh dài ≤ 1600px (canvas → webp/jpeg 0.85) trước khi upload |
| Body JSON | Giới hạn 1 MB; `DisallowUnknownFields`; trim chuỗi; validate độ dài theo ràng buộc DB |
| CORS | Chỉ bật khi có `ALLOWED_ORIGINS` (dev); production cùng origin |
| Header bảo mật | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'` |
| Lỗi nội bộ | Map lỗi domain → HTTP; lỗi lạ → 500 message chung + log `slog` có `request_id`. Lỗi pgx unique violation (23505) → 409, FK violation (23503) → 409 |
| SQL | 100% tham số hóa; không nối chuỗi từ input |
| Secret | Chỉ qua env; `.gitignore` chặn `.env*`; không in `DATABASE_URL` ra log |
| Server | `ReadHeaderTimeout 10s`, `ReadTimeout 30s`, `WriteTimeout 60s`, graceful shutdown 10s khi SIGTERM |

---

## 7. Thiết kế UI/UX

### 7.1 Route
| Path | View | Layout | Guard |
|---|---|---|---|
| `/` | HomeView (+ `?sp=slug` mở modal) | Public | — |
| `/bao-hanh` | WarrantyView | Public | — |
| `/dang-nhap` | LoginView (`?redirect=`) | Public tối giản | Đã đăng nhập → `/noi-bo` |
| `/noi-bo` | DashboardView | Internal | `requiresAuth` |
| `/noi-bo/san-pham` | ProductsView (user: xem; admin: thêm nút Sửa/Ẩn/Thêm) | Internal | auth |
| `/noi-bo/san-pham/moi`, `/noi-bo/san-pham/:id` | ProductEditView (thông tin + ImageUploader + SpecsEditor) | Internal | `role: admin` |
| `/noi-bo/bang-gia` | **PriceListView — trang ẩn KM/bảng giá nội bộ** | Internal | auth |
| `/noi-bo/khach-hang`, `/noi-bo/khach-hang/:id` | CustomersView / CustomerDetailView | Internal | auth |
| `/noi-bo/don-hang`, `/noi-bo/don-hang/moi`, `/noi-bo/don-hang/:id` | OrdersView / OrderFormView / OrderDetailView | Internal | auth (sửa/xóa: admin) |
| `/noi-bo/khuyen-mai`, `/noi-bo/khuyen-mai/moi`, `/noi-bo/khuyen-mai/:id` | PromotionsAdminView / PromotionEditView | Internal | admin |
| `/noi-bo/cong-ty` | CompanyEditView | Internal | admin |
| `/noi-bo/tai-khoan` | UsersView | Internal | admin |
| `/noi-bo/doi-mat-khau` | ChangePasswordView | Internal | auth |
| `/:pathMatch(.*)*` | NotFoundView | Public | — |

Guard `beforeEach`: `meta.requiresAuth` và chưa có token → `/dang-nhap?redirect=<path>`; `meta.role==='admin'` mà user thường → về `/noi-bo` + toast "Bạn không có quyền". Khi app khởi động có token → gọi `/api/auth/me` để xác thực lại; 401 bất kỳ lúc nào → xóa token, chuyển trang đăng nhập. Guard frontend chỉ là UX — **bảo vệ thật nằm ở API** (401/403). Router dùng `createWebHistory` (backend có SPA fallback).

### 7.2 Trang công khai
- **Header** (sticky): logo chữ "Đông Hải", menu: Sản phẩm (cuộn tới lưới) · Bảo hành · Liên hệ (cuộn tới footer); nút gọi điện `tel:`. Mobile: menu hamburger.
- **Hero**: nền gradient màu thương hiệu (xanh biển `#0B5ED7` → `#0A3D91`, nhấn cam `#F59E0B`), tên công ty, slogan, 2 CTA: "Xem sản phẩm", "Tra cứu bảo hành".
- **Lưới sản phẩm**: ô tìm kiếm lọc phía client; grid `repeat(auto-fill, minmax(220px,1fr))` (mobile 2 cột, ≥1024px 4 cột). ProductCard: ảnh vuông 1:1 `object-fit: cover`, `loading="lazy"`, tên (2 dòng, ellipsis), mô tả ngắn, giá đỏ đậm `5.990.000 ₫`, nút "Xem chi tiết". Chưa có ảnh → `placeholder-product.svg`. Trạng thái tải: skeleton; rỗng: "Đang cập nhật sản phẩm".
- **ProductModal**: overlay tối, đóng bằng nút ×, phím Esc, bấm nền; khóa cuộn body; trên desktop 2 cột (gallery trái, thông tin phải), mobile toàn màn hình 1 cột. Gallery: ảnh lớn + dãy thumbnail; bấm ảnh lớn → xem toàn màn hình (lightbox, vuốt/phím ←→). Thông tin: tên, giá, mô tả chi tiết (text, `pre-line`), bảng thông số, nút "Liên hệ mua: <SĐT>" (tel:) và "Chat Zalo". Mở modal cập nhật `?sp=slug` (router.replace) để chia sẻ link.
- **Footer** (SiteFooter): 3 cột — (1) tên + giới thiệu; (2) địa chỉ, SĐT, email, giờ làm việc; (3) "Kết nối với chúng tôi": 4 nút SocialButtons tròn có icon SVG inline + nhãn: Facebook (#1877F2), TikTok (#000), Zalo (#0068FF), Shopee (#EE4D2D), `target="_blank" rel="noopener noreferrer"`, ẩn nút nếu link rỗng. Dòng cuối: © năm + tên công ty · link nhỏ "Đăng nhập nội bộ".
- **Bảo hành**: card giữa trang, PhoneInput (`inputmode="tel"`), nút "Tra cứu"; kết quả: tên & SĐT đã che, mỗi đơn là 1 card: mã đơn, ngày mua, danh sách SP (+serial), "Hạn bảo hành: dd/MM/yyyy", badge xanh "Còn bảo hành · còn N ngày" / xám "Hết bảo hành". Không có đơn: "Không tìm thấy đơn hàng với số điện thoại này. Vui lòng liên hệ <SĐT công ty>". 429: "Bạn thao tác quá nhanh, vui lòng thử lại sau ít phút".
- **Đăng nhập**: form username/mật khẩu (hiện/ẩn mật khẩu), lỗi chung, nút loading.

### 7.3 Khu nội bộ
- InternalLayout: sidebar trái (mobile: drawer), topbar hiển thị tên + vai trò + nút Đăng xuất. Menu sinh từ mảng cấu hình lọc theo `role`:
  - Tất cả: Tổng quan · Sản phẩm · **Khuyến mãi & bảng giá** (`/noi-bo/bang-gia`) · Khách hàng · Đơn hàng · Đổi mật khẩu
  - Chỉ admin (nhóm "Quản trị"): Quản lý khuyến mãi · Thông tin công ty · Tài khoản
- Bảng dữ liệu: DataTable responsive (mobile chuyển dạng card), phân trang, ô tìm kiếm debounce 300ms.
- OrderFormView: bước 1 nhập SĐT → gọi `by-phone`: có → hiện thẻ khách; không → hiện ô Họ tên/Địa chỉ để tạo mới. Bước 2: ngày mua (mặc định hôm nay), dòng SP (select SP, SL, đơn giá MoneyInput tự điền giá niêm yết, serial), nút thêm/xóa dòng, tổng tiền. Hiển thị trước "Hạn bảo hành dự kiến". Lưu → chuyển tới chi tiết đơn + toast.
- PriceListView: khối CTKM (thẻ: tiêu đề, thời gian, badge "Đang diễn ra"/"Sắp diễn ra", nội dung, bảng SP–giá KM) + bảng giá: Sản phẩm | Giá niêm yết | Giá nội bộ | Giá KM hiện hành (gạch ngang giá cũ khi có KM).
- Xác nhận trước thao tác xóa/khóa (ConfirmDialog). Toast thành công/lỗi.

### 7.4 Định dạng & chung
- `formatVND(n)` = `new Intl.NumberFormat('vi-VN').format(n) + ' ₫'`; ngày hiển thị `dd/MM/yyyy`. `<html lang="vi">`, font hệ thống (`system-ui, "Segoe UI", Roboto, Arial`), không tải font ngoài (CSP).
- Responsive breakpoint: 640 / 1024 px; vùng chạm ≥ 44px; ảnh có `alt`; modal có `role="dialog" aria-modal="true"`, focus trap.
- Title/meta: "Đông Hải — <trang>"; `<meta name="robots" content="noindex">` cho route `/noi-bo*` và `/dang-nhap` (set qua router afterEach).

---

## 8. Dữ liệu mẫu (`backend/seed/demo.sql`, chạy khi `SEED_DEMO=true` và `products` rỗng)
- **company_info**: "Công ty TNHH Thương mại Đông Hải", slogan "Thiết bị gia dụng chính hãng – Bảo hành 2 năm", địa chỉ "Số 25 Lạch Tray, Ngô Quyền, Hải Phòng", SĐT `0225 3888 999`→ hiển thị nguyên văn, email `lienhe@donghai.vn`, giờ "8:00 – 17:30 (Thứ 2 – Thứ 7)", 4 link MXH mẫu (`https://facebook.com/`, `https://www.tiktok.com/`, `https://zalo.me/`, `https://shopee.vn/`) — ghi chú trong README để khách thay link thật.
- **products** (8, ảnh để trống → placeholder; admin upload sau):

| sort | Tên | slug | Giá niêm yết | Giá nội bộ |
|---|---|---|---|---|
| 1 | Máy lọc nước RO Đông Hải DH-9 | may-loc-nuoc-ro-dh-9 | 5.990.000 | 4.900.000 |
| 2 | Nồi cơm điện cao tần DH-RC18 | noi-com-dien-cao-tan-dh-rc18 | 2.490.000 | 1.950.000 |
| 3 | Bếp từ đôi DH-IH22 | bep-tu-doi-dh-ih22 | 7.690.000 | 6.300.000 |
| 4 | Quạt điều hòa DH-AC40 | quat-dieu-hoa-dh-ac40 | 3.290.000 | 2.650.000 |
| 5 | Máy hút bụi cầm tay DH-V12 | may-hut-bui-cam-tay-dh-v12 | 1.890.000 | 1.450.000 |
| 6 | Nồi chiên không dầu DH-AF6 | noi-chien-khong-dau-dh-af6 | 1.990.000 | 1.550.000 |
| 7 | Máy xay sinh tố DH-BL15 | may-xay-sinh-to-dh-bl15 | 1.290.000 | 990.000 |
| 8 | Bình nóng lạnh DH-WH30 (ẩn – ngừng KD) | binh-nong-lanh-dh-wh30 | 3.590.000 | 2.900.000 |
Mỗi SP có mô tả ngắn/chi tiết và 3–5 thông số (`specs`).
- **promotions**: (1) "Khuyến mãi cuối năm 2026" 01/11/2026–31/12/2026, giá KM cho SP 1 (5.490.000), 3 (6.990.000), 6 (1.690.000); (2) "Tết Đinh Mùi 2027" 15/01/2027–15/02/2027 (sắp diễn ra), SP 2 (2.190.000); (3) "Hè 2026" 01/06–31/07/2026 (đã kết thúc).
- **customers** (5): Nguyễn Văn An `0912345678`, Trần Thị Bình `0987654321`, Lê Văn Cường `0909111222`, Phạm Thị Dung `0356789012`, Hoàng Minh Đức `0778123456` (địa chỉ Hải Phòng/Hải Dương/Quảng Ninh).
- **orders** (8) phủ đủ trạng thái: An – 10/03/2025 (còn hạn) + 05/11/2026; Bình – 15/08/2024 (hạn 15/08/2026 → **đã hết hạn**); Cường – 29/02/2024 (kiểm thử 29/02 → hết hạn 28/02/2026); Dung – 20/10/2026; Đức – 01/01/2023 (hết hạn) và 12/12/2025; thêm 1 đơn nhiều dòng cho Bình. Một số dòng có serial (`DH9-2503-0001`...). `created_by` = NULL (seed).
- Không seed user — admin tạo từ env; tài liệu demo khuyên admin tạo user `nhanvien1`.

---

## 9. Kế hoạch test & triển khai

### 9.1 Test
| Loại | Nội dung | Lệnh |
|---|---|---|
| Unit Go | `phone.Normalize`: `"0912 345 678"`, `"+84912345678"`, `"84912345678"`, `"0084 912.345.678"`, `"(024) 3826 1234"` (11 số), lỗi: `"12345"`, `"09123456789012"`, `"abc"`, `""`; `Mask`, `MaskName` (có dấu, 1 từ) | `cd backend && go test ./...` |
| Unit Go | `warranty.EndDate`: 2025-03-10→2027-03-10; 2024-02-29→2026-02-28; 2024-12-31→2026-12-31; `Status` đúng ngày cuối = active, ngày sau = expired, `days_left` | |
| Unit Go | JWT: tạo/parse, hết hạn → lỗi, sai alg (`none`, HS512 khác secret) → lỗi; bcrypt hash/verify; slugify tiếng Việt | |
| Unit Go (httptest) | Middleware: không token → 401; user gọi `/api/admin/*` → 403; token có `tv` cũ → 401; user bị khóa → 401 | |
| Tích hợp (Postgres local 16) | `TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/donghai_test?sslmode=disable go test -tags=integration ./...`: migrate từ đầu (2 lần — kiểm idempotent), seed admin, login, CRUD SP, upload ảnh (đúng/ sai loại/ quá 2 MB), tạo đơn kèm khách mới & khách trùng SĐT, tra bảo hành (che dữ liệu, không lộ địa chỉ), **public products không chứa `price_internal`**, `/api/internal/price-list` không token → 401, rate limit lookup trả 429 ở lần thứ 6 | khởi động PG: `/usr/lib/postgresql/16/bin/initdb` + `pg_ctl` (hoặc `service postgresql start`) |
| Frontend | `cd frontend && npm ci && npm run build` không lỗi; (NÊN) `vitest` cho `formatVND`, `phone.js` | |
| Docker | `docker build -t donghai .` rồi `docker run -e DATABASE_URL=... -e JWT_SECRET=... -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=... -e SEED_DEMO=true -p 8080:8080 donghai`; kiểm tra `/`, `/bao-hanh` (refresh trực tiếp → SPA fallback OK), `/api/health` | |
| Smoke thủ công | Luồng: xem trang chủ → modal → tra bảo hành `0912345678` → đăng nhập admin → tạo user → đăng nhập user → nhập đơn → tra lại bảo hành → user truy cập `/noi-bo/tai-khoan` bị chặn | |

### 9.2 Triển khai
1. Agent 4 apply `0001_init.sql`, `0002_rls.sql` (và tùy chọn `seed/demo.sql`) lên Supabase project `lzmhasrctmekzlkngrah` qua MCP; kiểm `list_tables` + `get_advisors`.
2. Tạo Render Web Service (Docker, free, Singapore) từ repo `anhthu1205/web-app-donghai` nhánh `main` (qua Render API hoặc `render.yaml`), auto-deploy on push.
3. Biến môi trường:

| Biến | Bắt buộc | Ví dụ / Mặc định | Ghi chú |
|---|---|---|---|
| `DATABASE_URL` | Có | `postgresql://postgres.lzmhasrctmekzlkngrah:<MẬT_KHẨU>@aws-0-ap-northeast-2.pooler.supabase.com:5432/postgres?sslmode=require` | **Session pooler cổng 5432** (Render free chỉ có IPv4; host direct `db.<ref>.supabase.co` là IPv6). Mật khẩu do khách cung cấp (`sync: false`). Hostname pooler chính xác lấy ở Supabase Dashboard → Connect |
| `JWT_SECRET` | Có | `generateValue: true` trên Render | ≥ 32 ký tự |
| `ADMIN_USERNAME` | Có (lần đầu) | `admin` | |
| `ADMIN_PASSWORD` | Có (lần đầu) | `sync: false` | Khách đặt; đổi ngay sau lần đăng nhập đầu |
| `PORT` | Không | Render tự đặt / 8080 | |
| `STATIC_DIR` | Không | `/app/web` | Đặt trong Dockerfile |
| `JWT_TTL_HOURS` | Không | `12` | |
| `MIGRATE_ON_START` | Không | `true` | |
| `SEED_DEMO` | Không | `false` (đặt `true` cho bản demo đầu) | Chỉ seed khi `products` rỗng |
| `ALLOWED_ORIGINS` | Không | rỗng | Chỉ dùng khi dev khác origin |
| `DB_MAX_CONNS` | Không | `5` | Pooler free giới hạn kết nối |
| `LOG_LEVEL` | Không | `info` | |

4. Kiểm tra sau deploy: `GET https://<service>.onrender.com/api/health` = ok; smoke test như 9.1.

---

## 10. Rủi ro & giả định
| # | Rủi ro / Giả định | Ứng phó |
|---|---|---|
| R1 | Không có mật khẩu DB trong phiên → không thể chạy thật cho tới khi khách cung cấp `DATABASE_URL` | Schema apply trước bằng MCP; service Render tạo sẵn, khách điền env rồi Manual Deploy. Ghi rõ trong README |
| R2 | Render free **ngủ sau ~15 phút** không có truy cập, lần đầu thức dậy 30–60s; 750 giờ/tháng | Chấp nhận cho giai đoạn đầu; nâng gói Starter nếu cần; (tùy chọn) ping định kỳ `/api/health` |
| R3 | Supabase free tạm dừng project sau 7 ngày không hoạt động; DB 500 MB | Ảnh ≤ 2 MB, khuyến nghị resize phía client (~200 KB); 8 ảnh × ~50 SP ≈ 80 MB, đủ |
| R4 | Ảnh trong bytea làm tăng tải DB | Cache `immutable` 1 năm ở trình duyệt/CDN theo uuid; truy vấn danh sách **không SELECT cột `data`** |
| R5 | pgx + pooler: transaction pooler (6543) không hỗ trợ prepared statement | Dùng session pooler 5432; nếu buộc dùng 6543 thì đặt `default_query_exec_mode=simple_protocol` trong `DATABASE_URL` |
| R6 | Rate limit in-memory mất khi restart, chỉ đúng với 1 instance | Phù hợp gói free 1 instance |
| R7 | Khách viết "3 nút" nhưng liệt kê 4 MXH | Làm 4 nút, ẩn nút nếu link rỗng |
| R8 | "User chỉ xem + nhập" — giả định user được **thêm** khách/đơn nhưng không sửa/xóa; sửa/xóa do admin | Agent 2 xác nhận; dễ nới quyền sau (đổi nhóm route) |
| R9 | Bảo hành tính theo đơn (cùng ngày mua cho mọi dòng), 2 năm cố định cho mọi SP | Hằng số `WarrantyYears`; mở rộng theo SP sau nếu cần (thêm cột `warranty_months`) |
| R10 | Chỉ 1 admin (tạo từ env); API tạo tài khoản chỉ tạo role `user` | Đúng yêu cầu; nếu cần nhiều admin, bổ sung sau |
| R11 | JWT lưu localStorage có rủi ro XSS | CSP chặt, cấm `v-html`, không dùng script bên thứ ba |
| R12 | Tên miền chưa có — dùng `*.onrender.com` | Gắn domain riêng sau trên Render |
