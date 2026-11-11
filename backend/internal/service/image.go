package service

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg" // đăng ký decoder jpeg cho DecodeConfig
	_ "image/png"  // đăng ký decoder png cho DecodeConfig
	"net/http"
	"regexp"

	"donghai/internal/domain"
)

// Giới hạn ảnh sản phẩm.
const (
	MaxImageBytes       = 2 * 1024 * 1024
	MaxImagesPerProduct = 8
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ValidUUID kiểm tra định dạng uuid (tránh lỗi 22P02 → 500).
func ValidUUID(s string) bool { return uuidRe.MatchString(s) }

// SetMainRequest là body PATCH /api/admin/images/{uuid}.
type SetMainRequest struct {
	IsMain *bool `json:"is_main"`
}

// DetectImage xác định loại ảnh từ bytes thật và đọc kích thước (jpeg/png).
func DetectImage(data []byte) (mime string, w, h *int, err error) {
	if len(data) == 0 {
		return "", nil, nil, &domain.Error{Kind: domain.ErrUnsupportedMedia, Message: "Tệp ảnh rỗng"}
	}
	if len(data) > MaxImageBytes {
		return "", nil, nil, &domain.Error{Kind: domain.ErrPayloadTooLarge, Message: "Ảnh vượt quá 2 MB"}
	}
	mime = http.DetectContentType(data)
	bad := &domain.Error{Kind: domain.ErrUnsupportedMedia, Message: "Chỉ chấp nhận ảnh JPEG, PNG hoặc WebP"}
	switch mime {
	case "image/jpeg", "image/png":
		cfg, _, derr := image.DecodeConfig(bytes.NewReader(data))
		if derr != nil || cfg.Width <= 0 || cfg.Height <= 0 {
			return "", nil, nil, bad
		}
		return mime, &cfg.Width, &cfg.Height, nil
	case "image/webp":
		return mime, nil, nil, nil
	default:
		return "", nil, nil, bad
	}
}

// UploadImage thêm ảnh cho SP; ảnh đầu tiên tự là ảnh chính; tối đa 8 ảnh.
func (s *Service) UploadImage(ctx context.Context, productID int64, data []byte) (*domain.ProductImage, error) {
	mime, w, h, err := DetectImage(data)
	if err != nil {
		return nil, err
	}
	var out *domain.ProductImage
	err = s.store.InTx(ctx, func(tx *serviceTx) error {
		count, maxSort, err := tx.LockProductForImages(ctx, productID)
		if err != nil {
			return notFound(err, "Không tìm thấy sản phẩm")
		}
		if count >= MaxImagesPerProduct {
			return domain.BadRequest("Mỗi sản phẩm tối đa 8 ảnh")
		}
		out, err = tx.InsertImage(ctx, &domain.ProductImage{
			ProductID: productID, Data: data, MimeType: mime, SizeBytes: len(data),
			Width: w, Height: h, IsMain: count == 0, SortOrder: maxSort + 1,
		})
		return err
	})
	return out, err
}

// GetImage trả mime + bytes ảnh.
func (s *Service) GetImage(ctx context.Context, id string) (string, []byte, error) {
	if !ValidUUID(id) {
		return "", nil, domain.NotFound("Không tìm thấy ảnh")
	}
	mime, data, err := s.store.GetImageData(ctx, id)
	return mime, data, notFound(err, "Không tìm thấy ảnh")
}

// SetMainImage đặt ảnh làm ảnh chính (bỏ chính ảnh cũ) trong 1 transaction.
func (s *Service) SetMainImage(ctx context.Context, id string, in SetMainRequest) (*domain.ProductImage, error) {
	if !ValidUUID(id) {
		return nil, domain.NotFound("Không tìm thấy ảnh")
	}
	if in.IsMain == nil || !*in.IsMain {
		return nil, &domain.ValidationError{Fields: map[string]string{"is_main": "Chỉ hỗ trợ is_main = true"}}
	}
	var out *domain.ProductImage
	err := s.store.InTx(ctx, func(tx *serviceTx) error {
		im, err := tx.GetImageMeta(ctx, id)
		if err != nil {
			return notFound(err, "Không tìm thấy ảnh")
		}
		if _, _, err := tx.LockProductForImages(ctx, im.ProductID); err != nil {
			return err
		}
		if err := tx.SetMainImage(ctx, im.ProductID, id); err != nil {
			return err
		}
		out, err = tx.GetImageMeta(ctx, id)
		return err
	})
	return out, err
}

// DeleteImage xóa ảnh; nếu là ảnh chính → ảnh sort_order nhỏ nhất thành chính.
func (s *Service) DeleteImage(ctx context.Context, id string) error {
	if !ValidUUID(id) {
		return domain.NotFound("Không tìm thấy ảnh")
	}
	return s.store.InTx(ctx, func(tx *serviceTx) error {
		im, err := tx.GetImageMeta(ctx, id)
		if err != nil {
			return notFound(err, "Không tìm thấy ảnh")
		}
		if _, _, err := tx.LockProductForImages(ctx, im.ProductID); err != nil {
			return err
		}
		if err := tx.DeleteImage(ctx, id); err != nil {
			return notFound(err, "Không tìm thấy ảnh")
		}
		return tx.PromoteFirstImage(ctx, im.ProductID)
	})
}
