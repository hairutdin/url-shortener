package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClientPost(t *testing.T) {
	// Mock server for testing
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/short123"))
	}))
	defer mockServer.Close()

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/shorten", shortenURL)

	// Read the body
	body := strings.NewReader("url=https://example.com")
	req, err := http.NewRequest(http.MethodPost, "/shorten", body)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Check if the response status is correct
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	// Check the response body
	expectedBody := `{"long_url":"https://example.com","short_url":"http://localhost:8080/short123"}`
	if recorder.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, recorder.Body.String())
	}
}
