package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hairutdin/url-shortener/config"

	"github.com/gin-gonic/gin"
)

var mockConfig = &config.Config{
	ServerAddress: "localhost:8080",
	BaseURL:       "http://localhost:8080/",
}

func createTestRequest(method, url, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	return req, recorder
}

func TestClientPost(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(mockConfig.BaseURL + "short123"))
	}))
	defer mockServer.Close()

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/shorten", shortenURL)

	body := "url=https://example.com"
	req, recorder := createTestRequest(http.MethodPost, "/shorten", body)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	expectedBody := `{"long_url":"https://example.com","short_url":"` + mockConfig.BaseURL + `short123"}`
	if recorder.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, recorder.Body.String())
	}
}
