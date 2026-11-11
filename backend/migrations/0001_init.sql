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
