package storage

type Storage interface {
	Read(value string) (string, error)
	Write(id, key, value string) error

	ReadAll(id string) (map[string]string, error)
}
