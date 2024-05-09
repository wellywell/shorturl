package storage

import (
	"context"
	"fmt"
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
		return "", fmt.Errorf("%w", &KeyNotFoundError{Key: key})
	}
	return v, nil
}

func (m *Memory) Put(ctx context.Context, key string, val string) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, exists := m.urls[key]
	if exists && v != val {
		return fmt.Errorf("%w", &KeyExistsError{Key: key})
	}
	m.urls[key] = val
	return nil
}

func (m *Memory) PutBatch(ctx context.Context, records ...KeyValue) error {
	for _, rec := range records {
		if err := m.Put(ctx, rec.Key, rec.Value); err != nil {
			return err
		}
	}
	return nil
}

func (m *Memory) Close() error {
	return nil
}
