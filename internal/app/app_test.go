package app

import (
	"cmd/shortener/main.go/internal/auth"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cmd/shortener/main.go/internal/config"
	"cmd/shortener/main.go/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := config.Config{
		ServerAddress: "localhost:8080",
		Key:           "yandex",
	}

	name := "create app test"
	want := &app{
		db:   storage.New(),
		cfg:  cfg,
		auth: auth.New(cfg.Key),
	}

	t.Run(name, func(t *testing.T) {
		assert.Equal(t, want, New(cfg))
	})
}

func Test_app_AddURL(t *testing.T) {
	tests := []struct {
		name   string
		args   string
		want   []byte
		status int
	}{
		{
			name:   "test 1",
			args:   "helloworld",
			want:   nil,
			status: http.StatusBadRequest,
		},
		{
			name:   "test 2",
			args:   "http://google.com",
			want:   []byte("http://localhost:8080/cSDPTQ"),
			status: http.StatusCreated,
		},
		{
			name:   "test 3",
			args:   "",
			want:   nil,
			status: http.StatusBadRequest,
		},
	}

	app := &app{
		db: storage.NewMock(),
		cfg: config.Config{
			BaseURL: "http://localhost:8080",
			Key:     "yandex",
		},
		auth: auth.New("yandex"),
	}
	handler := http.HandlerFunc(app.AddURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := strings.NewReader(tt.args)
			req, err := http.NewRequest(http.MethodPost, "/", b)

			handler.ServeHTTP(rec, req)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rec.Body.Bytes())
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

func Test_app_GetURL(t *testing.T) {
	tests := []struct {
		name   string
		args   string
		want   []byte
		status int
	}{
		{
			name:   "test 1",
			args:   "kUxCqw",
			want:   []byte("https://www.delftstack.com/ru/howto/go/how-to-read-a-file-line-by-line-in-go/"),
			status: http.StatusTemporaryRedirect,
		},
		{
			name:   "test 2",
			args:   "D-rwfg",
			want:   []byte("https://www.jetbrains.com/ru-ru/"),
			status: http.StatusTemporaryRedirect,
		},
		{
			name:   "test 3",
			args:   "not_exist",
			want:   nil,
			status: http.StatusNotFound,
		},
	}

	app := &app{
		db: storage.NewMock(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tt.args), nil)

			r := chi.NewRouter()
			r.Get("/{id}", app.GetURL)
			r.ServeHTTP(rec, req)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rec.Body.Bytes())
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

func Test_app_Shorten(t *testing.T) {
	tests := []struct {
		name   string
		args   string
		want   []byte
		status int
	}{
		{
			name:   "test 1",
			args:   `{"url":"helloworld"}`,
			want:   nil,
			status: http.StatusBadRequest,
		},
		{
			name:   "test 2",
			args:   `{"url":"helloworld}`,
			want:   nil,
			status: http.StatusBadRequest,
		},
		{
			name:   "test 3",
			args:   `{"url":"https://habr.com/ru/post/479882/"}`,
			want:   []byte(`{"result":"http://localhost:8080/XlVFpw"}`),
			status: http.StatusCreated,
		},
	}

	app := &app{
		db: storage.NewMock(),
		cfg: config.Config{
			BaseURL: "http://localhost:8080",
		},
	}
	handler := http.HandlerFunc(app.Shorten)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := strings.NewReader(tt.args)
			req, err := http.NewRequest(http.MethodPost, "/shorten", b)
			handler.ServeHTTP(rec, req)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rec.Body.Bytes())
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

func Test_app_validateURL(t *testing.T) {
	type args struct {
		URL []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "test 1",
			args:    args{URL: []byte("google.com")},
			wantErr: assert.Error,
		},
		{
			name:    "test 2",
			args:    args{URL: []byte("https://golang-blog.blogspot.com/2021")},
			wantErr: assert.NoError,
		},
		{
			name:    "test 3",
			args:    args{URL: []byte("")},
			wantErr: assert.Error,
		},
		{
			name:    "test 4",
			args:    args{URL: []byte("https//golang-blog.blogspot.com/2021")},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app{}
			tt.wantErr(t, app.validateURL(tt.args.URL), fmt.Sprintf("validateURL(%v)", tt.args.URL))
		})
	}
}
