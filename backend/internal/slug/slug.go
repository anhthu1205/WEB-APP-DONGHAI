// Package slug sinh slug URL từ chuỗi tiếng Việt.
package slug

import (
	"regexp"
	"strings"
)

var viMap = func() map[rune]rune {
	groups := map[rune]string{
		'a': "àáạảãâầấậẩẫăằắặẳẵ",
		'e': "èéẹẻẽêềếệểễ",
		'i': "ìíịỉĩ",
		'o': "òóọỏõôồốộổỗơờớợởỡ",
		'u': "ùúụủũưừứựửữ",
		'y': "ỳýỵỷỹ",
		'd': "đ",
	}
	m := map[rune]rune{}
	for base, s := range groups {
		for _, r := range s {
			m[r] = base
		}
	}
	return m
}()

// Valid là định dạng slug hợp lệ (khớp ràng buộc CHECK trong DB).
var Valid = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Make: "Máy lọc nước RO Đông Hải DH-9" → "may-loc-nuoc-ro-dong-hai-dh-9".
func Make(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if m, ok := viMap[r]; ok {
			r = m
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		// dấu kết hợp (NFD) bị bỏ qua, ký tự khác thành gạch nối
		if r >= 0x0300 && r <= 0x036f {
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.TrimRight(b.String(), "-")
	if len(out) > 120 {
		out = strings.TrimRight(out[:120], "-")
	}
	return out
}
