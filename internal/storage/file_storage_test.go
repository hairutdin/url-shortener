package storage

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
)

const testFilePath = "/tmp/test_file_storage.json"

func setupFileStorage() (*FileStorage, func()) {
	_ = os.Remove(testFilePath)

	fs, err := NewFileStorage(testFilePath)
	if err != nil {
		panic("Failed to create FileStorage: " + err.Error())
	}
	cleanup := func() {
		fs.Close()
		os.Remove(testFilePath)
	}
	return fs, cleanup
}

func TestCreateShortURL(t *testing.T) {
	fs, cleanup := setupFileStorage()
	defer cleanup()

	testUUID := uuid.New().String()
	testShortURL := "short123"
	testOriginalURL := "https://example.com"

	log.Println("Creating short URL")
	err := fs.CreateShortURL(testUUID, testShortURL, testOriginalURL)
	assert.NoError(t, err, "CreateShortURL should not return an error")

	log.Println("Retrieving original URL")
	originalURL, err := fs.GetOriginalURL(testShortURL)
	assert.NoError(t, err, "GetOriginalURL should not return an error")
	assert.Equal(t, testOriginalURL, originalURL, "Original URL should match expected value")
}

func TestFileStorageBatchURLs(t *testing.T) {
	fs, cleanup := setupFileStorage()
	defer cleanup()

	batchRequests := []BatchURLRequest{
		{
			UUID:        uuid.New().String(),
			ShortURL:    "short1",
			OriginalURL: "https://example.com/1",
		},
		{
			UUID:        uuid.New().String(),
			ShortURL:    "short2",
			OriginalURL: "https://example.com/2",
		},
	}

	for _, req := range batchRequests {
		err := fs.CreateShortURL(req.UUID, req.ShortURL, req.OriginalURL)
		assert.NoError(t, err, "CreateShortURL should not return an error")
	}

	for _, req := range batchRequests {
		originalURL, err := fs.GetOriginalURL(req.ShortURL)
		assert.NoError(t, err, "GetOriginalURL should not return an error")
		assert.Equal(t, req.OriginalURL, originalURL, "The original URL should match the expected value")
	}
}

func TestGetOriginalURLNotFound(t *testing.T) {
	fs, cleanup := setupFileStorage()
	defer cleanup()

	log.Println("Testing GetOriginalURL with nonexistent URL")
	_, err := fs.GetOriginalURL("nonexistent")
	assert.Error(t, err, "GetOriginalURL should return an error for non-existent short URL")
}

func TestSaveAndLoadFromFile(t *testing.T) {
	fs, cleanup := setupFileStorage()
	defer cleanup()

	testUUID := uuid.New().String()
	testShortURL := "short123"
	testOriginalURL := "https://example.com"

	log.Println("Creating short URL and saving to file")
	err := fs.CreateShortURL(testUUID, testShortURL, testOriginalURL)
	assert.NoError(t, err, "CreateShortURL should not return an error")

	log.Println("Reloading FileStorage to test persistence")
	fs2, err := NewFileStorage(testFilePath)
	assert.NoError(t, err, "NewFileStorage should not return an error when loading from file")

	log.Println("Retrieving original URL from reloaded storage")
	originalURL, err := fs2.GetOriginalURL(testShortURL)
	assert.NoError(t, err, "GetOriginalURL should not return an error after reloading from file")
	assert.Equal(t, testOriginalURL, originalURL, "Original URL should match expected value after reloading")
}

func TestPing(t *testing.T) {
	fs, cleanup := setupFileStorage()
	defer cleanup()

	log.Println("Testing Ping method")
	err := fs.Ping()
	assert.NoError(t, err, "Ping should not return an error if the storage is reachable")
}

func TestClose(t *testing.T) {
	fs, cleanup := setupFileStorage()
	defer cleanup()

	log.Println("Testing Close method")
	err := fs.Close()
	assert.NoError(t, err, "Close should not return an error")
}
