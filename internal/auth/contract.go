package auth

import "net/http"

type Auth interface {
	Check(cookie *http.Cookie) bool
	NewToken() ([]byte, error)
	GetID(cookie *http.Cookie) (string, error)

	// CookieHandler middleware для чтения и установки куков
	CookieHandler(next http.Handler) http.Handler
}
