package repository

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type FileStorage struct {
	filePath string
	mu       sync.RWMutex
	urls     map[string]string // shortURL -> originalURL
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{filePath: filePath, urls: make(map[string]string)}
	if err := fs.loadFromFile(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (f *FileStorage) loadFromFile() error {
	if _, err := os.Stat(f.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &f.urls)
}

func (f *FileStorage) saveToFile() error {
	log.Println("Saving to file without acquiring lock")
	data, err := json.Marshal(f.urls)
	if err != nil {
		return err
	}

	err = os.WriteFile(f.filePath, data, 0644)
	if err != nil {
		log.Println("Error writing to file:", err)
	}
	log.Println("File saved successfully in saveToFile")
	return err
}

func (f *FileStorage) CreateShortURL(_, shortURL, originalURL string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	existingShortURL, _ := f.GetShortURLByOriginal(originalURL)
	if existingShortURL != "" {
		return existingShortURL, errors.New("URL already exists")
	}

	f.urls[shortURL] = originalURL
	if err := f.saveToFile(); err != nil {
		return "", err
	}
	return shortURL, nil
}

func (f *FileStorage) CreateBatchURLs(urls []BatchURLRequest) ([]BatchURLOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var output []BatchURLOutput
	for _, url := range urls {
		if _, exists := f.urls[url.ShortURL]; exists {
			return nil, errors.New("duplicate short URL")
		}
		f.urls[url.ShortURL] = url.OriginalURL
		output = append(output, BatchURLOutput{
			CorrelationID: url.UUID,
			ShortURL:      f.filePath + "/" + url.ShortURL,
		})
	}

	if err := f.saveToFile(); err != nil {
		return nil, err
	}

	return output, nil
}

func (f *FileStorage) GetShortURLByOriginal(originalURL string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for short, original := range f.urls {
		if original == originalURL {
			return short, nil
		}
	}
	return "", nil
}

func (f *FileStorage) GetOriginalURL(shortURL string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	originalURL, exists := f.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (f *FileStorage) Ping() error {
	return nil
}

func (f *FileStorage) Close() error {
	log.Println("Closing FileStorage and saving to file")
	return f.saveToFile()
}
