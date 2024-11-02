package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupInMemoryStorage() *InMemoryStorage {
	return NewInMemoryStorage()
}

func TestInMemoryStorageCreateAndGetURL(t *testing.T) {
	memoryStorage := setupInMemoryStorage()

	testUUID := uuid.New().String()
	testShortURL := "testShortURL"
	testOriginalURL := "https://example.com"

	err := memoryStorage.CreateShortURL(testUUID, testShortURL, testOriginalURL)
	assert.NoError(t, err, "CreateShortURL should not return an error")

	originalURL, err := memoryStorage.GetOriginalURL(testShortURL)
	assert.NoError(t, err, "GetOriginalURL should not return an error")
	assert.Equal(t, testOriginalURL, originalURL, "Original URL should match the expected value")
}

func TestInMemoryStorageGetOriginalURLNotFound(t *testing.T) {
	memoryStorage := setupInMemoryStorage()

	_, err := memoryStorage.GetOriginalURL("nonExistentShortURL")
	assert.Error(t, err, "GetOriginalURL should return an error for a non-existent URL")
	assert.Contains(t, err.Error(), "URL not found", "Error should indicate URL not found")
}

func TestInMemoryStoragePing(t *testing.T) {
	memoryStorage := setupInMemoryStorage()

	err := memoryStorage.Ping()
	assert.NoError(t, err, "Ping should not return an error if the storage is functioning")
}

func TestInMemoryStorageClose(t *testing.T) {
	memoryStorage := setupInMemoryStorage()

	err := memoryStorage.Close()
	assert.NoError(t, err, "Close should not return an error")
}
