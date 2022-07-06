package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
)

type auth struct {
	Key string
}

// New возвращает экземпляр авторизации
func New(key string) Auth {
	return &auth{
		Key: key,
	}
}

// CookieHandler читает куки и устанавливает, если их нет
func (auth *auth) CookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user")

		if err != nil || !auth.Check(cookie) {
			value, e := auth.NewToken()
			if e != nil {
				log.Println(err)
				//w.WriteHeader(http.StatusInternalServerError)
				return
			}

			cookie = &http.Cookie{
				Name:  "user",
				Value: hex.EncodeToString(value),
				Path:  "/",
			}
			http.SetCookie(w, cookie)
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Check проверяет подпись
func (auth *auth) Check(cookie *http.Cookie) bool {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		log.Println(err)
	}

	id := binary.BigEndian.Uint32(data[:4])
	log.Println("id", id)

	h := hmac.New(sha256.New, []byte(auth.Key))
	h.Write(data[:4])
	sign := h.Sum(nil)

	return hmac.Equal(sign, data[4:])
}

// NewToken создает новый токен
func (auth *auth) NewToken() ([]byte, error) {
	src, err := generateRandom(4)
	if err != nil {
		return nil, err
	}

	// подписываем алгоритмом HMAC, используя sha256
	h := hmac.New(sha256.New, []byte(auth.Key))
	h.Write(src)
	dst := h.Sum(nil)

	dst = append(src, dst...)
	return dst, nil
}

func (auth *auth) GetID(cookie *http.Cookie) (string, error) {
	if cookie.Value == "" {
		return "default", nil
	}
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return "default", err
	}

	return strconv.FormatUint(uint64(binary.BigEndian.Uint32(data[:4])), 10), nil
}

func generateRandom(size int) ([]byte, error) {
	// генерируем случайную последовательность байт
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
