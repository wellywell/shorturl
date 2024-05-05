package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

type MemoryStorage interface {
	Put(ctx context.Context, key string, val string) error
	Get(ctx context.Context, key string) (string, error)
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileMemory struct {
	file     *os.File
	writer   *bufio.Writer
	memory   MemoryStorage
	lastUUID int
	lock     sync.RWMutex
}

func NewFileMemory(path string, memory MemoryStorage) (*FileMemory, error) {
	storage := FileMemory{
		memory: memory,
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	if err := storage.loadFromFile(f); err != nil {
		return nil, err
	}
	storage.file = f
	storage.writer = bufio.NewWriter(f)

	return &storage, nil
}

func (f *FileMemory) Put(ctx context.Context, key string, val string) error {

	if err := f.memory.Put(ctx, key, val); err != nil {
		return err
	}
	if err := f.writeToFile(key, val); err != nil {
		return err
	}
	return nil
}

func (f *FileMemory) Get(ctx context.Context, key string) (string, error) {
	return f.memory.Get(ctx, key)
}

func (f *FileMemory) writeToFile(key string, val string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	nextUUID := f.lastUUID + 1

	record := URLRecord{
		UUID:        strconv.Itoa(nextUUID),
		ShortURL:    key,
		OriginalURL: val,
	}
	data, err := json.Marshal(record)

	if err != nil {
		return err
	}

	if _, err := f.writer.Write(data); err != nil {
		return err
	}

	if err := f.writer.WriteByte('\n'); err != nil {
		return err
	}
	f.lastUUID = nextUUID

	return f.writer.Flush()
}

func (f *FileMemory) loadFromFile(file *os.File) error {
	scanner := bufio.NewScanner(file)

	ctx := context.Background()

	for {
		if !scanner.Scan() {
			return scanner.Err()
		}
		data := scanner.Bytes()

		record := URLRecord{}
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}
		if err := f.memory.Put(ctx, record.ShortURL, record.OriginalURL); err != nil {
			return err
		}

		var err error
		f.lastUUID, err = strconv.Atoi(record.UUID)
		if err != nil {
			return err
		}
	}
}

func (f *FileMemory) Close() error {
	return f.file.Close()
}
