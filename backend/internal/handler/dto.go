package handler

import (
	"time"

	"donghai/internal/domain"
	"donghai/internal/service"
	"donghai/internal/warranty"
)

const dateFmt = "2006-01-02"

func imageURL(id string) string { return "/api/images/" + id }

func imageURLPtr(id *string) *string {
	if id == nil {
		return nil
	}
	u := imageURL(*id)
	return &u
}

// ---------- Công khai (KHÔNG có giá nội bộ / KM / is_visible) ----------

// SocialsDTO là 4 link MXH.
type SocialsDTO struct {
	Facebook string `json:"facebook"`
	Tiktok   string `json:"tiktok"`
	Zalo     string `json:"zalo"`
	Shopee   string `json:"shopee"`
}

// PublicCompanyDTO là thông tin công ty công khai.
type PublicCompanyDTO struct {
	Name         string     `json:"name"`
	Slogan       string     `json:"slogan"`
	About        string     `json:"about"`
	Address      string     `json:"address"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	WorkingHours string     `json:"working_hours"`
	Socials      SocialsDTO `json:"socials"`
	LogoURL      *string    `json:"logo_url"` // null = dùng logo mặc định
}

func toPublicCompany(c *domain.Company) PublicCompanyDTO {
	return PublicCompanyDTO{
		Name: c.Name, Slogan: c.Slogan, About: c.About, Address: c.Address, Phone: c.Phone,
		Email: c.Email, WorkingHours: c.WorkingHours,
		Socials: SocialsDTO{Facebook: c.FacebookURL, Tiktok: c.TiktokURL, Zalo: c.ZaloURL, Shopee: c.ShopeeURL},
		LogoURL: logoURL(c.LogoVersion),
	}
}

func logoURL(ver *string) *string {
	if ver == nil {
		return nil
	}
	u := "/api/public/logo?v=" + *ver
	return &u
}

// PublicProductDTO là SP trong danh sách công khai.
type PublicProductDTO struct {
	ID               int64   `json:"id"`
	Slug             string  `json:"slug"`
	Name             string  `json:"name"`
	ShortDescription string  `json:"short_description"`
	Price            int64   `json:"price"`
	Unit             string  `json:"unit"`
	MainImageURL     *string `json:"main_image_url"`
}

// PublicImageDTO là ảnh công khai.
type PublicImageDTO struct {
	URL    string `json:"url"`
	IsMain bool   `json:"is_main"`
	Width  *int   `json:"width"`
	Height *int   `json:"height"`
}

// PublicProductDetailDTO là chi tiết SP công khai.
type PublicProductDetailDTO struct {
	ID               int64            `json:"id"`
	Slug             string           `json:"slug"`
	Name             string           `json:"name"`
	ShortDescription string           `json:"short_description"`
	Description      string           `json:"description"`
	Specs            []domain.Spec    `json:"specs"`
	Price            int64            `json:"price"`
	Unit             string           `json:"unit"`
	Images           []PublicImageDTO `json:"images"`
}

func toPublicProduct(p *domain.Product) PublicProductDTO {
	return PublicProductDTO{ID: p.ID, Slug: p.Slug, Name: p.Name, ShortDescription: p.ShortDescription,
		Price: p.PricePublic, Unit: p.Unit, MainImageURL: imageURLPtr(p.MainImageID)}
}

func toPublicProductDetail(p *domain.Product) PublicProductDetailDTO {
	out := PublicProductDetailDTO{ID: p.ID, Slug: p.Slug, Name: p.Name, ShortDescription: p.ShortDescription,
		Description: p.Description, Specs: p.Specs, Price: p.PricePublic, Unit: p.Unit, Images: []PublicImageDTO{}}
	if out.Specs == nil {
		out.Specs = []domain.Spec{}
	}
	for _, im := range p.Images {
		out.Images = append(out.Images, PublicImageDTO{URL: imageURL(im.ID), IsMain: im.IsMain, Width: im.Width, Height: im.Height})
	}
	return out
}

// WarrantyItemDTO: chỉ tên SP, số lượng, serial (không giá).
type WarrantyItemDTO struct {
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	SerialNumber *string `json:"serial_number"`
}

// WarrantyOrderDTO là 1 đơn trong kết quả tra bảo hành.
type WarrantyOrderDTO struct {
	OrderCode       string            `json:"order_code"`
	PurchaseDate    string            `json:"purchase_date"`
	WarrantyEndDate string            `json:"warranty_end_date"`
	Status          string            `json:"status"`
	DaysLeft        int               `json:"days_left"`
	Items           []WarrantyItemDTO `json:"items"`
}

// WarrantyLookupDTO là kết quả tra bảo hành (đã che).
type WarrantyLookupDTO struct {
	PhoneMasked        string             `json:"phone_masked"`
	CustomerNameMasked string             `json:"customer_name_masked"`
	WarrantyYears      int                `json:"warranty_years"`
	Today              string             `json:"today"`
	Orders             []WarrantyOrderDTO `json:"orders"`
}

func toWarrantyLookup(res *service.WarrantyLookupResult) WarrantyLookupDTO {
	out := WarrantyLookupDTO{PhoneMasked: res.PhoneMasked, CustomerNameMasked: res.CustomerNameMasked,
		WarrantyYears: warranty.WarrantyYears, Today: res.Today.Format(dateFmt), Orders: []WarrantyOrderDTO{}}
	for _, o := range res.Orders {
		end := warranty.EndDate(o.PurchaseDate)
		st, days := warranty.Status(end, res.Today)
		wo := WarrantyOrderDTO{OrderCode: service.OrderCode(o.ID), PurchaseDate: o.PurchaseDate.Format(dateFmt),
			WarrantyEndDate: end.Format(dateFmt), Status: st, DaysLeft: days, Items: []WarrantyItemDTO{}}
		for _, it := range o.Items {
			wo.Items = append(wo.Items, WarrantyItemDTO{ProductName: it.ProductName, Quantity: it.Quantity, SerialNumber: it.SerialNumber})
		}
		out.Orders = append(out.Orders, wo)
	}
	return out
}

// ---------- Auth / user ----------

// UserDTO là thông tin user (không có password_hash / token_version).
type UserDTO struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	FullName    string     `json:"full_name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toUser(u *domain.User) UserDTO {
	return UserDTO{ID: u.ID, Username: u.Username, FullName: u.FullName, Role: u.Role, IsActive: u.IsActive,
		LastLoginAt: u.LastLoginAt, CreatedAt: u.CreatedAt}
}

// ---------- Nội bộ ----------

// InternalImageDTO là ảnh trong khu nội bộ.
type InternalImageDTO struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsMain    bool   `json:"is_main"`
	SortOrder int    `json:"sort_order"`
	MimeType  string `json:"mime_type"`
	SizeBytes int    `json:"size_bytes"`
	Width     *int   `json:"width"`
	Height    *int   `json:"height"`
}

func toInternalImage(im *domain.ProductImage) InternalImageDTO {
	return InternalImageDTO{ID: im.ID, URL: imageURL(im.ID), IsMain: im.IsMain, SortOrder: im.SortOrder,
		MimeType: im.MimeType, SizeBytes: im.SizeBytes, Width: im.Width, Height: im.Height}
}

// InternalProductDTO là SP đầy đủ cho khu nội bộ.
type InternalProductDTO struct {
	ID               int64              `json:"id"`
	Name             string             `json:"name"`
	Slug             string             `json:"slug"`
	SKU              *string            `json:"sku"`
	ShortDescription string             `json:"short_description"`
	Description      string             `json:"description"`
	Specs            []domain.Spec      `json:"specs"`
	Unit             string             `json:"unit"`
	PricePublic      int64              `json:"price_public"`
	PriceInternal    *int64             `json:"price_internal"`
	IsVisible        bool               `json:"is_visible"`
	SortOrder        int                `json:"sort_order"`
	CategoryID       *int64             `json:"category_id"`
	CategoryName     *string            `json:"category_name"`
	MainImageURL     *string            `json:"main_image_url"`
	Images           []InternalImageDTO `json:"images,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

func toInternalProduct(p *domain.Product, withImages bool) InternalProductDTO {
	out := InternalProductDTO{ID: p.ID, Name: p.Name, Slug: p.Slug, SKU: p.SKU, ShortDescription: p.ShortDescription,
		Description: p.Description, Specs: p.Specs, Unit: p.Unit, PricePublic: p.PricePublic, PriceInternal: p.PriceInternal,
		IsVisible: p.IsVisible, SortOrder: p.SortOrder, CategoryID: p.CategoryID, CategoryName: p.CategoryName,
		MainImageURL: imageURLPtr(p.MainImageID),
		CreatedAt:    p.CreatedAt, UpdatedAt: p.UpdatedAt}
	if out.Specs == nil {
		out.Specs = []domain.Spec{}
	}
	if withImages {
		out.Images = []InternalImageDTO{}
		for i := range p.Images {
			out.Images = append(out.Images, toInternalImage(&p.Images[i]))
		}
	}
	return out
}

// PromotionItemDTO là 1 SP trong KM.
type PromotionItemDTO struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	PricePublic int64  `json:"price_public"`
	PromoPrice  int64  `json:"promo_price"`
	Note        string `json:"note"`
}

// PromotionDTO là chương trình KM.
type PromotionDTO struct {
	ID         int64              `json:"id"`
	Title      string             `json:"title"`
	Content    string             `json:"content"`
	StartDate  string             `json:"start_date"`
	EndDate    string             `json:"end_date"`
	IsActive   bool               `json:"is_active"`
	Status     string             `json:"status"`
	ItemsCount int                `json:"items_count"`
	Items      []PromotionItemDTO `json:"items,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

func toPromotion(p *domain.Promotion, today time.Time) PromotionDTO {
	out := PromotionDTO{ID: p.ID, Title: p.Title, Content: p.Content, StartDate: p.StartDate.Format(dateFmt),
		EndDate: p.EndDate.Format(dateFmt), IsActive: p.IsActive, Status: service.PromotionStatus(p, today),
		ItemsCount: p.ItemsCount, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	if p.Items != nil {
		out.Items = []PromotionItemDTO{}
		for _, it := range p.Items {
			out.Items = append(out.Items, PromotionItemDTO{ProductID: it.ProductID, ProductName: it.ProductName,
				PricePublic: it.PricePublic, PromoPrice: it.PromoPrice, Note: it.Note})
		}
		out.ItemsCount = len(p.Items)
	}
	return out
}

// PriceListProductDTO là 1 dòng bảng giá nội bộ.
type PriceListProductDTO struct {
	ID                    int64   `json:"id"`
	Name                  string  `json:"name"`
	Unit                  string  `json:"unit"`
	PricePublic           int64   `json:"price_public"`
	PriceInternal         *int64  `json:"price_internal"`
	CurrentPromoPrice     *int64  `json:"current_promo_price"`
	CurrentPromotionTitle *string `json:"current_promotion_title"`
}

// PriceListDTO là dữ liệu trang ẩn.
type PriceListDTO struct {
	Today      string                `json:"today"`
	Promotions []PromotionDTO        `json:"promotions"`
	Products   []PriceListProductDTO `json:"products"`
}

func toPriceList(pl *service.PriceList) PriceListDTO {
	out := PriceListDTO{Today: pl.Today.Format(dateFmt), Promotions: []PromotionDTO{}, Products: []PriceListProductDTO{}}
	for _, p := range pl.Promotions {
		out.Promotions = append(out.Promotions, toPromotion(p, pl.Today))
	}
	for _, p := range pl.Products {
		out.Products = append(out.Products, PriceListProductDTO(p))
	}
	return out
}

// CustomerDTO là khách hàng.
type CustomerDTO struct {
	ID          int64     `json:"id"`
	FullName    string    `json:"full_name"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	Note        string    `json:"note"`
	OrdersCount int       `json:"orders_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toCustomer(c *domain.Customer) CustomerDTO {
	return CustomerDTO{ID: c.ID, FullName: c.FullName, Phone: c.Phone, Address: c.Address, Note: c.Note,
		OrdersCount: c.OrdersCount, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

// OrderCustomerDTO là khách rút gọn trong đơn.
type OrderCustomerDTO struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address,omitempty"`
}

// OrderItemDTO là 1 dòng đơn.
type OrderItemDTO struct {
	ID           int64   `json:"id"`
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    int64   `json:"unit_price"`
	LineTotal    int64   `json:"line_total"`
	SerialNumber *string `json:"serial_number"`
}

// OrderDTO là đơn hàng (danh sách + chi tiết).
type OrderDTO struct {
	ID                int64            `json:"id"`
	OrderCode         string           `json:"order_code"`
	PurchaseDate      string           `json:"purchase_date"`
	Note              string           `json:"note"`
	Customer          OrderCustomerDTO `json:"customer"`
	TotalAmount       int64            `json:"total_amount"`
	ItemsCount        int              `json:"items_count"`
	WarrantyEndDate   string           `json:"warranty_end_date"`
	Status            string           `json:"status"`
	DaysLeft          int              `json:"days_left"`
	CreatedByUsername *string          `json:"created_by_username"`
	Items             []OrderItemDTO   `json:"items,omitempty"`
	CustomerReused    *bool            `json:"customer_reused,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

func toOrder(o *domain.Order, today time.Time, withAddress bool) OrderDTO {
	end := warranty.EndDate(o.PurchaseDate)
	st, days := warranty.Status(end, today)
	out := OrderDTO{ID: o.ID, OrderCode: service.OrderCode(o.ID), PurchaseDate: o.PurchaseDate.Format(dateFmt),
		Note: o.Note, TotalAmount: o.TotalAmount, ItemsCount: o.ItemsCount, WarrantyEndDate: end.Format(dateFmt),
		Status: st, DaysLeft: days, CreatedByUsername: o.CreatedByUsername, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt}
	if o.Customer != nil {
		out.Customer = OrderCustomerDTO{ID: o.Customer.ID, FullName: o.Customer.FullName, Phone: o.Customer.Phone}
		if withAddress {
			out.Customer.Address = o.Customer.Address
		}
	}
	if o.Items != nil {
		out.Items = []OrderItemDTO{}
		for _, it := range o.Items {
			out.Items = append(out.Items, OrderItemDTO{ID: it.ID, ProductID: it.ProductID, ProductName: it.ProductName,
				Quantity: it.Quantity, UnitPrice: it.UnitPrice, LineTotal: int64(it.Quantity) * it.UnitPrice,
				SerialNumber: it.SerialNumber})
		}
	}
	return out
}
