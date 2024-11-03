package shortener

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hairutdin/url-shortener/config"
	"github.com/hairutdin/url-shortener/internal/storage"
	"go.uber.org/zap"
)

var mockConfig = &config.Config{
	ServerAddress:   "localhost:8080",
	BaseURL:         "http://localhost:8080/",
	FileStoragePath: "/tmp/test-short-url-db.json",
	DatabaseDSN:     "postgres://postgres:berlin@localhost:5432/testdb?sslmode=disable",
}

func setupTestStorage() (func(), error) {
	var err error
	storageInstance, err = initializeStorage(mockConfig)
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		if fileStorage, ok := storageInstance.(*storage.FileStorage); ok {
			fileStorage.Close()
			os.Remove(mockConfig.FileStoragePath)
		}
	}

	return cleanup, nil
}

func init() {
	gin.SetMode(gin.ReleaseMode) // Set Gin to release mode
}

func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	return c, res
}

func TestHandleShortenPost(t *testing.T) {
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	c, res := createTestContext()

	body := `{"url": "https://example.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c, mockConfig)

	if res.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), mockConfig.BaseURL) {
		t.Errorf("Expected base URL in response, got %s", res.Body.String())
	}
}

func TestHandleShortenPostInvalidBody(t *testing.T) {
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	c, res := createTestContext()
	body := `{"url": ""}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c, mockConfig)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	expectedError := `{"error":"Invalid URL"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected '%s', got %s", expectedError, res.Body.String())
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
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	c, res := createTestContext()

	gzipBody := createGzipRequestBody(`{"url": "https://example.com"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", gzipBody)
	c.Request.Header.Set("Content-Encoding", "gzip")
	c.Request.Header.Set("Accept-Encoding", "gzip")

	handleShortenPost(c, mockConfig)

	if res.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", res.Code)
	}

	if res.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding header to be gzip, got %s", res.Header().Get("Content-Encoding"))
	}

	gzipReader, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	decodedBody := new(strings.Builder)
	if _, err := io.Copy(decodedBody, gzipReader); err != nil {
		t.Fatalf("Failed to read gzip response: %v", err)
	}

	expectedContent := `{"result":"` + mockConfig.BaseURL
	if !strings.Contains(decodedBody.String(), expectedContent) {
		t.Errorf("Expected response to contain %s, got %s", expectedContent, decodedBody.String())
	}
}

func TestHandleShortenPostWithInvalidGzip(t *testing.T) {
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	c, res := createTestContext()

	invalidGzipBody := strings.NewReader("invalid gzip data")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", invalidGzipBody)
	c.Request.Header.Set("Content-Encoding", "gzip")
	c.Request.Header.Set("Content-Type", "application/json")

	handleShortenPost(c, mockConfig)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid gzip, got %d", res.Code)
	}

	expectedError := `{"error":"Failed to read gzip body"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected '%s', got %s", expectedError, res.Body.String())
	}
}

func TestHandleBatchShortenPost(t *testing.T) {
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/api/shorten/batch", func(c *gin.Context) { handleBatchShortenPost(c, mockConfig) })

	batchRequest := `[{
		"correlation_id": "id1",
		"original_url": "https://example.com/1"
    }, {
        "correlation_id": "id2",
        "original_url": "https://example.com/2"
    }]`

	req, _ := http.NewRequest("POST", "/api/shorten/batch", strings.NewReader(batchRequest))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusCreated {
		t.Errorf("Expected status 201, got %v", status)
	}

	var response []BatchShortenResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 shortened URLs in response, got %d", len(response))
	}
}

func TestHandleGet(t *testing.T) {
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)

	shortURL := uuid.New().String()
	originalURL := "https://example.com"
	validUUID := uuid.New().String()

	e := storageInstance.CreateShortURL(validUUID, shortURL, originalURL)
	if e != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}

	c, res := createTestContext()
	r := gin.Default()
	r.GET("/:id", handleGet)

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
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
	cleanup, err := setupTestStorage()
	if err != nil {
		panic("Failed to set up test storage: " + err.Error())
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)

	c, res := createTestContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)

	handleGet(c)

	if res.Code != http.StatusNotFound {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	expectedError := `{"error":"URL not found"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected 'Expected '%s', got %s", expectedError, res.Body.String())
	}
}

func TestPingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Database connected", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
		handlePing(c)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestMain(m *testing.M) {
	cfg := config.LoadConfig()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var err error
	storageInstance, err = initializeStorage(cfg)
	if err != nil {
		panic("Failed to initialize storage for testing: " + err.Error())
	}
	defer storageInstance.Close()

	r := setupRouter(cfg, logger)

	routes := []struct {
		method string
		path   string
		body   io.Reader
		expect int
	}{
		{"POST", "/", strings.NewReader(`{"url": "https://example.com"}`), http.StatusCreated},
		{"POST", "/api/shorten", strings.NewReader(`{"url": "https://example.com"}`), http.StatusCreated},
		{"GET", "/ping", nil, http.StatusOK},
		{"GET", "/nonexistent", nil, http.StatusNotFound},
	}

	hasErrors := false
	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, route.body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != route.expect {
			hasErrors = true
			logger.Error("Route test failed",
				zap.String("method", route.method),
				zap.String("path", route.path),
				zap.Int("expected", route.expect),
				zap.Int("got", w.Code),
			)
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestInitializeStorage(t *testing.T) {
	cfg := &config.Config{DatabaseDSN: mockConfig.DatabaseDSN}
	storage, err := initializeStorage(cfg)
	if err != nil {
		t.Fatalf("Expected no error for PostgreSQL storage, got %v", err)
	}
	defer storage.Close()

	cfg = &config.Config{FileStoragePath: mockConfig.FileStoragePath}
	storage, err = initializeStorage(cfg)
	if err != nil {
		t.Fatalf("Expected no error for file storage, got %v", err)
	}
	defer storage.Close()

	cfg = &config.Config{}
	storage, err = initializeStorage(cfg)
	if err != nil {
		t.Fatalf("Expected no error for in-memory storage, got %v", err)
	}
	defer storage.Close()
}
