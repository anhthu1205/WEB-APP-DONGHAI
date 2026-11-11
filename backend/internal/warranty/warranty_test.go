package warranty

import (
	"testing"
	"time"
)

func d(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestEndDate(t *testing.T) {
	cases := map[string]string{
		"2025-03-10": "2027-03-10",
		"2024-12-31": "2026-12-31",
		"2024-02-29": "2026-02-28",
		"2026-02-28": "2028-02-28",
	}
	for in, want := range cases {
		p, err := time.Parse("2006-01-02", in)
		if err != nil {
			continue
		}
		if got := EndDate(p).Format("2006-01-02"); got != want {
			t.Errorf("EndDate(%s) = %s; muốn %s", in, got, want)
		}
	}
	// 29/02 sang năm nhuận đích giữ nguyên 29/02
	if got := EndDate(d("2028-02-29")).Format("2006-01-02"); got != "2030-02-28" {
		t.Errorf("EndDate(2028-02-29) = %s", got)
	}
}

func TestStatus(t *testing.T) {
	today := d("2026-11-11")
	cases := []struct {
		purchase string
		status   string
		days     int
	}{
		{"2024-11-11", StatusActive, 0},   // T-W3: ngày cuối vẫn còn
		{"2024-11-10", StatusExpired, 0},  // T-W4
		{"2025-03-10", StatusActive, 119}, // T-W4
		{"2024-08-15", StatusExpired, 0},
		{"2024-02-29", StatusExpired, 0},
		{"2026-11-11", StatusActive, 731}, // qua 29/02/2028
	}
	for _, c := range cases {
		s, n := Status(EndDate(d(c.purchase)), today)
		if s != c.status || n != c.days {
			t.Errorf("mua %s: được %s/%d; muốn %s/%d", c.purchase, s, n, c.status, c.days)
		}
	}
}

func TestToday(t *testing.T) {
	// 2026-11-10 17:30 UTC = 2026-11-11 00:30 giờ VN
	now := time.Date(2026, 11, 10, 17, 30, 0, 0, time.UTC)
	if got := Today(now); !got.Equal(d("2026-11-11")) {
		t.Errorf("Today = %v", got)
	}
	now = time.Date(2026, 11, 10, 16, 59, 0, 0, time.UTC)
	if got := Today(now); !got.Equal(d("2026-11-10")) {
		t.Errorf("Today = %v", got)
	}
}
