package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cmd/shortener/main.go/internal/model"
	"cmd/shortener/main.go/internal/storage"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
)

type app struct {
	db  storage.Storage
	cfg Config
}

// GetURL возвращает в ответе реальный url
func (app *app) GetURL(w http.ResponseWriter, r *http.Request) {
	s := chi.URLParam(r, "id")

	long, err := app.db.Read(s)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Location", long)
	//w.WriteHeader(http.StatusTemporaryRedirect)
	fmt.Println("99999", long)
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

// Shorten обработка запроса и формирование ответа в json
func (app *app) Shorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var data model.ShortenerRequest
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := app.convertUrlToKey([]byte(data.URL))
	err = app.db.Write(key, data.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	short := url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   key,
	}

	host := os.Getenv("BASE_URL")
	if host != "" {
		short.Scheme = strings.Split(host, "://")[0]
		short.Host = strings.Split(host, "://")[1]
	}

	resp := model.ShortenerResponse{Result: short.String()}
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respJSON)
}

// Start запускает сервер
func (app *app) Start() error {
	router := chi.NewRouter()

	router.Post("/", app.AddURL)
	router.Get("/{id}", app.GetURL)
	router.Post("/api/shorten", app.Shorten)

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	address := os.Getenv("SERVER_ADDRESS")
	strs := strings.Split(address, ":")

	if len(strs) == 2 {
		port := strs[1]
		server.Addr = fmt.Sprintf(":%s", port)
	}

	return server.ListenAndServe()
}

// convertUrlToKey возвращает уникальный идентификатор для строки
func (app *app) convertUrlToKey(url []byte) string {
	qq := crc32.ChecksumIEEE(url)
	eb := big.NewInt(int64(qq))
	return base64.RawURLEncoding.EncodeToString(eb.Bytes())
}

// New возвращает новый экземпляр приложения
func New() App {
	path := os.Getenv("FILE_STORAGE_PATH")
	if path != "" {
		return &app{
			db: storage.NewFileStorage(path),
		}
	}

	return &app{
		db: storage.New(),
	}
}

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}
