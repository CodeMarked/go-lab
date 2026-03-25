package auth

import (
	"strings"
	"testing"
	"time"
)

func TestMintAndParseRoundTrip(t *testing.T) {
	ts, err := NewTokenService(strings.Repeat("a", 32), "", "iss", "aud", 15*time.Minute)
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
	_, err := NewTokenService("short", "", "iss", "aud", time.Minute)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseAcceptsJWTSignedWithPreviousSecret(t *testing.T) {
	oldS := strings.Repeat("o", 32)
	newS := strings.Repeat("n", 32)
	oldTS, err := NewTokenService(oldS, "", "iss", "aud", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	tok, _, err := oldTS.MintAccessToken("sub:1")
	if err != nil {
		t.Fatal(err)
	}
	both, err := NewTokenService(newS, oldS, "iss", "aud", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := both.ParseAccessToken(tok); err != nil {
		t.Fatalf("previous secret should verify: %v", err)
	}
	newTok, _, err := both.MintAccessToken("sub:2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := both.ParseAccessToken(newTok); err != nil {
		t.Fatal(err)
	}
}

func TestParseRejectsWrongSecrets(t *testing.T) {
	a := strings.Repeat("a", 32)
	b := strings.Repeat("b", 32)
	tsA, _ := NewTokenService(a, "", "iss", "aud", time.Minute)
	tok, _, _ := tsA.MintAccessToken("x")
	tsB, _ := NewTokenService(b, "", "iss", "aud", time.Minute)
	if _, err := tsB.ParseAccessToken(tok); err == nil {
		t.Fatal("expected verify failure")
	}
}
