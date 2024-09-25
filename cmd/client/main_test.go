package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientPost(t *testing.T) {
	// Mock server for testing
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the method is POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check the content type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected content-type application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		// Read the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Error reading body: %v", err)
		}

		// Check if the body contains the long URL
		expectedURL := "url=https%3A%2F%2Fexample.com"
		if string(body) != expectedURL {
			t.Errorf("Expected body %s, got %s", expectedURL, string(body))
		}

		// Respond with a mock short URL
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/short123"))
	}))
	defer mockServer.Close()

	// Simulate input from the console
	longURL := "https://example.com\n"
	reader := strings.NewReader(longURL)

	err := shortenURL(reader, mockServer.URL)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
