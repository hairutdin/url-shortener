package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePost(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	res := httptest.NewRecorder()

	handlePost(res, req)

	if res.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), baseURL) {
		t.Errorf("Expected base URL in response, got %s", res.Body.String())
	}
}

func TestHandlePostInvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "text/plain")

	res := httptest.NewRecorder()

	handlePost(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	if res.Body.String() != "Invalid request body\n" {
		t.Errorf("Expected 'Invalid request body', got %s", res.Body.String())
	}
}

func TestHandleGet(t *testing.T) {
	shortID := "short12345"
	originalURL := "https://example.com"
	urlStore.Lock()
	urlStore.m[shortID] = originalURL
	urlStore.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)

	res := httptest.NewRecorder()

	handleGet(res, req)

	if res.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307: got %d", res.Code)
	}

	location := res.Header().Get("Location")
	if location != originalURL {
		t.Errorf("Expected location header to be %s: got %s", originalURL, location)
	}
}

func TestHandleGetInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)

	res := httptest.NewRecorder()

	handleGet(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}
}
