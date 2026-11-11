package phone

import "testing"

func TestNormalizeOK(t *testing.T) {
	cases := map[string]string{
		"0912 345 678":     "0912345678",
		"+84 912 345 678":  "0912345678",
		"+84912345678":     "0912345678",
		"84912345678":      "0912345678",
		"0084 912.345.678": "0912345678",
		"0912-345-678":     "0912345678",
		"(024) 3826 1234":  "02438261234",
		"842438261234":     "02438261234",
		" 0987654321 ":     "0987654321",
	}
	for in, want := range cases {
		got, err := Normalize(in)
		if err != nil || got != want {
			t.Errorf("Normalize(%q) = %q, %v; muốn %q", in, got, err, want)
		}
	}
}

func TestNormalizeErr(t *testing.T) {
	for _, in := range []string{"", "abc", "12345", "09123456789012", "+84+912", "+1 202 555 0100", "0912a45678", "912345678"} {
		if got, err := Normalize(in); err != ErrInvalidPhone {
			t.Errorf("Normalize(%q) = %q, %v; muốn lỗi", in, got, err)
		}
	}
}

func TestMask(t *testing.T) {
	if got := Mask("0912345678"); got != "091****678" {
		t.Errorf("Mask = %q", got)
	}
	if got := Mask("02438261234"); got != "024*****234" {
		t.Errorf("Mask 11 số = %q", got)
	}
}

func TestMaskName(t *testing.T) {
	cases := map[string]string{
		"Nguyễn Văn An":   "N*** V*** An",
		"An":              "An",
		"Đặng  Thị   Hoa": "Đ*** T*** Hoa",
		"Ánh Tuyết":       "Á*** Tuyết",
		"":                "",
	}
	for in, want := range cases {
		if got := MaskName(in); got != want {
			t.Errorf("MaskName(%q) = %q; muốn %q", in, got, want)
		}
	}
}
