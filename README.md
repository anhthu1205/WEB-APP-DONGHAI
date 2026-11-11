# WEB-APP-DONGHAI

Web app cho Công ty Đông Hải: trang giới thiệu & sản phẩm kiểu TMĐT, tra cứu bảo hành theo SĐT, khu vực nội bộ (admin/user) quản lý sản phẩm, khách hàng, đơn hàng, khuyến mãi.

- Backend: Go (`backend/`) — REST JSON `/api/...`
- Frontend: Vue 3 + Vite (`frontend/`)
- Database: Supabase (Postgres)
- Triển khai: 1 Docker image trên Render

Tài liệu: xem thư mục [`docs/`](docs/) — bắt đầu từ [`docs/01-PHUONG-AN.md`](docs/01-PHUONG-AN.md).

Mọi bí mật (`DATABASE_URL`, `JWT_SECRET`, `ADMIN_PASSWORD`) cấu hình qua biến môi trường, không commit vào repo.
