-- Dữ liệu mẫu (chỉ chạy khi SEED_DEMO=true và bảng products rỗng).
-- Có thể apply tay trên Supabase; tham chiếu theo slug/SĐT nên không phụ thuộc id.

UPDATE company_info SET
  name          = 'Công ty TNHH Thương mại Đông Hải',
  slogan        = 'Thiết bị gia dụng chính hãng – Bảo hành 2 năm',
  about         = 'Đông Hải chuyên phân phối thiết bị gia dụng chính hãng: máy lọc nước, bếp từ, nồi cơm điện, quạt điều hòa… Hơn 10 năm phục vụ khách hàng khu vực Hải Phòng và các tỉnh lân cận, cam kết hàng chính hãng, giá minh bạch và bảo hành tận nơi 2 năm.',
  address       = 'Số 25 Lạch Tray, Ngô Quyền, Hải Phòng',
  phone         = '0225 3888 999',
  email         = 'lienhe@donghai.vn',
  working_hours = '8:00 – 17:30 (Thứ 2 – Thứ 7)',
  facebook_url  = 'https://facebook.com/',
  tiktok_url    = 'https://www.tiktok.com/',
  zalo_url      = 'https://zalo.me/',
  shopee_url    = 'https://shopee.vn/'
WHERE id = 1;

INSERT INTO products (name, slug, sku, short_description, description, specs, unit, price_public, price_internal, is_visible, sort_order) VALUES
('Máy lọc nước RO Đông Hải DH-9', 'may-loc-nuoc-ro-dh-9', 'DH-9',
 'Máy lọc nước RO 9 lõi, nước tinh khiết uống trực tiếp, tủ đứng sang trọng.',
 E'Máy lọc nước RO DH-9 sử dụng màng RO nhập khẩu cùng hệ 9 lõi lọc, loại bỏ đến 99,9% vi khuẩn, kim loại nặng và tạp chất.\nNước sau lọc bổ sung khoáng chất có lợi, có thể uống trực tiếp.\nThiết kế tủ đứng cường lực, phù hợp mọi không gian bếp.',
 '[{"label":"Công suất lọc","value":"10 lít/giờ"},{"label":"Số lõi lọc","value":"9 lõi"},{"label":"Dung tích bình chứa","value":"10 lít"},{"label":"Công suất điện","value":"36 W"},{"label":"Xuất xứ","value":"Việt Nam"}]',
 'cái', 5990000, 4900000, true, 1),
('Nồi cơm điện cao tần DH-RC18', 'noi-com-dien-cao-tan-dh-rc18', 'DH-RC18',
 'Công nghệ nấu cao tần IH, lòng nồi niêu 6 lớp, cơm chín đều, dẻo thơm.',
 E'Nồi cơm điện cao tần DH-RC18 làm nóng toàn bộ lòng nồi bằng công nghệ IH giúp hạt cơm chín đều từ trong ra ngoài.\n12 chế độ nấu: cơm trắng, cơm gạo lứt, cháo, hấp, nấu chậm…\nLòng nồi phủ chống dính 6 lớp, dễ vệ sinh.',
 '[{"label":"Dung tích","value":"1,8 lít"},{"label":"Công suất","value":"1.200 W"},{"label":"Chế độ nấu","value":"12 chế độ"},{"label":"Lòng nồi","value":"Niêu 6 lớp chống dính"}]',
 'cái', 2490000, 1950000, true, 2),
('Bếp từ đôi DH-IH22', 'bep-tu-doi-dh-ih22', 'DH-IH22',
 'Bếp từ đôi mặt kính chịu nhiệt, 9 mức công suất, khóa trẻ em an toàn.',
 E'Bếp từ đôi DH-IH22 với mặt kính chịu lực, chịu nhiệt cao, dễ lau chùi.\nĐiều khiển cảm ứng trượt, hẹn giờ tắt, tự nhận diện nồi.\nCông nghệ inverter tiết kiệm điện năng.',
 '[{"label":"Tổng công suất","value":"4.000 W"},{"label":"Số vùng nấu","value":"2"},{"label":"Mặt kính","value":"Kính chịu nhiệt nhập khẩu"},{"label":"Kích thước","value":"730 x 430 mm"}]',
 'cái', 7690000, 6300000, true, 3),
('Quạt điều hòa DH-AC40', 'quat-dieu-hoa-dh-ac40', 'DH-AC40',
 'Làm mát bằng hơi nước, bình chứa 40 lít, phù hợp phòng 30–40 m².',
 E'Quạt điều hòa DH-AC40 làm mát nhanh nhờ tấm làm mát Honeycomb và bình nước lớn 40 lít.\nĐiều khiển từ xa, 3 tốc độ gió, đảo gió tự động.\nTiêu thụ điện thấp, vận hành êm.',
 '[{"label":"Dung tích bình nước","value":"40 lít"},{"label":"Công suất","value":"150 W"},{"label":"Diện tích làm mát","value":"30 – 40 m²"},{"label":"Điều khiển","value":"Remote + cảm ứng"}]',
 'cái', 3290000, 2650000, true, 4),
('Máy hút bụi cầm tay DH-V12', 'may-hut-bui-cam-tay-dh-v12', 'DH-V12',
 'Không dây, lực hút mạnh 12.000 Pa, pin dùng 40 phút, kèm 3 đầu hút.',
 E'Máy hút bụi cầm tay DH-V12 nhỏ gọn, không dây, thuận tiện vệ sinh sofa, ô tô, góc khuất.\nBộ lọc HEPA giữ lại bụi mịn, có thể rửa bằng nước.',
 '[{"label":"Lực hút","value":"12.000 Pa"},{"label":"Thời gian dùng pin","value":"40 phút"},{"label":"Trọng lượng","value":"1,3 kg"}]',
 'cái', 1890000, 1450000, true, 5),
('Nồi chiên không dầu DH-AF6', 'noi-chien-khong-dau-dh-af6', 'DH-AF6',
 'Dung tích 6 lít, chiên nướng giảm 80% dầu mỡ, màn hình cảm ứng.',
 E'Nồi chiên không dầu DH-AF6 dùng công nghệ khí nóng tuần hoàn, món ăn giòn ngon mà giảm đến 80% dầu mỡ.\n8 chương trình nấu cài sẵn, hẹn giờ đến 60 phút.',
 '[{"label":"Dung tích","value":"6 lít"},{"label":"Công suất","value":"1.700 W"},{"label":"Nhiệt độ","value":"80 – 200 °C"},{"label":"Điều khiển","value":"Cảm ứng"}]',
 'cái', 1990000, 1550000, true, 6),
('Máy xay sinh tố DH-BL15', 'may-xay-sinh-to-dh-bl15', 'DH-BL15',
 'Cối thủy tinh 1,5 lít, 6 lưỡi dao thép không gỉ, xay đá mịn.',
 E'Máy xay sinh tố DH-BL15 với động cơ mạnh mẽ, cối thủy tinh dày chịu nhiệt.\nPhù hợp xay sinh tố, nước ép, xay đá, xay thịt (cối phụ).',
 '[{"label":"Dung tích cối","value":"1,5 lít"},{"label":"Công suất","value":"600 W"},{"label":"Lưỡi dao","value":"6 lưỡi inox 304"}]',
 'cái', 1290000, 990000, true, 7),
('Bình nóng lạnh DH-WH30', 'binh-nong-lanh-dh-wh30', 'DH-WH30',
 'Bình nóng lạnh gián tiếp 30 lít (ngừng kinh doanh).',
 E'Bình nóng lạnh DH-WH30 – sản phẩm đã ngừng kinh doanh, vẫn tiếp tục bảo hành cho khách đã mua.',
 '[{"label":"Dung tích","value":"30 lít"},{"label":"Công suất","value":"2.500 W"},{"label":"Chống giật","value":"ELCB"}]',
 'cái', 3590000, 2900000, false, 8);

INSERT INTO promotions (title, content, start_date, end_date, is_active) VALUES
('Khuyến mãi cuối năm 2026', E'Giảm giá sâu các sản phẩm chủ lực dịp cuối năm.\nÁp dụng tại cửa hàng và đơn đặt qua hotline.', '2026-11-01', '2026-12-31', true),
('Tết Đinh Mùi 2027', E'Chương trình đón Tết Đinh Mùi: ưu đãi nồi cơm điện cao tần.', '2027-01-15', '2027-02-15', true),
('Hè 2026', E'Ưu đãi mùa hè cho quạt điều hòa.', '2026-06-01', '2026-07-31', true);

INSERT INTO promotion_items (promotion_id, product_id, promo_price, note)
SELECT pr.id, p.id, v.price, v.note
FROM (VALUES
  ('Khuyến mãi cuối năm 2026', 'may-loc-nuoc-ro-dh-9', 5490000, 'Tặng kèm bộ lõi số 1'),
  ('Khuyến mãi cuối năm 2026', 'bep-tu-doi-dh-ih22', 6990000, ''),
  ('Khuyến mãi cuối năm 2026', 'noi-chien-khong-dau-dh-af6', 1690000, ''),
  ('Tết Đinh Mùi 2027', 'noi-com-dien-cao-tan-dh-rc18', 2190000, 'Tặng muỗng xới cơm'),
  ('Hè 2026', 'quat-dieu-hoa-dh-ac40', 2890000, '')
) AS v(title, slug, price, note)
JOIN promotions pr ON pr.title = v.title
JOIN products p ON p.slug = v.slug;

INSERT INTO customers (full_name, phone, address, note) VALUES
('Nguyễn Văn An',  '0912345678', 'Số 12 Lê Lợi, Ngô Quyền, Hải Phòng', 'Khách quen'),
('Trần Thị Bình',  '0987654321', 'Số 8 Trần Hưng Đạo, TP. Hải Dương', ''),
('Lê Văn Cường',   '0909111222', 'Tổ 5, Hồng Hải, Hạ Long, Quảng Ninh', ''),
('Phạm Thị Dung',  '0356789012', 'Số 40 Văn Cao, Lê Chân, Hải Phòng', ''),
('Hoàng Minh Đức', '0778123456', 'Phường Cẩm Thượng, TP. Hải Dương', 'Giao hàng giờ hành chính')
ON CONFLICT (phone) DO NOTHING;

-- 8 đơn; thứ tự id = thứ tự dòng (seq = cột thứ nhất)
CREATE TEMP TABLE _seed_orders (seq int, phone text, purchase_date date, note text) ON COMMIT DROP;
INSERT INTO _seed_orders VALUES
(1, '0912345678', '2025-03-10', 'Lắp đặt tại nhà'),
(2, '0987654321', '2024-08-15', ''),
(3, '0909111222', '2024-02-29', 'Đơn ngày 29/02 (kiểm thử năm nhuận)'),
(4, '0778123456', '2023-01-01', ''),
(5, '0987654321', '2025-06-20', 'Đơn nhiều sản phẩm'),
(6, '0778123456', '2025-12-12', ''),
(7, '0356789012', '2026-10-20', ''),
(8, '0912345678', '2026-11-05', 'Mua trong chương trình KM cuối năm');

CREATE TEMP TABLE _seed_order_ids (seq int, order_id bigint) ON COMMIT DROP;
DO $$ DECLARE r record; new_id bigint; BEGIN
  FOR r IN SELECT * FROM _seed_orders ORDER BY seq LOOP
    INSERT INTO orders (customer_id, purchase_date, note)
    SELECT c.id, r.purchase_date, r.note FROM customers c WHERE c.phone = r.phone
    RETURNING id INTO new_id;
    INSERT INTO _seed_order_ids VALUES (r.seq, new_id);
  END LOOP;
END $$;

INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, serial_number)
SELECT o.order_id, p.id, p.name, v.qty, v.price, v.serial
FROM (VALUES
  (1, 'may-loc-nuoc-ro-dh-9', 1, 5990000, 'DH9-2503-0001'),
  (2, 'noi-com-dien-cao-tan-dh-rc18', 1, 2490000, 'RC18-2408-0002'),
  (3, 'quat-dieu-hoa-dh-ac40', 1, 3290000, 'AC40-2402-0003'),
  (4, 'binh-nong-lanh-dh-wh30', 1, 3590000, 'WH30-2301-0004'),
  (5, 'bep-tu-doi-dh-ih22', 1, 7690000, 'IH22-2506-0005'),
  (5, 'may-xay-sinh-to-dh-bl15', 2, 1290000, NULL),
  (6, 'may-loc-nuoc-ro-dh-9', 1, 5990000, 'DH9-2512-0006'),
  (7, 'may-hut-bui-cam-tay-dh-v12', 1, 1890000, 'V12-2610-0007'),
  (8, 'noi-chien-khong-dau-dh-af6', 1, 1690000, NULL)
) AS v(seq, slug, qty, price, serial)
JOIN _seed_order_ids o ON o.seq = v.seq
JOIN products p ON p.slug = v.slug;
