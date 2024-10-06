package shortener

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
)

var mockConfig = &config.Config{
	ServerAddress: "localhost:8080",
	BaseURL:       "http://localhost:8080/",
}

func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	return c, res
}

func TestHandleShortenPost(t *testing.T) {
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

func TestHandleGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	shortID := "short12345"
	originalURL := "https://example.com"
	urlStore.Lock()
	urlStore.m[shortID] = originalURL
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
