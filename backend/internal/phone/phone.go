// Package phone chuẩn hóa và che số điện thoại Việt Nam.
package phone

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ErrInvalidPhone báo số điện thoại không hợp lệ.
var ErrInvalidPhone = errors.New("số điện thoại không hợp lệ")

var validRe = regexp.MustCompile(`^0[0-9]{9,10}$`)

// Normalize đưa SĐT về dạng 0xxxxxxxxx (10–11 chữ số).
// Chấp nhận khoảng trắng . - ( ), tiền tố +84, 84, 0084.
func Normalize(s string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		switch r {
		case ' ', '.', '-', '(', ')', '\t':
			continue
		}
		b.WriteRune(r)
	}
	v := b.String()
	if v == "" {
		return "", ErrInvalidPhone
	}
	if v[0] == '+' {
		v = v[1:]
		if !strings.HasPrefix(v, "84") {
			return "", ErrInvalidPhone
		}
		v = "0" + v[2:]
	}
	for i := 0; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return "", ErrInvalidPhone
		}
	}
	switch {
	case strings.HasPrefix(v, "0084"):
		v = "0" + v[4:]
	case strings.HasPrefix(v, "84") && (len(v) == 11 || len(v) == 12):
		v = "0" + v[2:]
	}
	if !validRe.MatchString(v) {
		return "", ErrInvalidPhone
	}
	return v, nil
}

// Mask giữ 3 số đầu và 3 số cuối: 0912345678 → 091****678.
func Mask(p string) string {
	if len(p) <= 6 {
		return p
	}
	return p[:3] + strings.Repeat("*", len(p)-6) + p[len(p)-3:]
}

// MaskName giữ nguyên tên (từ cuối), các từ khác chỉ giữ ký tự đầu:
// "Nguyễn Văn An" → "N*** V*** An".
func MaskName(name string) string {
	words := strings.Fields(name)
	if len(words) <= 1 {
		return strings.Join(words, "")
	}
	out := make([]string, len(words))
	for i, w := range words {
		if i == len(words)-1 {
			out[i] = w
			continue
		}
		r, _ := utf8.DecodeRuneInString(w)
		out[i] = string(r) + "***"
	}
	return strings.Join(out, " ")
}
