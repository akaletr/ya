package gziper

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что запрос поддерживает gzip-сжатие
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			reader := bytes.NewReader(body)
			gzReader, err := gzip.NewReader(reader)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			o, err := io.ReadAll(gzReader)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			output := strings.NewReader(string(o))
			oc := io.NopCloser(output)
			r.Body = oc
		}

		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			_, err = io.WriteString(w, err.Error())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
		defer func() {
			_ = gz.Close()
		}()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
