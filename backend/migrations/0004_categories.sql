-- Danh mục sản phẩm (ngách): mỗi danh mục có trang riêng /danh-muc/{slug}.
-- SP không thuộc danh mục nào = sản phẩm chính, hiện ở trang chủ. SP thuộc danh mục chỉ hiện ở trang
-- danh mục đó, trừ khi danh mục bật show_on_home.
CREATE TABLE IF NOT EXISTS categories (
  id            bigserial PRIMARY KEY,
  name          text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
  slug          text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) <= 120),
  description   text NOT NULL DEFAULT '' CHECK (char_length(description) <= 2000),
  is_visible    boolean NOT NULL DEFAULT true,   -- false: trang danh mục và SP trong đó ẩn với khách
  show_in_menu  boolean NOT NULL DEFAULT true,   -- hiện mục trên menu đầu trang
  show_on_home  boolean NOT NULL DEFAULT false,  -- SP của danh mục cũng hiện ở trang chủ
  sort_order    integer NOT NULL DEFAULT 0,
  -- Ảnh bìa (tùy chọn), lưu bytea như ảnh SP; version = hash nội dung, dùng cho ETag/?v=
  cover_data    bytea,
  cover_mime    text CHECK (cover_mime IN ('image/jpeg','image/png','image/webp')),
  cover_version text CHECK (cover_version ~ '^[0-9a-f]{16}$'),
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  CHECK ((cover_data IS NULL) = (cover_mime IS NULL) AND (cover_data IS NULL) = (cover_version IS NULL)),
  CHECK (cover_data IS NULL OR octet_length(cover_data) <= 2097152)
);
DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;
CREATE TRIGGER trg_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Xóa danh mục còn SP → lỗi FK (RESTRICT) → 409.
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id bigint REFERENCES categories(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);

INSERT INTO categories (name, slug, description, sort_order) VALUES
  ('Đèn dự án', 'den-du-an', 'Đèn chiếu sáng cho công trình, nhà xưởng, đường phố và các dự án lớn — chính hãng, báo giá theo số lượng.', 1)
ON CONFLICT (slug) DO NOTHING;

ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
    REVOKE ALL ON TABLE categories FROM anon, authenticated;
    REVOKE ALL ON SEQUENCE categories_id_seq FROM anon, authenticated;
  END IF;
END $$;

INSERT INTO schema_migrations (version) VALUES ('0004_categories') ON CONFLICT DO NOTHING;
