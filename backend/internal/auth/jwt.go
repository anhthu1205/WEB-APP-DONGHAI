// Package auth: JWT HS256 và băm mật khẩu bcrypt.
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken báo token sai/hết hạn/sai thuật toán.
var ErrInvalidToken = errors.New("token không hợp lệ")

// Claims của JWT: sub = user id, role, tv = token_version.
type Claims struct {
	Role string `json:"role"`
	TV   int    `json:"tv"`
	jwt.RegisteredClaims
}

// UserID trả id người dùng từ claim sub.
func (c *Claims) UserID() (int64, error) {
	return strconv.ParseInt(c.Subject, 10, 64)
}

// Issuer ký và kiểm tra JWT.
type Issuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewIssuer tạo Issuer; now = nil → time.Now.
func NewIssuer(secret string, ttl time.Duration, now func() time.Time) *Issuer {
	if now == nil {
		now = time.Now
	}
	return &Issuer{secret: []byte(secret), ttl: ttl, now: now}
}

// Sign tạo token mới cho người dùng.
func (i *Issuer) Sign(userID int64, role string, tokenVersion int) (string, time.Time, error) {
	now := i.now()
	exp := now.Add(i.ttl)
	c := Claims{
		Role: role,
		TV:   tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("ký jwt: %w", err)
	}
	return s, exp, nil
}

// Parse kiểm tra chữ ký (chỉ HS256), hạn dùng và trả claims.
func (i *Issuer) Parse(token string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return i.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(i.now),
	)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if _, err := c.UserID(); err != nil {
		return nil, ErrInvalidToken
	}
	return &c, nil
}
