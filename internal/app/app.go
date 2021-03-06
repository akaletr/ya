package app

import (
	"cmd/shortener/main.go/internal/auth"
	"cmd/shortener/main.go/internal/config"
	"cmd/shortener/main.go/internal/gziper"
	"cmd/shortener/main.go/internal/model"
	"cmd/shortener/main.go/internal/storage"
	"cmd/shortener/main.go/internal/workerpool"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type app struct {
	db   storage.Storage
	cfg  config.Config
	auth auth.Auth
	pool *workerpool.Pool
}

// GetURL возвращает в ответе реальный url
func (app *app) GetURL(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	note, err := app.db.Read(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if note.Deleted {
		w.WriteHeader(http.StatusGone)
		return
	}
	w.Header().Set("Location", note.Long)
	w.WriteHeader(http.StatusTemporaryRedirect)
	_, err = w.Write([]byte(note.Long))
	if err != nil {
		log.Println(err)
	}
}

// AddURL добавляет в базу данных пару ключ/ссылка и отправляет в ответе короткую ссылку
func (app *app) AddURL(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		c = &http.Cookie{}
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

	short := app.convertURLToKey(longBS)

	note := model.Note{
		ID:    id,
		Short: short,
		Long:  string(longBS),
	}
	err = app.db.Write(note)
	if err != nil {
		e, ok := err.(*storage.Error)
		if ok && e.Code == storage.CONFLICT {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	shortURL := fmt.Sprintf("%s/%s", app.cfg.BaseURL, note.Short)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		log.Println(err)
	}
}

// Shorten обрабатываут запрос и формирует ответ в json
func (app *app) Shorten(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		c = &http.Cookie{}
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

	short := app.convertURLToKey([]byte(data.URL))
	note := model.Note{
		ID:    id,
		Short: short,
		Long:  data.URL,
	}
	err = app.db.Write(note)
	if err != nil {
		e, ok := err.(*storage.Error)
		if ok && e.Code == storage.CONFLICT {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	shortURL := fmt.Sprintf("%s/%s", app.cfg.BaseURL, short)

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

	result := make([]model.Item, 0)
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
	err := app.db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	var data []model.BatchRequestItem
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		dataBatch []model.DataBatchItem
		result    []model.BatchResponseItem
	)

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
		e, ok := err.(*storage.Error)
		if ok && e.Code == storage.CONFLICT {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
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

func (app *app) Delete(w http.ResponseWriter, r *http.Request) {
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

	var data []string
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, item := range data {
		note := model.Note{
			ID:    id,
			Short: item,
		}

		job := workerpool.NewJob(app.db.Delete, note)
		app.pool.AddJob(job)
	}

	w.WriteHeader(http.StatusAccepted)
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
	router.Use(app.auth.CookieHandler)

	router.Get("/{key}", app.GetURL)
	router.Post("/", app.AddURL)
	router.Post("/api/shorten", app.Shorten)
	router.Post("/api/shorten/batch", app.Batch)
	router.Get("/api/user/urls", app.GetAllURLs)
	router.Get("/ping", app.DatabasePing)
	router.Delete("/api/user/urls", app.Delete)

	server := http.Server{
		Addr:    app.cfg.ServerAddress,
		Handler: router,
	}

	go app.pool.RunBackground()

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
		db, err := storage.NewPostgresDatabase(cfg.DatabaseDSN)
		if err != nil {
			return &app{}, err
		}

		return &app{
			db:   db,
			cfg:  cfg,
			auth: auth.New(cfg.SecretKey),
			pool: workerpool.NewPool(2),
		}, nil
	}

	if cfg.FileStoragePath != "" {
		return &app{
			db:   storage.NewFileStorage(cfg.FileStoragePath),
			cfg:  cfg,
			auth: auth.New(cfg.SecretKey),
			pool: workerpool.NewPool(2),
		}, nil
	}

	return &app{
		db:   storage.New(),
		cfg:  cfg,
		auth: auth.New(cfg.SecretKey),
		pool: workerpool.NewPool(2),
	}, nil
}
