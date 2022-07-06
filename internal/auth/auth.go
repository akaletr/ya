package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
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

// Check проверяет подпись
func (auth *auth) Check(cookie *http.Cookie) bool {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		log.Println(err)
		return false
	}

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
		return "", errors.New("")
	}
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return "", err
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
