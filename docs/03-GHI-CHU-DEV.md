# 03 — GHI CHÚ DEV: WEB-APP-DONGHAI

> Người viết: Agent 3 (DEV). Code theo `01-PHUONG-AN.md` + `02-DUYET-VA-BAN-GIAO.md`. Tài liệu này ghi các điểm lệch/bổ sung và lý do, cùng kết quả tự kiểm tra.

## 1. Điểm lệch / bổ sung so với phương án

| # | Hạng mục | Phương án | Hiện thực | Lý do |
|---|---|---|---|---|
| L1 | Font chữ | 7.4: font hệ thống, không tải font ngoài | **Be Vietnam Pro** tự host qua `@fontsource/be-vietnam-pro` (đóng gói vào `dist`, phục vụ cùng origin) | Font thiết kế cho dấu tiếng Việt; trên Linux/Android font hệ thống hiển thị dấu kém. Vẫn đúng CSP (`default-src 'self'`), không gọi CDN. Trình duyệt chỉ tải subset cần (woff2). |
| L2 | Tên module Go | không quy định | `module donghai` | Ngắn gọn, import nội bộ `donghai/internal/...` |
| L3 | Slugify | "slugify tiếng Việt" (M2), không chỉ rõ package | `internal/slug` | Tách riêng để test độc lập |
| L4 | Test tích hợp | `-tags=integration` + `TEST_DATABASE_URL` | Giữ nguyên; thêm: mỗi test tạo **schema riêng** (`search_path`) rồi xóa | Các package test chạy song song không giẫm dữ liệu của nhau; không cần `-p 1` |
| L5 | Cổng Postgres local | 02 §3 dùng 5432 | `54329` (Makefile `PG_PORT` đổi được) | Tránh đụng Postgres khác trên máy |
| L6 | Component FE | `Pagination.vue`, `DataTable.vue` | `PaginationBar.vue`, bảng dùng class CSS `.table.responsive` (mobile tự thành thẻ); thêm `AppIcon`, `StateBlock`, `ModalShell`, `ProductPlaceholder`, `CustomerFormModal` | Tên component nhiều từ (quy ước Vue); bảng thuần CSS gọn hơn component |
| L7 | Route sửa đơn | 7.1 không có | `/noi-bo/don-hang/:id/sua` (admin) dùng lại `OrderFormView` | Cần giao diện cho `PUT /api/admin/orders/{id}` |
| L8 | `price-list.products` | không nói rõ SP ẩn | Chỉ SP **đang hiển thị** | SP ẩn = ngừng kinh doanh, không cần trong bảng giá bán |
| L9 | Đổi MK sai mật khẩu hiện tại | không quy định | 400 `VALIDATION_ERROR` `fields.current_password` | Nếu trả 401, frontend sẽ coi là hết phiên và đăng xuất |
| L10 | Tạo đơn với `customer.phone` đã tồn tại | 5.4 | Cho phép bỏ trống `customer.full_name` khi SĐT đã có khách | Nhân viên chỉ cần SĐT để dùng lại khách |
| L11 | `items_count` của đơn | không định nghĩa | Tổng **số lượng** sản phẩm (sum quantity) | Hiển thị "3 sản phẩm" tự nhiên hơn số dòng |
| L12 | `PUT /api/admin/products/{id}` không gửi `slug` | "cập nhật toàn bộ trường" | Giữ slug cũ (không sinh lại) | Tránh làm hỏng link `?sp=` đã chia sẻ |
| L13 | Ô tra bảo hành nhanh ở hero trang chủ | không có | Có; SĐT chuyển sang `/bao-hanh` qua `history.state` (**không** đưa lên URL) | Giữ nguyên tắc "SĐT không nằm trong URL/log" |
| L14 | Dockerfile | 01 §2.3 | Đúng 3 stage; bỏ dòng `# syntax=` | Tránh kéo thêm image frontend của BuildKit từ Docker Hub (bị rate-limit) |
| L15 | Log request | slog JSON có `request_id` + IP | Chỉ log `path` (không query string) | Tránh lộ dữ liệu trong query (vd `by-phone?phone=`) |

## 2. Ghi chú kỹ thuật
- `clientIP`: `CF-Connecting-IP` (phải là IP hợp lệ) → phần tử **cuối** của `X-Forwarded-For` → `RemoteAddr`. Agent 4 xác nhận trên Render bằng trường `ip` trong log request (D3).
- Rate limit: lookup 5/phút burst 5 + toàn cục 300/giờ; login 10/phút burst 5; `/api/*` 120/phút; `/api/images/*` ngoài rate limit.
- Ảnh: `DetectContentType` trên bytes thật; jpeg/png phải `DecodeConfig` được (không → 415); webp nhận theo chữ ký, `width/height = NULL`. Upload và đặt/xóa ảnh chính khóa dòng `products ... FOR UPDATE` để 2 request song song không vượt 8 ảnh / không có 2 ảnh chính.
- Lỗi Postgres 23505/23503 → 409 với message tiếng Việt theo tên constraint; 23514/22xxx → 400; không trả text SQL.
- Frontend: không `v-html`; ảnh xem trước khi upload dùng `data:` URL (CSP chặn `blob:`); 401 bất kỳ → xóa token, về `/dang-nhap?redirect=`.

## 3. Kết quả tự kiểm tra (2026-11-11, máy DEV)
| Lệnh | Kết quả |
|---|---|
| `gofmt -l .` | rỗng |
| `GOTOOLCHAIN=local go vet ./...` và `go vet -tags=integration ./...` | sạch |
| `go test ./...` | PASS (auth, config, middleware, phone, slug, warranty) |
| `TEST_DATABASE_URL=... go test -tags=integration ./...` (PG 16 local) | PASS — `TestMigrateTwiceAndSeed`, `TestAPI` (14 nhóm ca T-PUB/T-W/T-A/T-I/T-AD/T-IMG/T-S), `TestRateLimits`, `TestBootstrapAdmin` |
| `npm ci && npm run build` | OK; `grep -l "price_internal\|current_promo_price" dist/assets/index-*.js` rỗng (T-S6); `grep -r v-html src` rỗng |
| Chạy binary + PG local + `dist` | curl: health 200; public products 7 SP không có `price_internal/promo/is_visible`; price-list không token 401; user → 5 API admin 403; tra `+84 912 345 678` ra 2 đơn An (`091****678`, `N*** V*** An`); SĐT `abc` 400; upload 3 MB 413; txt giả jpg 415; user upload 403; `/api/khong-ton-tai` 404 JSON; `/bao-hanh`, `/noi-bo/bang-gia` trả HTML; header bảo mật đủ |
| Server thiếu env / JWT_SECRET ngắn | thoát mã 2, log rõ lỗi (T-S4) |
| `docker build` | OK (image ~24 MB). Trong sandbox phải build bằng bản Dockerfile tạm có CA proxy + mirror `mirror.gcr.io` (Docker Hub 429) — **không** commit bản này; `Dockerfile` trong repo là bản chuẩn. `docker run --network host` với PG local: health 200, `/`, `/bao-hanh` HTML, public 7 SP, price-list 401 |
| Playwright | Ảnh trong `docs/screenshots/` (trang chủ, modal, bảo hành, mobile 375px không tràn ngang, bảng giá nội bộ, nhập đơn); không có lỗi JS |

## 4. Việc còn lại (Agent 4)
- Apply migration lên Supabase, kiểm RLS/owner (T-S7, D1), deploy Render (D2), xác nhận IP client thật trong log (D3).
- Khách cung cấp `DATABASE_URL`, `ADMIN_PASSWORD`, link MXH thật, ảnh sản phẩm.
- (NÊN) Vitest cho `formatVND`/`phone.js`, Lighthouse a11y — chưa làm.
