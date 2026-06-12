package auth

import (
	"net/http"
	"strings"
	"time"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"

	accessTokenCookiePath  = "/"
	refreshTokenCookiePath = "/api/v1/auth"
)

func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, r *http.Request, accessToken string, refreshToken string) {
	http.SetCookie(w, authCookie(r, accessTokenCookieName, accessToken, accessTokenCookiePath, h.accessTokenTTL))
	http.SetCookie(w, authCookie(r, refreshTokenCookieName, refreshToken, refreshTokenCookiePath, h.refreshTokenTTL))
}

func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, expiredAuthCookie(r, accessTokenCookieName, accessTokenCookiePath))
	http.SetCookie(w, expiredAuthCookie(r, refreshTokenCookieName, refreshTokenCookiePath))
}

func authCookie(r *http.Request, name string, value string, path string, ttl time.Duration) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		Secure:   requestIsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	}

	if ttl > 0 {
		cookie.MaxAge = int(ttl.Seconds())
		cookie.Expires = time.Now().UTC().Add(ttl)
	}

	return cookie
}

func expiredAuthCookie(r *http.Request, name string, path string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
		HttpOnly: true,
		Secure:   requestIsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	}
}

func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}

	forwarded := strings.ToLower(r.Header.Get("Forwarded"))
	return strings.Contains(forwarded, "proto=https")
}
