package storage

import (
	"fmt"
)

// Ошибка записи дублирующего значения в БД Postgres
const (
	UniqueViolationErr = "23505"
)

// KeyExistsError ошибка повторной записи ключа
type KeyExistsError struct {
	Key string
}

func (e *KeyExistsError) Error() string {
	return fmt.Sprintf("Key %s already exists with different value", e.Key)
}

// KeyNotFoundError ошибка при попытке достать несуществующих ключ
type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("Key %s not found", e.Key)
}

// ValueExistsError ошибка при записи значения-дубликата
type ValueExistsError struct {
	Value       string
	ExistingKey string
}

func (e *ValueExistsError) Error() string {
	return fmt.Sprintf("Value %s already exists", e.Value)
}

// RecordIsDeleted ошибка при попытке достать уже удаленную запись
type RecordIsDeleted struct {
	Key string
}

func (e *RecordIsDeleted) Error() string {
	return fmt.Sprintf("Record is deleted %s", e.Key)
}
