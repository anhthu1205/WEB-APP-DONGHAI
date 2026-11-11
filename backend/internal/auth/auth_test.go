package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const secret = "0123456789abcdef0123456789abcdef-test"

func TestSignParse(t *testing.T) {
	is := NewIssuer(secret, time.Hour, nil)
	tok, exp, err := is.Sign(42, "admin", 3)
	if err != nil {
		t.Fatal(err)
	}
	if time.Until(exp) < 59*time.Minute {
		t.Fatalf("exp sai: %v", exp)
	}
	c, err := is.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := c.UserID()
	if id != 42 || c.Role != "admin" || c.TV != 3 {
		t.Fatalf("claims sai: %+v", c)
	}
}

func TestParseExpired(t *testing.T) {
	past := func() time.Time { return time.Now().Add(-2 * time.Hour) }
	tok, _, _ := NewIssuer(secret, time.Hour, past).Sign(1, "user", 0)
	if _, err := NewIssuer(secret, time.Hour, nil).Parse(tok); err != ErrInvalidToken {
		t.Fatalf("token hết hạn phải lỗi, được %v", err)
	}
}

func TestParseWrongSecretOrAlg(t *testing.T) {
	is := NewIssuer(secret, time.Hour, nil)
	other, _, _ := NewIssuer(strings.Repeat("z", 40), time.Hour, nil).Sign(1, "admin", 0)
	if _, err := is.Parse(other); err == nil {
		t.Fatal("HS256 khác secret phải lỗi")
	}
	claims := Claims{Role: "admin", RegisteredClaims: jwt.RegisteredClaims{
		Subject: "1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := is.Parse(none); err == nil {
		t.Fatal("alg=none phải lỗi")
	}
	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(secret))
	if _, err := is.Parse(hs512); err == nil {
		t.Fatal("HS512 phải lỗi (chỉ nhận HS256)")
	}
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "1"}}).SignedString([]byte(secret))
	if _, err := is.Parse(noExp); err == nil {
		t.Fatal("thiếu exp phải lỗi")
	}
	if _, err := is.Parse("rác.rác.rác"); err == nil {
		t.Fatal("token rác phải lỗi")
	}
}

func TestPassword(t *testing.T) {
	BcryptCost = bcrypt.MinCost
	defer func() { BcryptCost = 12 }()
	h, err := HashPassword("matkhau123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(h, "matkhau123") {
		t.Fatal("mật khẩu đúng phải khớp")
	}
	if CheckPassword(h, "matkhau124") {
		t.Fatal("mật khẩu sai không được khớp")
	}
}
