package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost là cost băm mật khẩu (có thể giảm trong test để chạy nhanh).
var BcryptCost = 12

// Độ dài mật khẩu cho phép (bcrypt chỉ dùng 72 byte đầu).
const (
	MinPasswordLen = 8
	MaxPasswordLen = 72
)

// HashPassword băm mật khẩu bằng bcrypt.
func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("băm mật khẩu: %w", err)
	}
	return string(h), nil
}

// CheckPassword so khớp mật khẩu với hash.
func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// dummyHash dùng để so khớp giả khi không có user (cân bằng thời gian phản hồi).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dong-hai-dummy-password"), bcrypt.MinCost)

// BurnPasswordCheck tốn thời gian tương đương một lần so khớp (chống dò username).
func BurnPasswordCheck(pw string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(pw))
}
