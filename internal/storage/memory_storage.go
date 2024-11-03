package storage

import (
	"errors"
	"sync"
)

type InMemoryStorage struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{urls: make(map[string]string)}
}

func (m *InMemoryStorage) CreateShortURL(uuid, shortURL, originalURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortURL] = originalURL
	return nil
}

func (m *InMemoryStorage) CreateBatchURLs(urls []BatchURLRequest) ([]BatchURLOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var output []BatchURLOutput

	for _, url := range urls {
		if _, exists := m.urls[url.ShortURL]; exists {
			return nil, errors.New("duplicate short URL")
		}
		m.urls[url.ShortURL] = url.OriginalURL
		output = append(output, BatchURLOutput{
			CorrelationID: url.UUID,
			ShortURL:      "http://localhost:8080/" + url.ShortURL,
		})
	}

	return output, nil
}

func (m *InMemoryStorage) GetOriginalURL(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, exists := m.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (m *InMemoryStorage) Ping() error {
	return nil
}

func (m *InMemoryStorage) Close() error {
	return nil
}
