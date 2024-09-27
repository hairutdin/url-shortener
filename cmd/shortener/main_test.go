package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	res := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(res)

	body := `{"url": "https://example.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handlePost(c)

	if res.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), baseURL) {
		t.Errorf("Expected base URL in response, got %s", res.Body.String())
	}
}

func TestHandlePostInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	res := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(res)

	body := `{"url": ""}`
	c.Request = httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handlePost(c)

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

	res := httptest.NewRecorder()

	c, r := gin.CreateTestContext(res)

	r.GET("/:id", handleGet)

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	c.Request = req

	r.ServeHTTP(res, req)

	t.Logf("Requested URL ID: %s", shortID)
	t.Logf("Response Code: %d", res.Code)
	t.Logf("Response Body: %s", res.Body.String())

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

	res := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(res)

	c.Request = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	handleGet(c)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	expectedError := `{"error":"URL not found"}`
	if strings.TrimSpace(res.Body.String()) != expectedError {
		t.Errorf("Expected 'Expected '%s', got %s", expectedError, res.Body.String())
	}
}
