package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const Issuer = "cruisecontrol"

var ErrInvalid = errors.New("invalid token")

type Claims struct{ jwt.RegisteredClaims }
type Manager struct {
	Key   []byte
	TTL   time.Duration
	Redis *redis.Client
}

func New(secret, address string, ttl time.Duration) (*Manager, error) {
	if len(secret) < 32 || address == "" || ttl < time.Second {
		return nil, errors.New("JWT_SECRET (32+ bytes), REDIS_ADDR and positive JWT_TTL required")
	}
	return &Manager{Key: []byte(secret), TTL: ttl, Redis: redis.NewClient(&redis.Options{Addr: address, DialTimeout: 2 * time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second})}, nil
}
func (m *Manager) Issue(id uint) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	now := time.Now()
	c := Claims{jwt.RegisteredClaims{Issuer: Issuer, Subject: strconv.FormatUint(uint64(id), 10), Audience: jwt.ClaimStrings{Issuer}, ExpiresAt: jwt.NewNumericDate(now.Add(m.TTL)), IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ID: hex.EncodeToString(b)}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.Key)
}
func (m *Manager) Parse(raw string) (*Claims, error) {
	c := new(Claims)
	token, err := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (interface{}, error) { return m.Key, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(Issuer), jwt.WithAudience(Issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid || c.ID == "" {
		return nil, ErrInvalid
	}
	id, err := strconv.ParseUint(c.Subject, 10, 32)
	if err != nil || id == 0 {
		return nil, ErrInvalid
	}
	return c, nil
}
func BlacklistKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("jwt:blacklist:%x", h)
}
func (m *Manager) Revoked(ctx context.Context, raw string) (bool, error) {
	n, e := m.Redis.Exists(ctx, BlacklistKey(raw)).Result()
	return n > 0, e
}
func (m *Manager) Revoke(ctx context.Context, raw string, c *Claims) error {
	ttl := time.Until(c.ExpiresAt.Time)
	if ttl <= 0 {
		return ErrInvalid
	}
	return m.Redis.Set(ctx, BlacklistKey(raw), c.Subject, ttl).Err()
}
