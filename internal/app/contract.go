package app

import "net/http"

type App interface {
	Start() error
	AddURL(w http.ResponseWriter, r *http.Request)
	GetURL(w http.ResponseWriter, r *http.Request)
}
