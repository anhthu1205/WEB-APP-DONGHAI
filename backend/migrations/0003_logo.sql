-- Logo công ty (1 dòng). Lưu bytea như ảnh sản phẩm; version (hash nội dung) dùng làm ETag và tham số ?v= của URL.
CREATE TABLE IF NOT EXISTS site_logo (
  id          smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  data        bytea NOT NULL,
  mime_type   text NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/webp')),
  size_bytes  integer NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 2097152),
  version     text NOT NULL CHECK (version ~ '^[0-9a-f]{16}$'),
  updated_by  bigint REFERENCES users(id) ON DELETE SET NULL,
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Giống 0002: bật RLS, không policy, thu hồi quyền anon/authenticated (chỉ backend truy cập).
ALTER TABLE site_logo ENABLE ROW LEVEL SECURITY;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
    REVOKE ALL ON TABLE site_logo FROM anon, authenticated;
  END IF;
END $$;

INSERT INTO schema_migrations (version) VALUES ('0003_logo') ON CONFLICT DO NOTHING;
