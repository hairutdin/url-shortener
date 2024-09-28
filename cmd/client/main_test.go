package main

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

func createTestRequest(method, url, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	return req, recorder
}

func testShortenURL(c *gin.Context) { // Renamed to avoid conflict
	var requestBody struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil || requestBody.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	shortenedURL := mockConfig.BaseURL + "short123" // Use mock config for testing

	c.JSON(http.StatusCreated, gin.H{
		"long_url":  requestBody.URL,
		"short_url": shortenedURL,
	})
}

func TestClientPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/", testShortenURL)

	body := `{"url": "https://example.com"}`
	req, recorder := createTestRequest(http.MethodPost, "/", body)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	expectedBody := `{"long_url":"https://example.com","short_url":"` + mockConfig.BaseURL + `short123"}`
	if recorder.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, recorder.Body.String())
	}
}
