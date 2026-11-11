package slug

import "testing"

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Máy lọc nước RO Đông Hải DH-9": "may-loc-nuoc-ro-dong-hai-dh-9",
		"  Nồi cơm điện  cao tần!! ":    "noi-com-dien-cao-tan",
		"Bếp từ đôi DH-IH22":            "bep-tu-doi-dh-ih22",
		"ĐƯỜNG ỐNG":                     "duong-ong",
		"!!!":                           "",
	}
	for in, want := range cases {
		got := Make(in)
		if got != want {
			t.Errorf("Make(%q) = %q; muốn %q", in, got, want)
		}
		if got != "" && !Valid.MatchString(got) {
			t.Errorf("slug %q không hợp lệ", got)
		}
	}
}
