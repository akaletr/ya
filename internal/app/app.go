package app

import (
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"math/big"
	"net/http"
	"net/url"

	"cmd/shortener/main.go/internal/storage"

	"github.com/go-chi/chi/v5"
)

type app struct {
	db storage.Storage
}

// GetURL возвращает в ответе реальный url
func (app *app) GetURL(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("hello"))
	s := chi.URLParam(r, "id")

	long, err := app.db.Read(s)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Header().Set("Location", long)
	_, err = w.Write([]byte(long))
	if err != nil {
		fmt.Println(err)
	}
}

// AddURL добавляет в базу данных пару ключ/ссылка и отправляет в ответе короткую ссылку
func (app *app) AddURL(w http.ResponseWriter, r *http.Request) {
	long, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil || len(long) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := app.convertUrlToKey(long)
	err = app.db.Write(key, string(long))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	short := url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   key,
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(short.String()))
	if err != nil {
		fmt.Println(err)
	}
}

// Start запускает сервер
func (app *app) Start() error {
	router := chi.NewRouter()

	router.Post("/", app.AddURL)
	router.Get("/{id}", app.GetURL)

	return http.ListenAndServe(":8080", router)
}

// convertUrlToKey возвращает уникальный идентификатор для строки
func (app *app) convertUrlToKey(url []byte) string {
	qq := crc32.ChecksumIEEE(url)
	eb := big.NewInt(int64(qq))
	return base64.RawURLEncoding.EncodeToString(eb.Bytes())
}

// New возвращает новый экземпляр приложения
func New() App {
	return &app{
		db: storage.New(),
	}
}
