package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestJWTValidation(t *testing.T) {
	m, e := New("0123456789abcdef0123456789abcdef", "localhost:6379", time.Minute)
	if e != nil {
		t.Fatal(e)
	}
	a, _ := m.Issue(7)
	b, _ := m.Issue(7)
	if a == b {
		t.Fatal("tokens must have unique jti")
	}
	c, e := m.Parse(a)
	if e != nil || c.Subject != "7" {
		t.Fatal("valid subject", e)
	}
	for _, raw := range []string{"", a + "x", "invalid"} {
		if _, e = m.Parse(raw); e == nil {
			t.Fatal("accepted invalid token")
		}
	}
	for _, tc := range []struct {
		name     string
		alg      jwt.SigningMethod
		iss, aud string
		exp      *jwt.NumericDate
	}{
		{"expired", jwt.SigningMethodHS256, Issuer, Issuer, jwt.NewNumericDate(time.Now().Add(-time.Second))},
		{"missing exp", jwt.SigningMethodHS256, Issuer, Issuer, nil},
		{"issuer", jwt.SigningMethodHS256, "other", Issuer, jwt.NewNumericDate(time.Now().Add(time.Minute))},
		{"audience", jwt.SigningMethodHS256, Issuer, "other", jwt.NewNumericDate(time.Now().Add(time.Minute))},
		{"algorithm", jwt.SigningMethodHS384, Issuer, Issuer, jwt.NewNumericDate(time.Now().Add(time.Minute))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := jwt.NewWithClaims(tc.alg, Claims{jwt.RegisteredClaims{Subject: "7", ID: "unique", Issuer: tc.iss, Audience: jwt.ClaimStrings{tc.aud}, ExpiresAt: tc.exp}}).SignedString(m.Key)
			if _, e := m.Parse(raw); e == nil {
				t.Fatal("accepted invalid claims")
			}
		})
	}
	if BlacklistKey(a) == BlacklistKey(b) {
		t.Fatal("blacklist keys collide")
	}
}
