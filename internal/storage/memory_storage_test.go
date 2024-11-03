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

func TestInMemoryStorageBatchCreate(t *testing.T) {
	memoryStorage := setupInMemoryStorage()

	batchURLs := []BatchURLRequest{
		{UUID: uuid.New().String(), ShortURL: "short1", OriginalURL: "https://example.com/1"},
		{UUID: uuid.New().String(), ShortURL: "short2", OriginalURL: "https://example.com/2"},
	}

	// Test batch creation of short URLs
	batchOutputs, err := memoryStorage.CreateBatchURLs(batchURLs)
	assert.NoError(t, err, "CreateBatchURLs should not return an error")
	assert.Len(t, batchOutputs, len(batchURLs), "The output length should match the input length")

	// Verify each URL was created correctly
	for i := range batchOutputs {
		originalURL, err := memoryStorage.GetOriginalURL(batchURLs[i].ShortURL)
		assert.NoError(t, err, "GetOriginalURL should not return an error")
		assert.Equal(t, batchURLs[i].OriginalURL, originalURL, "The retrieved URL should match the original URL")
	}
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
