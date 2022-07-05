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
			e, value := auth.NewToken()
			if e != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			cookie = &http.Cookie{
				Name:  "user",
				Value: hex.EncodeToString(value),
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

	if hmac.Equal(sign, data[4:]) {
		return true
	}

	return false
}

// NewToken создает новый токен
func (auth *auth) NewToken() (error, []byte) {
	src, err := generateRandom(4)
	if err != nil {
		return err, nil
	}

	// подписываем алгоритмом HMAC, используя sha256
	h := hmac.New(sha256.New, []byte(auth.Key))
	h.Write(src)
	dst := h.Sum(nil)

	dst = append(src, dst...)
	return nil, dst
}

func (auth *auth) GetID(cookie *http.Cookie) (error, string) {
	if cookie.Value == "" {
		return nil, ""
	}
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return err, ""
	}

	return nil, strconv.FormatUint(uint64(binary.BigEndian.Uint32(data[:4])), 10)
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
