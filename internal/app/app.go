package app

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"

	"cmd/shortener/main.go/internal/auth"
	"cmd/shortener/main.go/internal/config"
	"cmd/shortener/main.go/internal/gziper"
	"cmd/shortener/main.go/internal/model"
	"cmd/shortener/main.go/internal/storage"

	"github.com/go-chi/chi/v5"
)

type app struct {
	db   storage.Storage
	cfg  config.Config
	auth auth.Auth
}

// GetURL возвращает в ответе реальный url
func (app *app) GetURL(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	long, err := app.db.Read(key)
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
	c, err := r.Cookie("user")
	if err != nil || !app.auth.Check(c) {
		value, e := app.auth.NewToken()
		if e != nil {
			log.Println(err)
			//w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c = &http.Cookie{
			Name:  "user",
			Value: hex.EncodeToString(value),
			Path:  "/",
		}
		http.SetCookie(w, c)
	}

	id, err := app.auth.GetID(c)
	if err != nil {
		log.Println(err)
	}

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
	err = app.db.Write(id, key, string(longBS))
	if err != nil {
		err, _ := err.(*pq.Error)
		log.Println(err.Code)
		if err.Code != "23505" {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusConflict)
	}

	w.WriteHeader(http.StatusCreated)
	shortURL := fmt.Sprintf("%s/%s", app.cfg.BaseURL, key)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		log.Println(err)
	}
}

// Shorten обрабатываут запрос и формирует ответ в json
func (app *app) Shorten(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil || !app.auth.Check(c) {
		value, e := app.auth.NewToken()
		if e != nil {
			log.Println(err)
			//w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c = &http.Cookie{
			Name:  "user",
			Value: hex.EncodeToString(value),
			Path:  "/",
		}
		http.SetCookie(w, c)
	}

	id, err := app.auth.GetID(c)
	if err != nil {
		log.Println(err)
	}

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

	err = app.validateURL([]byte(data.URL))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	key := app.convertURLToKey([]byte(data.URL))
	err = app.db.Write(id, key, data.URL)
	if err != nil {
		err, _ := err.(*pq.Error)
		log.Println(err.Code)
		if err.Code != "23505" {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	shortURL := fmt.Sprintf("%s/%s", app.cfg.BaseURL, key)

	resp := model.ShortenerResponse{Result: shortURL}
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respJSON)
	if err != nil {

		log.Println(err)
	}
}

func (app *app) GetAllURLs(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		c = &http.Cookie{}
	}

	id, err := app.auth.GetID(c)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")

	data, err := app.db.ReadAll(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	result := model.AllShortenerRequest{}
	for key, value := range data {
		item := model.Item{
			ShortURL:    fmt.Sprintf("%s/%s", app.cfg.BaseURL, key),
			OriginalURL: value,
		}
		result = append(result, item)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resultJSON)
	if err != nil {
		log.Println(err)
	}
}

func (app *app) DatabasePing(w http.ResponseWriter, r *http.Request) {
	if app.cfg.DatabaseDSN != "" {
		db, err := sql.Open("postgres", app.cfg.DatabaseDSN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (app *app) Batch(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		c = &http.Cookie{}
	}

	id, err := app.auth.GetID(c)
	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(r.Body)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var data model.BatchRequest
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dataBatch := model.DataBatch{}
	result := model.BatchResponse{}

	for _, item := range data {
		short := app.convertURLToKey([]byte(item.OriginalURL))
		resultItem := model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", app.cfg.BaseURL, short),
		}

		dataBatchItem := model.DataBatchItem{
			ID:    id,
			Short: short,
			Long:  item.OriginalURL,
		}

		dataBatch = append(dataBatch, dataBatchItem)
		result = append(result, resultItem)
	}

	err = app.db.WriteBatch(dataBatch)
	if err != nil {
		err, _ := err.(*pq.Error)
		log.Println(err.Code)
		if err.Code != "23505" {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
	}

	w.Header().Set("Content-Type", "application/json")

	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resultJSON)
	if err != nil {
		log.Println(err)
	}
}

// Start запускает сервер
func (app *app) Start() error {
	// подготавливаем базу к работе
	err := app.db.Start()
	if err != nil {
		return err
	}

	router := chi.NewRouter()

	router.Use(gziper.GzipHandle)

	router.Get("/{key}", app.GetURL)
	router.Post("/", app.AddURL)
	router.Post("/api/shorten", app.Shorten)
	router.Post("/api/shorten/batch", app.Batch)
	router.Get("/api/user/urls", app.GetAllURLs)
	router.Get("/ping", app.DatabasePing)

	server := http.Server{
		Addr:    app.cfg.ServerAddress,
		Handler: router,
	}

	return server.ListenAndServe()
}

// convertURLToKey возвращает уникальный идентификатор для строки
func (app *app) convertURLToKey(URL []byte) string {
	qq := crc32.ChecksumIEEE(URL)
	eb := big.NewInt(int64(qq))
	return base64.RawURLEncoding.EncodeToString(eb.Bytes())
}

// validateURL проверяет URL на валидность
func (app *app) validateURL(URL []byte) error {
	_, err := url.ParseRequestURI(string(URL))
	if err != nil {
		return err
	}
	return nil
}

// New возвращает новый экземпляр приложения
func New(cfg config.Config) (App, error) {
	if cfg.DatabaseDSN != "" {
		db, err := storage.NewPostgresDatabase(cfg.FileStoragePath)
		if err != nil {
			return &app{}, err
		}

		return &app{
			db:   db,
			cfg:  cfg,
			auth: auth.New(cfg.SecretKey),
		}, nil
	}

	if cfg.FileStoragePath != "" {
		return &app{
			db:   storage.NewFileStorage(cfg.FileStoragePath),
			cfg:  cfg,
			auth: auth.New(cfg.SecretKey),
		}, nil
	}

	return &app{
		db:   storage.New(),
		cfg:  cfg,
		auth: auth.New(cfg.SecretKey),
	}, nil
}
