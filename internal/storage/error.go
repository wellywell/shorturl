package storage

import (
	"fmt"
)

const (
	UniqueViolationErr = "23505"
)

type KeyExistsError struct {
	Key string
}

func (e *KeyExistsError) Error() string {
	return fmt.Sprintf("Key %s already exists with different value", e.Key)
}

type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("Key %s not found", e.Key)
}
