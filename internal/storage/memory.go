package storage

import (
	"fmt"
	"sync"
)

type KeyExistsError struct {
	Key string
}

func (e *KeyExistsError) Error() string {
	return fmt.Sprintf("Key %s already exists with different value", e.Key)
}

type Memory struct {
	urls map[string]string
	lock sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		urls: make(map[string]string),
	}
}

func (m *Memory) Get(key string) (string, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, ok := m.urls[key]
	return v, ok
}

func (m *Memory) Put(key string, val string) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, exists := m.urls[key]
	if exists && v != val {
		return &KeyExistsError{Key: key}
	}
	m.urls[key] = val
	return nil
}
