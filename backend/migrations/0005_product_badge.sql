-- Nhãn góc thẻ sản phẩm (vd. "Mua 5 tặng 1", "Mới", "-10%"). Rỗng = không hiện nhãn.
ALTER TABLE products ADD COLUMN IF NOT EXISTS badge text NOT NULL DEFAULT '' CHECK (char_length(badge) <= 40);

INSERT INTO schema_migrations (version) VALUES ('0005_product_badge') ON CONFLICT DO NOTHING;
