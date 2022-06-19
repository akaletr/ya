package storage

type Storage interface {
	Read(value string) (string, error)
	Write(key, value string) error
}
