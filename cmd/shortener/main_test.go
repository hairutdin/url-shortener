package shortener

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
)

var mockConfig = &config.Config{
	ServerAddress:   "localhost:8080",
	BaseURL:         "http://localhost:8080/",
	FileStoragePath: "/tmp/test-short-url-db.json",
}

func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	return c, res
}

func TestHandleShortenPost(t *testing.T) {
	defer os.Remove(mockConfig.FileStoragePath)

	gin.SetMode(gin.TestMode)
	c, res := createTestContext()

	body := `{"url": "https://example.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c)

	if res.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), mockConfig.BaseURL) {
		t.Errorf("Expected base URL in response, got %s", res.Body.String())
	}
}

func TestHandleShortenPostInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, res := createTestContext()
	body := `{"url": ""}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	expectedError := `{"error":"Invalid URL"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected 'Expected '%s', got %s", expectedError, res.Body.String())
	}
}

func createGzipRequestBody(body string) *bytes.Buffer {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(body))
	gz.Close()
	return &buf
}

func TestHandleShortenPostWithGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, res := createTestContext()

	gzipBody := createGzipRequestBody(`{"url": "https://example.com"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", gzipBody)
	c.Request.Header.Set("Content-Encoding", "gzip")
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c)

	if res.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), mockConfig.BaseURL) {
		t.Errorf("Expected base URL in response, got %s", res.Body.String())
	}
}

func TestHandleShortenPostWithInvalidGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, res := createTestContext()

	invalidGzipBody := strings.NewReader("invalid gzip data")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", invalidGzipBody)
	c.Request.Header.Set("Content-Encoding", "gzip")
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid gzip, got %d", res.Code)
	}

	expectedError := `{"error":"Failed to read gzip body"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected '%s', got %s", expectedError, res.Body.String())
	}
}

func TestHandleGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	shortID := "short12345"
	originalURL := "https://example.com"
	urlStore.Lock()
	urlStore.m[shortID] = ShortenedURL{
		UUID:        "1",
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}
	urlStore.Unlock()

	c, res := createTestContext()
	r := gin.Default()
	r.GET("/:id", handleGet)

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	c.Request = req

	r.ServeHTTP(res, req)

	if res.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307: got %d", res.Code)
	}

	location := res.Header().Get("Location")
	if location != originalURL {
		t.Errorf("Expected location header to be %s: got %s", originalURL, location)
	}
}

func TestHandleGetInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, res := createTestContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)

	urlStore.RLock()
	defer urlStore.RUnlock()

	handleGet(c)

	if res.Code != http.StatusNotFound {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	expectedError := `{"error":"URL not found"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected 'Expected '%s', got %s", expectedError, res.Body.String())
	}
}

func TestSaveURLsToFile(t *testing.T) {
	defer os.Remove(mockConfig.FileStoragePath)

	urlStore.m["short123"] = ShortenedURL{
		UUID:        "1",
		ShortURL:    "short123",
		OriginalURL: "https://example.com",
	}

	err := saveURLsToFile(mockConfig.FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to save URLs to file: %v", err)
	}

	_, err = os.Stat(mockConfig.FileStoragePath)
	if os.IsNotExist(err) {
		t.Fatalf("File was not created: %v", err)
	}
}

func resetURLStore() {
	urlStore.Lock()
	defer urlStore.Unlock()
	urlStore.m = make(map[string]ShortenedURL)
}

func TestLoadURLsFromFile(t *testing.T) {
	resetURLStore()

	tmpFile, err := os.CreateTemp("", "test_url_store_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	urls := []ShortenedURL{
		{UUID: "1", ShortURL: "short123", OriginalURL: "https://example.com"},
		{UUID: "2", ShortURL: "short456", OriginalURL: "https://another.com"},
	}
	fileData, _ := json.Marshal(urls)
	tmpFile.Write(fileData)
	tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
	}

	err = loadURLsFromFile(cfg.FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to load URLs from file: %v", err)
	}

	urlStore.RLock()
	defer urlStore.RUnlock()
	if len(urlStore.m) != 2 {
		t.Errorf("Expected 2 URLs to be loaded, but got %d", len(urlStore.m))
	}

	if urlStore.m["short123"].OriginalURL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", urlStore.m["short123"].OriginalURL)
	}
}

func TestMain(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_url_store_main_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	urlData := []ShortenedURL{
		{UUID: "1", ShortURL: "short123", OriginalURL: "https://example.com"},
	}
	fileData, _ := json.Marshal(urlData)
	tmpFile.Write(fileData)
	tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
	}

	err = loadURLsFromFile(cfg.FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to load URLs from file: %v", err)
	}

	urlStore.RLock()
	defer urlStore.RUnlock()
	if _, exists := urlStore.m["short123"]; !exists {
		t.Errorf("Expected URL 'short123' to be loaded")
	}
}
