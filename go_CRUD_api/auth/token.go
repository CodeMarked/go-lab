package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService issues and verifies access tokens (HS256).
type TokenService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

// Claims embedded in access JWTs.
type Claims struct {
	jwt.RegisteredClaims
}

// NewTokenService builds a verifier/issuer. secret must be non-empty.
func NewTokenService(secret, issuer, audience string, ttl time.Duration) (*TokenService, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret too short")
	}
	if issuer == "" || audience == "" {
		return nil, errors.New("issuer and audience required")
	}
	return &TokenService{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}, nil
}

// MintAccessToken returns a signed JWT access token.
func (s *TokenService) MintAccessToken(subject string) (string, time.Time, error) {
	if subject == "" {
		return "", time.Time{}, errors.New("subject required")
	}
	now := time.Now().UTC()
	exp := now.Add(s.ttl)
	jti, err := randomJTI()
	if err != nil {
		return "", time.Time{}, err
	}
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signed, err := t.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// ParseAccessToken validates signature, expiry, issuer, and audience.
func (s *TokenService) ParseAccessToken(tokenString string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithAudience(s.audience), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func randomJTI() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
