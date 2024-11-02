package storage

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

var testDSN = "postgres://postgres:password@localhost:5432/testdb?sslmode=disable"

func setupTestDB(t *testing.T) *PostgresStorage {
	db, err := NewPostgresStorage(testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	_, err = db.db.Exec(context.Background(), "TRUNCATE TABLE shortened_urls")
	if err != nil {
		t.Fatalf("Failed to clear table: %v", err)
	}

	return db
}

func TestPostgresStorageCreateAndGetURL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testUUID := uuid.New().String()
	testShortURL := "testShortURL"
	testOriginalURL := "https://example.com"

	err := db.CreateShortURL(testUUID, testShortURL, testOriginalURL)
	assert.NoError(t, err, "CreateShortURL should not return an error")

	originalURL, err := db.GetOriginalURL(testShortURL)
	assert.NoError(t, err, "GetOriginalURL should not return an error")
	assert.Equal(t, testOriginalURL, originalURL, "Original URL should match the expected value")
}

func TestPostgresStorageGetOriginalURLNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetOriginalURL("nonExistentShortURL")
	assert.Error(t, err, "GetOriginalURL should return an error for non-existent URL")
	assert.Contains(t, err.Error(), "no rows", "Error should indicate no rows found")
}

func TestPostgresStoragePing(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.Ping()
	assert.NoError(t, err, "Ping should not return an error if the database is connected")
}

func TestPostgresStorageClose(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.Close()
	assert.NoError(t, err, "Close should not return an error")
}

func TestMain(m *testing.M) {
	code := m.Run()

	db, _ := pgx.Connect(context.Background(), testDSN)
	db.Exec(context.Background(), "DROP TABLE IF EXISTS shortened_urls")
	db.Close(context.Background())

	os.Exit(code)
}
