package auth

import "net/http"

type Auth interface {
	Check(cookie *http.Cookie) bool
	NewToken() ([]byte, error)
	GetID(cookie *http.Cookie) (string, error)

	CookieHandler(next http.Handler) http.Handler
}
