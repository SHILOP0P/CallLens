package auth

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestIsHTTPS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if requestIsHTTPS(req) {
		t.Fatal("plain request detected as HTTPS")
	}
	req.TLS = &tls.ConnectionState{}
	if !requestIsHTTPS(req) {
		t.Fatal("TLS request not detected")
	}
	req.TLS = nil
	req.Header.Set("X-Forwarded-Proto", "HTTPS")
	if !requestIsHTTPS(req) {
		t.Fatal("forwarded proto not detected")
	}
	req.Header.Del("X-Forwarded-Proto")
	req.Header.Set("Forwarded", "for=127.0.0.1; proto=https")
	if !requestIsHTTPS(req) {
		t.Fatal("Forwarded header not detected")
	}
}

func TestAuthCookieSecurityAttributes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cookie := authCookie(req, "name", "value", "/", 0)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie = %+v", cookie)
	}
	if cookie.MaxAge != 0 || !cookie.Expires.IsZero() {
		t.Fatalf("cookie = %+v", cookie)
	}
	req.TLS = &tls.ConnectionState{}
	if !authCookie(req, "name", "value", "/", time.Minute).Secure {
		t.Fatal("HTTPS cookie must be Secure")
	}
}

func TestOptionalStringIPAddressAndFirstNonEmpty(t *testing.T) {
	if optionalString(" ") != nil {
		t.Fatal("blank optional string must be nil")
	}
	if got := optionalString(" value "); got == nil || *got != "value" {
		t.Fatalf("optionalString = %v", got)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	if got := clientIPAddress(req); got == nil || *got != "127.0.0.1" {
		t.Fatalf("clientIPAddress host:port = %v", got)
	}
	req.RemoteAddr = "127.0.0.1"
	if got := clientIPAddress(req); got == nil || *got != "127.0.0.1" {
		t.Fatalf("clientIPAddress IP = %v", got)
	}
	req.RemoteAddr = "invalid"
	if clientIPAddress(req) != nil {
		t.Fatal("invalid address should return nil")
	}

	if got := firstNonEmpty("", "second", "third"); got != "second" {
		t.Fatalf("firstNonEmpty = %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("firstNonEmpty empty = %q", got)
	}
}

func TestRefreshTokenFromRequestOnlyAcceptsCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: " token "})
	if got, err := refreshTokenFromRequest(req); err != nil || got != "token" {
		t.Fatalf("cookie token = %q, %v", got, err)
	}
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"refresh_token":"body"}`))
	if got, err := refreshTokenFromRequest(req); err != nil || got != "" {
		t.Fatalf("body token must be ignored, got %q, %v", got, err)
	}

	handler := NewAuthHandler(nil, time.Minute, time.Hour)
	rec := httptest.NewRecorder()
	handler.setAuthCookies(rec, httptest.NewRequest(http.MethodGet, "/", nil), "access", "refresh")
	if len(rec.Result().Cookies()) != 2 {
		t.Fatal("expected two auth cookies")
	}
}
