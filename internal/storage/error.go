package storage

import "fmt"

const (
	CONFLICT = "conflict"
)

type Error struct {
	Code    string
	Message string
}

func (err *Error) Error() string {
	return fmt.Sprintf("Err: %s (%s)", err.Message, err.Code)
}

func NewError(code, message string) error {
	return &Error{Code: code, Message: message}
}
