// Package warranty tính hạn bảo hành theo ngày dân sự.
//
// Quy ước: mọi ngày là time.Time 00:00 UTC (ngày dân sự). "Hôm nay" lấy theo
// Asia/Ho_Chi_Minh rồi dựng lại thành 00:00 UTC bằng Today.
package warranty

import (
	"time"
	_ "time/tzdata" // image distroless không có tzdata
)

// WarrantyYears là số năm bảo hành kể từ ngày mua.
const WarrantyYears = 2

// Trạng thái bảo hành.
const (
	StatusActive  = "active"
	StatusExpired = "expired"
)

var vnLoc = mustLoc("Asia/Ho_Chi_Minh")

func mustLoc(name string) *time.Location {
	l, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("ICT", 7*3600)
	}
	return l
}

// Location trả múi giờ nghiệp vụ (Asia/Ho_Chi_Minh).
func Location() *time.Location { return vnLoc }

// CivilDate đưa t về ngày dân sự 00:00 UTC theo ngày của chính t (không đổi múi giờ).
func CivilDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Today trả ngày hôm nay theo giờ Việt Nam, dạng 00:00 UTC.
func Today(now time.Time) time.Time {
	return CivilDate(now.In(vnLoc))
}

func isLeap(y int) bool { return y%4 == 0 && (y%100 != 0 || y%400 == 0) }

// EndDate = ngày mua + 2 năm; 29/02 rơi vào năm không nhuận → 28/02.
func EndDate(purchase time.Time) time.Time {
	y, m, d := purchase.Date()
	ty := y + WarrantyYears
	if m == time.February && d == 29 && !isLeap(ty) {
		d = 28
	}
	return time.Date(ty, m, d, 0, 0, 0, 0, time.UTC)
}

// Status: active nếu today <= end (còn bảo hành đến hết ngày end), ngược lại expired.
// daysLeft = max(0, end - today) tính theo ngày.
func Status(end, today time.Time) (status string, daysLeft int) {
	e, t := CivilDate(end), CivilDate(today)
	if t.After(e) {
		return StatusExpired, 0
	}
	return StatusActive, int(e.Sub(t).Hours() / 24)
}
