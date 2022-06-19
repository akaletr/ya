package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cmd/shortener/main.go/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	name := "create app test"
	want := &app{db: storage.New()}

	t.Run(name, func(t *testing.T) {
		assert.Equal(t, want, New())
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
			want:   []byte("http://-esgrQ"),
			status: http.StatusCreated,
		},
		{
			name:   "test 2",
			args:   "google.com",
			want:   []byte("http://4U8Jkw"),
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
	}
	handler := http.HandlerFunc(app.AddURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := strings.NewReader(tt.args)
			req, _ := http.NewRequest(http.MethodPost, "/", b)

			handler.ServeHTTP(rec, req)
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
			args:   "hello",
			want:   []byte("world"),
			status: http.StatusTemporaryRedirect,
		},
		{
			name:   "test 2",
			args:   "exist",
			want:   []byte("yes"),
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
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tt.args), nil)

			r := chi.NewRouter()
			r.Get("/{id}", app.GetURL)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.want, rec.Body.Bytes())
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}
