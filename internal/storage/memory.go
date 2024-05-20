package storage

import (
	"context"
	"fmt"
	"sync"
)

type FullUrlData struct {
	UserID  int
	FullURL string
}

type Memory struct {
	urls      map[string]FullUrlData
	maxUserID int
	lock      sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		urls:      make(map[string]FullUrlData),
		maxUserID: 0,
	}
}

func (m *Memory) Get(ctx context.Context, key string) (string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, ok := m.urls[key]
	if !ok {
		return "", fmt.Errorf("%w", &KeyNotFoundError{Key: key})
	}
	return v.FullURL, nil
}

func (m *Memory) Put(ctx context.Context, key string, val string, user int) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, exists := m.urls[key]
	if exists && v.FullURL != val {
		return fmt.Errorf("%w", &KeyExistsError{Key: key})
	}
	m.urls[key] = FullUrlData{FullURL: val, UserID: user}
	if user > m.maxUserID {
		m.maxUserID = user
	}
	return nil
}

func (m *Memory) CreateNewUser(ctx context.Context) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.maxUserID = m.maxUserID + 1
	return m.maxUserID, nil
}

func (m *Memory) PutBatch(ctx context.Context, records ...URLRecord) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, rec := range records {
		if err := m.Put(ctx, rec.ShortURL, rec.FullURL, rec.UserID); err != nil {
			return err
		}
	}
	return nil
}

func (m *Memory) GetUserURLS(ctx context.Context, userID int) ([]URLRecord, error) {
	var urls []URLRecord

	m.lock.RLock()
	defer m.lock.RUnlock()

	// O(n)
	for short, record := range m.urls {
		if record.UserID == userID {
			urls = append(urls, URLRecord{UserID: userID, ShortURL: short, FullURL: record.FullURL})
		}
	}
	return urls, nil
}

func (m *Memory) Close() error {
	return nil
}
