package domain

import "time"

// Vai trò người dùng.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// User là tài khoản nội bộ.
type User struct {
	ID           int64
	Username     string
	FullName     string
	PasswordHash string
	Role         string
	IsActive     bool
	TokenVersion int
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsAdmin cho biết user có quyền admin.
func (u *User) IsAdmin() bool { return u != nil && u.Role == RoleAdmin }

// Company là thông tin công ty (1 dòng).
type Company struct {
	Name         string
	Slogan       string
	About        string
	Address      string
	Phone        string
	Email        string
	WorkingHours string
	FacebookURL  string
	TiktokURL    string
	ZaloURL      string
	ShopeeURL    string
	UpdatedBy    *int64
	UpdatedAt    time.Time
}

// Spec là một dòng thông số kỹ thuật.
type Spec struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Product là sản phẩm.
type Product struct {
	ID               int64
	Name             string
	Slug             string
	SKU              *string
	ShortDescription string
	Description      string
	Specs            []Spec
	Unit             string
	PricePublic      int64
	PriceInternal    *int64
	IsVisible        bool
	SortOrder        int
	MainImageID      *string
	Images           []ProductImage
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ProductImage là metadata ảnh (Data chỉ nạp khi phục vụ ảnh).
type ProductImage struct {
	ID        string
	ProductID int64
	MimeType  string
	SizeBytes int
	Width     *int
	Height    *int
	IsMain    bool
	SortOrder int
	CreatedAt time.Time
	Data      []byte
}

// Promotion là chương trình khuyến mãi.
type Promotion struct {
	ID         int64
	Title      string
	Content    string
	StartDate  time.Time
	EndDate    time.Time
	IsActive   bool
	CreatedBy  *int64
	Items      []PromotionItem
	ItemsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PromotionItem là giá KM của 1 SP trong chương trình.
type PromotionItem struct {
	ProductID   int64
	ProductName string
	PricePublic int64
	PromoPrice  int64
	Note        string
}

// PriceListProduct là 1 dòng bảng giá nội bộ.
type PriceListProduct struct {
	ID                    int64
	Name                  string
	Unit                  string
	PricePublic           int64
	PriceInternal         *int64
	CurrentPromoPrice     *int64
	CurrentPromotionTitle *string
}

// Customer là khách hàng.
type Customer struct {
	ID          int64
	FullName    string
	Phone       string
	Address     string
	Note        string
	OrdersCount int
	CreatedBy   *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Order là đơn hàng.
type Order struct {
	ID                int64
	CustomerID        int64
	Customer          *Customer
	PurchaseDate      time.Time
	Note              string
	CreatedBy         *int64
	CreatedByUsername *string
	Items             []OrderItem
	TotalAmount       int64
	ItemsCount        int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// OrderItem là 1 dòng sản phẩm trong đơn.
type OrderItem struct {
	ID           int64
	ProductID    int64
	ProductName  string
	Quantity     int
	UnitPrice    int64
	SerialNumber *string
}

// Page là kết quả phân trang.
type Page[T any] struct {
	Items    []T
	Total    int
	Page     int
	PageSize int
}

// PageParams là tham số phân trang đã chuẩn hóa.
type PageParams struct {
	Page     int
	PageSize int
}

// NewPageParams chuẩn hóa page/page_size (mặc định 1/20, tối đa 100).
func NewPageParams(page, size int) PageParams {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return PageParams{Page: page, PageSize: size}
}

// Offset trả vị trí bắt đầu.
func (p PageParams) Offset() int { return (p.Page - 1) * p.PageSize }
