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

type ValueExistsError struct {
	Value       string
	ExistingKey string
}

func (e *ValueExistsError) Error() string {
	return fmt.Sprintf("Value %s already exists", e.Value)
}

type RecordIsDeleted struct {
	Key string
}

func (e *RecordIsDeleted) Error() string {
	return fmt.Sprintf("Record is deleted %s", e.Key)
}
