package storage

import (
	"context"
	"sync"
)

type Memory struct {
	urls map[string]string
	lock sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		urls: make(map[string]string),
	}
}

func (m *Memory) Get(ctx context.Context, key string) (string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, ok := m.urls[key]
	if !ok {
		return "", &KeyNotFoundError{Key: key}
	}
	return v, nil
}

func (m *Memory) Put(ctx context.Context, key string, val string) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, exists := m.urls[key]
	if exists && v != val {
		return &KeyExistsError{Key: key}
	}
	m.urls[key] = val
	return nil
}

func (f *Memory) Close() error {
	return nil
}
