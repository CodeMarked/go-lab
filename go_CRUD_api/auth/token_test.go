package auth

import (
	"strings"
	"testing"
	"time"
)

func TestMintAndParseRoundTrip(t *testing.T) {
	ts, err := NewTokenService(strings.Repeat("a", 32), "iss", "aud", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	tok, exp, err := ts.MintAccessToken("client:demo")
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}
	if !exp.After(time.Now()) {
		t.Fatal("expiry not in future")
	}
	claims, err := ts.ParseAccessToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "client:demo" {
		t.Fatalf("subject: %q", claims.Subject)
	}
}

func TestNewTokenServiceRejectsShortSecret(t *testing.T) {
	_, err := NewTokenService("short", "iss", "aud", time.Minute)
	if err == nil {
		t.Fatal("expected error")
	}
}
