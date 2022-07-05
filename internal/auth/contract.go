package auth

import "net/http"

type Auth interface {
	Check(cookie *http.Cookie) bool
	NewToken() (error, []byte)
	CookieHandler(next http.Handler) http.Handler
	GetID(cookie *http.Cookie) (error, string)
}
