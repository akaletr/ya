package app

import (
	"encoding/base64"
	"encoding/json"
	"flag"
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
	"cmd/shortener/main.go/internal/mw"
	"cmd/shortener/main.go/internal/storage"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
)

var (
	baseURL, serverAddress, p string
)

func init() {
	flag.StringVar(&baseURL, "b", "", "base url")
	flag.StringVar(&serverAddress, "a", "", "host to listen on")
	flag.StringVar(&p, "f", "", "file path")
}

type app struct {
	db  storage.Storage
	cfg Config
}

// GetURL возвращает в ответе реальный url
func (app *app) GetURL(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")

	long, err := app.db.Read(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
	_, err = w.Write([]byte(long))
	if err != nil {
		log.Println(err)
	}
}

// AddURL добавляет в базу данных пару ключ/ссылка и отправляет в ответе короткую ссылку
func (app *app) AddURL(w http.ResponseWriter, r *http.Request) {
	longBS, err := io.ReadAll(r.Body)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	if err != nil || len(longBS) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = app.validateURL(longBS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	key := app.convertURLToKey(longBS)
	err = app.db.Write(key, string(longBS))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	shortURL := url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   key,
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL.String()))
	if err != nil {
		log.Println(err)
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

	key := app.convertURLToKey([]byte(data.URL))
	err = app.db.Write(key, data.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURL := url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   key,
	}

	host := os.Getenv("BASE_URL")

	flag.Parse()

	if baseURL != "" {
		host = baseURL
	}

	if host != "" {
		shortURL.Scheme = strings.Split(host, "://")[0]
		shortURL.Host = strings.Split(host, "://")[1]
	}

	resp := model.ShortenerResponse{Result: shortURL.String()}
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(respJSON)
	if err != nil {
		log.Println(err)
	}
}

// Start запускает сервер
func (app *app) Start() error {
	router := chi.NewRouter()

	router.Use(mw.GzipHandle)
	router.Get("/{id}", app.GetURL)
	router.Post("/", app.AddURL)
	router.Post("/api/shorten", app.Shorten)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	address := os.Getenv("SERVER_ADDRESS")
	flag.Parse()

	if serverAddress != "" {
		address = serverAddress
	}

	strs := strings.Split(address, ":")
	if len(strs) == 2 {
		port := strs[1]
		server.Addr = fmt.Sprintf(":%s", port)
	}

	return server.ListenAndServe()
}

// convertURLToKey возвращает уникальный идентификатор для строки
func (app *app) convertURLToKey(URL []byte) string {
	qq := crc32.ChecksumIEEE(URL)
	eb := big.NewInt(int64(qq))
	return base64.RawURLEncoding.EncodeToString(eb.Bytes())
}

// validateURL проверка URL на валидность
func (app *app) validateURL(URL []byte) error {
	_, err := url.ParseRequestURI(string(URL))
	if err != nil {
		return err
	}
	return nil
}

// New возвращает новый экземпляр приложения
func New() App {
	flag.Parse()

	if p != "" {
		return &app{
			db: storage.NewFileStorage(p),
		}
	}

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
